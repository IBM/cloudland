/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"
	"web/src/utils/encrpt"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	instanceAdmin = &InstanceAdmin{}
	instanceView  = &InstanceView{}
)

const MaxmumSnapshot = 96

type InstanceAdmin struct{}

type InstanceView struct{}

type NetworkRoute struct {
	Network string `json:"network"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

type ExecutionCommand struct {
	Control string
	Command string
}

type InstanceNetwork struct {
	Type    string          `json:"type,omitempty"`
	Address string          `json:"ip_address"`
	Netmask string          `json:"netmask"`
	Link    string          `json:"link"`
	ID      string          `json:"id"`
	Routes  []*NetworkRoute `json:"routes,omitempty"`
}

type NetworkLink struct {
	MacAddr string `json:"ethernet_mac_address"`
	Mtu     uint   `json:"mtu"`
	ID      string `json:"id"`
	Type    string `json:"type,omitempty"`
}

type VolumeInfo struct {
	ID      int64 `json:"id"`
	UUID    string `json:"uuid"`
	Device  string `json:"device"`
	Booting bool   `json:"booting"`
}

type InstanceData struct {
	Userdata   string             `json:"userdata"`
	VirtType   string             `json:"virt_type"`
	DNS        string             `json:"dns"`
	Vlans      []*VlanInfo        `json:"vlans"`
	Networks   []*InstanceNetwork `json:"networks"`
	Links      []*NetworkLink     `json:"links"`
	Volumes    []*VolumeInfo      `json:"volumes"`
	Keys       []string           `json:"keys"`
	RootPasswd string             `json:"root_passwd"`
	OSCode     string             `json:"os_code"`
}

type InstancesData struct {
	Instances []*model.Instance `json:"instancedata"`
	IsAdmin   bool              `json:"is_admin"`
}

func (a *InstanceAdmin) getHyperGroup(ctx context.Context, imageType string, zoneID int64) (hyperGroup string, err error) {
	ctx, db := GetContextDB(ctx)
	hypers := []*model.Hyper{}
	where := fmt.Sprintf("zone_id = %d and status = 1", zoneID)
	if imageType != "" {
		where = fmt.Sprintf("%s and virt_type = '%s'", where, imageType)
	}
	if err = db.Where(where).Find(&hypers).Error; err != nil {
		logger.Error("Hypers query failed", err)
		return
	}
	if len(hypers) == 0 {
		logger.Error("No qualified hypervisor")
		return
	}
	hyperGroup = fmt.Sprintf("group-zone-%d", zoneID)
	for i, h := range hypers {
		if i == 0 {
			hyperGroup = fmt.Sprintf("%s:%d", hyperGroup, h.Hostid)
		} else {
			hyperGroup = fmt.Sprintf("%s,%d", hyperGroup, h.Hostid)
		}
	}
	return
}

func (a *InstanceAdmin) Create(ctx context.Context, count int, prefix, userdata string, image *model.Image,
	flavor *model.Flavor, zone *model.Zone, routerID int64, primaryIface *InterfaceInfo, secondaryIfaces []*InterfaceInfo,
	keys []*model.Key, rootPasswd string, hyperID int) (instances []*model.Instance, err error) {
	logger.Debugf("Create %d instances with image %s, flavor %s, zone %s, router %d, primary interface %v, secondary interfaces %v, keys %v, root password %s, hyper %d",
		count, image.Name, flavor.Name, zone.Name, routerID, primaryIface, secondaryIfaces, keys, "********", hyperID)
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	if image.Status != "available" {
		err = fmt.Errorf("Image status not available")
		logger.Error("Image status not available")
		return
	}
	zoneID := zone.ID
	if hyperID >= 0 {
		hyper := &model.Hyper{}
		err = db.Where("hostid = ?", hyperID).Take(hyper).Error
		if err != nil {
			logger.Error("Failed to query hypervisor", err)
			return
		}
		if hyper.ZoneID != zone.ID {
			logger.Errorf("Hypervisor %v is not in zone %d, %v", hyper, zoneID, err)
			err = fmt.Errorf("Hypervisor is not in this zone")
			return
		}
	}
	hyperGroup, err := a.getHyperGroup(ctx, image.VirtType, zoneID)
	if err != nil {
		logger.Error("No valid hypervisor", err)
		return
	}
	passwdLogin := false
	if rootPasswd != "" {
		passwdLogin = true
		logger.Debug("Root password login enabled")
	}
	execCommands := []*ExecutionCommand{}
	i := 0
	hostname := prefix
	for i < count {
		if count > 1 {
			hostname = fmt.Sprintf("%s-%d", prefix, i+1)
		}
		total := 0
		if err = db.Unscoped().Model(&model.Instance{}).Where("image_id = ?", image.ID).Count(&total).Error; err != nil {
			logger.Error("Failed to query total instances with the image", err)
			return
		}
		snapshot := total/MaxmumSnapshot + 1 // Same snapshot reference can not be over 128, so use 96 here
		instance := &model.Instance{
			Model:       model.Model{Creater: memberShip.UserID},
			Owner:       memberShip.OrgID,
			Hostname:    hostname,
			ImageID:     image.ID,
			Snapshot:    int64(snapshot),
			FlavorID:    flavor.ID,
			Keys:        keys,
			PasswdLogin: passwdLogin,
			Userdata:    userdata,
			Status:      "pending",
			ZoneID:      zoneID,
			RouterID:    routerID,
		}
		err = db.Create(instance).Error
		if err != nil {
			logger.Error("DB create instance failed", err)
			return
		}
		instance.Image = image
		instance.Flavor = flavor
		instance.Zone = zone
		var bootVolume *model.Volume
		imagePrefix := fmt.Sprintf("image-%d-%s", image.ID, strings.Split(image.UUID, "-")[0])
		// boot volume name format: instance-15-boot-volume-10
		bootVolume, err = volumeAdmin.CreateVolume(ctx, fmt.Sprintf("instance-%d-boot-volume", instance.ID), flavor.Disk, instance.ID, true, 0, 0, 0, 0, "")
		if err != nil {
			logger.Error("Failed to create boot volume", err)
			return
		}
		metadata := ""
		var ifaces []*model.Interface
		// cloud-init does not support set encrypted password for windows
		// so we only encrypt the password for linux and others
		if rootPasswd != "" && image.OSCode != "windows" {
			rootPasswd, err = encrpt.Mkpasswd(rootPasswd, "sha512")
			if err != nil {
				logger.Errorf("Failed to encrypt admin password, %v", err)
				return
			}
		}
		ifaces, metadata, err = a.buildMetadata(ctx, primaryIface, secondaryIfaces, rootPasswd, keys, instance, userdata, routerID, zoneID, "")
		if err != nil {
			logger.Error("Build instance metadata failed", err)
			return
		}
		instance.Interfaces = ifaces
		rcNeeded := fmt.Sprintf("cpu=%d memory=%d disk=%d network=%d", flavor.Cpu, flavor.Memory*1024, (flavor.Disk+flavor.Swap+flavor.Ephemeral)*1024*1024, 0)
		control := "select=" + hyperGroup + " " + rcNeeded
		if i == 0 && hyperID >= 0 {
			control = fmt.Sprintf("inter=%d %s", hyperID, rcNeeded)
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/launch_vm.sh '%d' '%s.%s' '%t' '%d' '%s' '%d' '%d' '%d' '%d'<<EOF\n%s\nEOF", instance.ID, imagePrefix, image.Format, image.QAEnabled, snapshot, hostname, flavor.Cpu, flavor.Memory, flavor.Disk, bootVolume.ID, base64.StdEncoding.EncodeToString([]byte(metadata)))
		execCommands = append(execCommands, &ExecutionCommand{
			Control: control,
			Command: command,
		})
		instances = append(instances, instance)
		i++
	}
	a.executeCommandList(ctx, execCommands)
	return
}

func (a *InstanceAdmin) executeCommandList(ctx context.Context, cmdList []*ExecutionCommand) {
	var err error
	for _, cmd := range cmdList {
		err = HyperExecute(ctx, cmd.Control, cmd.Command)
		if err != nil {
			logger.Error("Command execution failed", err)
		}
	}
	return
}

func (a *InstanceAdmin) ChangeInstanceStatus(ctx context.Context, instance *model.Instance, action string) (err error) {
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/action_vm.sh '%d' '%s'", instance.ID, action)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Delete vm command execution failed", err)
		return
	}
	return
}

func (a *InstanceAdmin) Migrate(ctx context.Context, instance *model.Instance, fromHyper, toHyper *model.Hyper) (err error) {
	memberShip := GetMemberShip(ctx)
	permit, err := memberShip.CheckAdmin(model.Admin, "instances", instance.ID)
	if err != nil {
		logger.Error("Failed to check owner")
		return
	}
	if !permit {
		logger.Error("Not authorized to migrate the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	if toHyper.Hostid == instance.Hyper {
			logger.Error("No need to migrate the instance to the same host")
			err = fmt.Errorf("No need to migrate the instance to the same host")
			return
	}
	metadata, err := a.getMetadata(ctx, instance)
	if err != nil {
		logger.Error("Failed to get metadata")
		return
	}
	flavor := instance.Flavor
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/pre_migrate_vm.sh '%d' '%t' '%s' '%d' '%d' '%d'<<EOF\n%s\nEOF", instance.ID, instance.Image.QAEnabled, instance.Hostname, flavor.Cpu, flavor.Memory, flavor.Disk, base64.StdEncoding.EncodeToString([]byte(metadata)))
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Pre migrate vm command execution failed", err)
		return
	}
	return
}

func (a *InstanceAdmin) Update(ctx context.Context, instance *model.Instance, flavor *model.Flavor, hostname string, action PowerAction, hyperID int) (err error) {
	memberShip := GetMemberShip(ctx)
	permit, err := memberShip.CheckOwner(model.Writer, "instances", instance.ID)
	if err != nil {
		logger.Error("Failed to check owner")
		return
	}
	if !permit {
		logger.Error("Not authorized to delete the instance")
		err = fmt.Errorf("Not authorized")
		return
	}

	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	if hyperID != int(instance.Hyper) {
		permit, err = memberShip.CheckAdmin(model.Admin, "instances", instance.ID)
		if !permit {
			logger.Error("Not authorized to migrate VM")
			err = fmt.Errorf("Not authorized to migrate VM")
			return
		}
		// TODO: migrate VM
	}
	if flavor != nil && flavor.ID != instance.FlavorID {
		if instance.Status == "running" {
			err = fmt.Errorf("Instance must be shutdown first before resize")
			logger.Error(err)
			return
		}
		if flavor.Disk < instance.Flavor.Disk {
			err = fmt.Errorf("Disk(s) can not be resized to smaller size")
			logger.Error("Disk(s) can not be resized to smaller size")
			return
		}
		cpu := flavor.Cpu - instance.Flavor.Cpu
		if cpu < 0 {
			cpu = 0
		}
		memory := flavor.Memory - instance.Flavor.Memory
		if memory < 0 {
			memory = 0
		}
		disk := flavor.Disk - instance.Flavor.Disk + flavor.Ephemeral - instance.Flavor.Ephemeral
		control := fmt.Sprintf("inter=%d cpu=%d memory=%d disk=%d network=%d", instance.Hyper, cpu, memory, disk, 0)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/resize_vm.sh '%d' '%d' '%d' '%d' '%d' '%d' '%d'", instance.ID, flavor.Cpu, flavor.Memory, flavor.Disk, flavor.Swap, flavor.Ephemeral, disk)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Resize vm command execution failed", err)
			return
		}
		instance.FlavorID = flavor.ID
		instance.Flavor = flavor
	}
	if instance.Hostname != hostname {
		instance.Hostname = hostname
	}
	if err = db.Model(instance).Updates(instance).Error; err != nil {
		logger.Error("Failed to save instance", err)
		return
	}
	if string(action) != "" {
		err = instanceAdmin.ChangeInstanceStatus(ctx, instance, string(action))
		if err != nil {
			logger.Error("action vm command execution failed", err)
			return
		}
	}
	return
}

func (a *InstanceAdmin) SetUserPassword(ctx context.Context, id int64, user, password string) (err error) {
	logger.Debugf("Set password for user %s of instance %d", user, id)
	ctx, db := GetContextDB(ctx)
	instance := &model.Instance{Model: model.Model{ID: id}}
	if err = db.Preload("Image").Take(instance).Error; err != nil {
		logger.Error("Failed to get instance ", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit, err := memberShip.CheckOwner(model.Writer, "instances", instance.ID)
	if err != nil {
		logger.Error("Failed to check owner")
		return
	}
	if !permit {
		logger.Error("Not authorized to set password for the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	if !instance.Image.QAEnabled {
		err = fmt.Errorf("Guest Agent is not enabled for the image of instance")
		logger.Error(err)
		return
	}
	if instance.Status != "running" {
		err = fmt.Errorf("Instance must be running")
		logger.Error(err)
		return
	}
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_user_passwd.sh '%d' '%s' '%s'", instance.ID, user, password)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Set password command execution failed", err)
		return
	}
	return
}

func (a *InstanceAdmin) deleteInterfaces(ctx context.Context, instance *model.Instance) (err error) {
	for _, iface := range instance.Interfaces {
		err = a.deleteInterface(ctx, iface)
		if err != nil {
			logger.Error("Failed to delete interface", err)
			continue
		}
	}
	return
}

func (a *InstanceAdmin) deleteInterface(ctx context.Context, iface *model.Interface) (err error) {
	err = DeleteInterface(ctx, iface)
	if err != nil {
		logger.Error("Failed to create interface")
		return
	}
	vlan := iface.Address.Subnet.Vlan
	control := ""
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/del_host.sh '%d' '%s' '%s'", vlan, iface.MacAddr, iface.Address.Address)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Delete interface failed")
		return
	}
	return
}

func (a *InstanceAdmin) createInterface(ctx context.Context, subnet *model.Subnet, address, mac string, instance *model.Instance, ifname string, inbound, outbound int32, secgroups []*model.SecurityGroup, zoneID int64) (iface *model.Interface, err error) {
	memberShip := GetMemberShip(ctx)
	if subnet.Type == "public" {
		permit := memberShip.CheckPermission(model.Admin)
		if !permit {
			logger.Error("Not authorized to create interface in public subnet")
			err = fmt.Errorf("Not authorized")
			return
		}
	}
	iface, err = CreateInterface(ctx, subnet, instance.ID, memberShip.OrgID, instance.Hyper, inbound, outbound, address, mac, ifname, "instance", secgroups)
	if err != nil {
		logger.Error("Failed to create interface")
		return
	}
	return
}

func (a *InstanceAdmin) getMetadata(ctx context.Context, instance *model.Instance) (metadata string, err error) {
	vlans := []*VlanInfo{}
	instNetworks := []*InstanceNetwork{}
	instLinks := []*NetworkLink{}
	volumes := []*VolumeInfo{}
	var instKeys []string
	for _, key := range instance.Keys {
		instKeys = append(instKeys, key.PublicKey)
	}
	for _, volume := range instance.Volumes {
		volumes = append(volumes, &VolumeInfo{
			ID: volume.ID,
			UUID: volume.GetOriginVolumeID(),
			Device: volume.Target,
			Booting: volume.Booting,
		})
	}
	dns := ""
	for i, iface := range instance.Interfaces {
		subnet := iface.Address.Subnet
		instNetwork := &InstanceNetwork{
                        Address: iface.Address.Address,
                        Netmask: subnet.Netmask,
                        Type:    "ipv4",
                        Link:    iface.Name,
                        ID:      fmt.Sprintf("network%d", i+1),
                }
		if iface.PrimaryIf {
			instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: subnet.Gateway}
			instNetwork.Routes = append(instNetwork.Routes, instRoute)
			dns = subnet.NameServer
		}
		instNetworks = append(instNetworks, instNetwork)
		instLinks = append(instLinks, &NetworkLink{MacAddr: iface.MacAddr, Mtu: uint(iface.Mtu), ID: iface.Name, Type: "phy"})
		vlans = append(vlans, &VlanInfo{Device: iface.Name, Vlan: subnet.Vlan, Inbound: iface.Inbound, Outbound: iface.Outbound, AllowSpoofing: iface.AllowSpoofing, Gateway: subnet.Gateway, Router: subnet.RouterID, IpAddr: iface.Address.Address, MacAddr: iface.MacAddr})
	}
	instData := &InstanceData{
		Userdata:   instance.Userdata,
		DNS:        dns,
		Vlans:      vlans,
		Networks:   instNetworks,
		Links:      instLinks,
		Volumes:    volumes,
		Keys:       instKeys,
		RootPasswd: "",
		OSCode:     instance.Image.OSCode,
	}
	jsonData, err := json.Marshal(instData)
	if err != nil {
		logger.Errorf("Failed to marshal instance json data, %v", err)
		return
	}
	return string(jsonData), nil
}

func (a *InstanceAdmin) buildMetadata(ctx context.Context, primaryIface *InterfaceInfo, secondaryIfaces []*InterfaceInfo,
	rootPasswd string, keys []*model.Key, instance *model.Instance, userdata string, routerID, zoneID int64,
	service string) (interfaces []*model.Interface, metadata string, err error) {
	if rootPasswd == "" {
		logger.Debugf("Build instance metadata with primaryIface: %v, secondaryIfaces: %+v, keys: %+v, instance: %+v, userdata: %s, routerID: %d, zoneID: %d, service: %s",
			primaryIface, secondaryIfaces, keys, instance, userdata, routerID, zoneID, service)
	} else {
		logger.Debugf("Build instance metadata with primaryIface: %v, secondaryIfaces: %+v, keys: %+v, instance: %+v, userdata: %s, routerID: %d, zoneID: %d, service: %s, root password: %s",
			primaryIface, secondaryIfaces, keys, instance, userdata, routerID, zoneID, service, "******")
	}
	vlans := []*VlanInfo{}
	instNetworks := []*InstanceNetwork{}
	instLinks := []*NetworkLink{}
	primary := primaryIface.Subnet
	primaryIP := primaryIface.IpAddress
	primaryMac := primaryIface.MacAddress
	inbound := primaryIface.Inbound
	outbound := primaryIface.Outbound
	gateway := strings.Split(primary.Gateway, "/")[0]
	instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: gateway}
	iface, err := a.createInterface(ctx, primary, primaryIP, primaryMac, instance, "eth0", inbound, outbound, primaryIface.SecurityGroups, zoneID)
	if err != nil {
		logger.Errorf("Allocate address for primary subnet %s--%s/%s failed, %v", primary.Name, primary.Network, primary.Netmask, err)
		return
	}
	interfaces = append(interfaces, iface)
	address := strings.Split(iface.Address.Address, "/")[0]
	instNetwork := &InstanceNetwork{Address: address, Netmask: primary.Netmask, Type: "ipv4", Link: iface.Name, ID: "network0"}
	instNetwork.Routes = append(instNetwork.Routes, instRoute)
	instNetworks = append(instNetworks, instNetwork)
	instLinks = append(instLinks, &NetworkLink{MacAddr: iface.MacAddr, Mtu: uint(iface.Mtu), ID: iface.Name, Type: "phy"})
	securityData, err := GetSecurityData(ctx, iface.SecurityGroups)
	if err != nil {
		logger.Error("Get security data for interface failed", err)
		return
	}
	vlans = append(vlans, &VlanInfo{Device: "eth0", Vlan: primary.Vlan, Inbound: inbound, Outbound: outbound, AllowSpoofing: iface.AllowSpoofing, Gateway: primary.Gateway, Router: primary.RouterID, IpAddr: iface.Address.Address, MacAddr: iface.MacAddr, SecRules: securityData})
	for i, ifaceInfo := range secondaryIfaces {
		subnet := ifaceInfo.Subnet
		ifname := fmt.Sprintf("eth%d", i+1)
		inbound = ifaceInfo.Inbound
		outbound = ifaceInfo.Outbound
		iface, err = a.createInterface(ctx, subnet, ifaceInfo.IpAddress, ifaceInfo.MacAddress, instance, ifname, inbound, outbound, ifaceInfo.SecurityGroups, zoneID)
		if err != nil {
			logger.Errorf("Allocate address for secondary subnet %s--%s/%s failed, %v", subnet.Name, subnet.Network, subnet.Netmask, err)
			return
		}
		interfaces = append(interfaces, iface)
		address = strings.Split(iface.Address.Address, "/")[0]
		instNetworks = append(instNetworks, &InstanceNetwork{
			Address: address,
			Netmask: subnet.Netmask,
			Type:    "ipv4",
			Link:    iface.Name,
			ID:      fmt.Sprintf("network%d", i+1),
		})
		instLinks = append(instLinks, &NetworkLink{MacAddr: iface.MacAddr, Mtu: uint(iface.Mtu), ID: iface.Name, Type: "phy"})
		securityData, err = GetSecurityData(ctx, iface.SecurityGroups)
		if err != nil {
			logger.Error("Get security data for interface failed", err)
			return
		}
		vlans = append(vlans, &VlanInfo{Device: ifname, Vlan: subnet.Vlan, Inbound: inbound, Outbound: outbound, AllowSpoofing: iface.AllowSpoofing, Gateway: subnet.Gateway, Router: subnet.RouterID, IpAddr: iface.Address.Address, MacAddr: iface.MacAddr, SecRules: securityData})
	}
	var instKeys []string
	for _, key := range keys {
		instKeys = append(instKeys, key.PublicKey)
	}
	image := &model.Image{Model: model.Model{ID: instance.ImageID}}
	err = DB().Take(image).Error
	if err != nil {
		logger.Error("Invalid image ", instance.ImageID)
		return
	}
	virtType := image.VirtType
	dns := primary.NameServer
	if dns == primaryIP {
		dns = ""
	}
	instData := &InstanceData{
		Userdata:   userdata,
		VirtType:   virtType,
		DNS:        dns,
		Vlans:      vlans,
		Networks:   instNetworks,
		Links:      instLinks,
		Keys:       instKeys,
		RootPasswd: rootPasswd,
		OSCode:     image.OSCode,
	}
	jsonData, err := json.Marshal(instData)
	if err != nil {
		logger.Errorf("Failed to marshal instance json data, %v", err)
		return
	}
	return interfaces, string(jsonData), nil
}

func (a *InstanceAdmin) Delete(ctx context.Context, instance *model.Instance) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, instance.Owner)
	if !permit {
		logger.Error("Not authorized to delete the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		logger.Errorf("Failed to query floating ip(s), %v", err)
		return
	}
	if instance.FloatingIps != nil {
		for _, fip := range instance.FloatingIps {
			fip.Instance = instance
			err = floatingIpAdmin.Detach(ctx, fip)
			if err != nil {
				logger.Errorf("Failed to detach floating ip, %v", err)
				return
			}
		}
		instance.FloatingIps = nil
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.Volumes).Error; err != nil {
		logger.Errorf("Failed to query floating ip(s), %v", err)
		return
	}
	bootVolumeUUID := ""
	if instance.Volumes != nil {
		for _, volume := range instance.Volumes {
			if volume.Booting {
				bootVolumeUUID = volume.GetOriginVolumeID()
				// delete the boot volume directly
				if err = db.Delete(volume).Error; err != nil {
					logger.Error("DB: delete volume failed", err)
					return
				}
			} else {
				_, err = volumeAdmin.Update(ctx, volume.ID, "", 0)
				if err != nil {
					logger.Error("Failed to detach volume, %v", err)
					return
				}
			}
		}
		instance.Volumes = nil
	}
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	if instance.Hyper == -1 {
		control = "toall="
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_vm.sh '%d' '%d' '%s'", instance.ID, instance.RouterID, bootVolumeUUID)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Delete vm command execution failed ", err)
		return
	}
	instance.Status = "deleting"
	err = db.Model(instance).Updates(instance).Error
	if err != nil {
		logger.Errorf("Failed to mark vm as deleting ", err)
		return
	}
	return
}

func (a *InstanceAdmin) Get(ctx context.Context, id int64) (instance *model.Instance, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid instance ID: %d", id)
		logger.Error(err)
		return
	}
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	instance = &model.Instance{Model: model.Model{ID: id}}
	if err = db.Preload("Volumes").Preload("Image").Preload("Zone").Preload("Flavor").Preload("Keys").Where(where).Take(instance).Error; err != nil {
		logger.Errorf("Failed to query instance, %v", err)
		return
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		logger.Errorf("Failed to query floating ip(s), %v", err)
		return
	}
	if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where("instance = ?", instance.ID).Find(&instance.Interfaces).Error; err != nil {
		logger.Errorf("Failed to query interfaces %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, instance.Owner)
	if !permit {
		logger.Error("Not authorized to read the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	permit = memberShip.CheckPermission(model.Admin)
	if permit {
		instance.OwnerInfo = &model.Organization{Model: model.Model{ID: instance.Owner}}
		if err = db.Take(instance.OwnerInfo).Error; err != nil {
			logger.Error("Failed to query owner info", err)
			return
		}
	}

	return
}

func (a *InstanceAdmin) GetInstanceByUUID(ctx context.Context, uuID string) (instance *model.Instance, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	instance = &model.Instance{}
	if err = db.Preload("Volumes").Preload("Image").Preload("Zone").Preload("Flavor").Preload("Keys").Where(where).Where("uuid = ?", uuID).Take(instance).Error; err != nil {
		logger.Errorf("Failed to query instance, %v", err)
		return
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		logger.Errorf("Failed to query floating ip(s), %v", err)
		return
	}
	if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where("instance = ?", instance.ID).Find(&instance.Interfaces).Error; err != nil {
		logger.Errorf("Failed to query interfaces %v", err)
		return
	}
	if instance.RouterID > 0 {
		instance.Router = &model.Router{Model: model.Model{ID: instance.RouterID}}
		if err = db.Take(instance.Router).Error; err != nil {
			logger.Errorf("Failed to query floating ip(s), %v", err)
			return
		}
	}
	permit := memberShip.ValidateOwner(model.Reader, instance.Owner)
	if !permit {
		logger.Error("Not authorized to read the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *InstanceAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, instances []*model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}
	logger.Debugf("The query in admin console is %s", query)

	where := memberShip.GetWhere()
	instances = []*model.Instance{}
	if err = db.Model(&model.Instance{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Volumes").Preload("Image").Preload("Zone").Preload("Flavor").Preload("Keys").Where(where).Where(query).Find(&instances).Error; err != nil {
		logger.Errorf("Failed to query instance(s), %v", err)
		return
	}
	db = db.Offset(0).Limit(-1)
	for _, instance := range instances {
		if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where("instance = ?", instance.ID).Find(&instance.Interfaces).Error; err != nil {
			logger.Errorf("Failed to query interfaces %v", err)
			return
		}
		if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
			logger.Errorf("Failed to query floating ip(s), %v", err)
			return
		}
		if instance.RouterID > 0 {
			instance.Router = &model.Router{Model: model.Model{ID: instance.RouterID}}
			if err = db.Take(instance.Router).Error; err != nil {
				logger.Errorf("Failed to query floating ip(s), %v", err)
				return
			}
		}
		permit := memberShip.CheckPermission(model.Admin)
		if permit {
			instance.OwnerInfo = &model.Organization{Model: model.Model{ID: instance.Owner}}
			if err = db.Take(instance.OwnerInfo).Error; err != nil {
				logger.Error("Failed to query owner info", err)
				return
			}
		}
	}

	return
}

func (v *InstanceView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	hostname := c.QueryTrim("hostname")
	router_id := c.QueryTrim("router_id")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	if query != "" {
		query = fmt.Sprintf("hostname like '%%%s%%'", query)
	}
	if router_id != "" {
		routerID, err := strconv.Atoi(router_id)
		if err != nil {
				logger.Debugf("Error to convert router_id to integer: %+v ", err)
		}	
		query = fmt.Sprintf("router_id = %d", routerID)
	}

	total, instances, err := instanceAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["Instances"] = instances
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.Data["HostName"] = hostname
	c.HTML(200, "instances")
}

func (v *InstanceView) UpdateTable(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	_, instances, err := instanceAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	var jsonData *InstancesData
	jsonData = &InstancesData{
		Instances: instances,
		IsAdmin:   memberShip.CheckPermission(model.Admin),
	}

	c.JSON(200, jsonData)
	return
}

func (v *InstanceView) Status(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}

	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instance, err := instanceAdmin.Get(ctx, int64(instanceID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Data["Instance"] = instance
	logger.Debugf("Instance status %+v", instance)
	c.HTML(200, "instances_status")
}

func (v *InstanceView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	instance, err := instanceAdmin.Get(ctx, int64(instanceID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = instanceAdmin.Delete(ctx, instance)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "instances",
	})
	return
}

func (v *InstanceView) New(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	images := []*model.Image{}
	if err := db.Find(&images).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	flavors := []*model.Flavor{}
	if err := db.Find(&flavors).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, subnets, err := subnetAdmin.List(ctx, 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, secgroups, err := secgroupAdmin.List(ctx, 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, keys, err := keyAdmin.List(ctx, 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	hypers := []*model.Hyper{}
	err = db.Where("hostid >= 0").Find(&hypers).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	zones := []*model.Zone{}
	err = db.Find(&zones).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["HostName"] = c.QueryTrim("hostname")
	c.Data["Images"] = images
	c.Data["Flavors"] = flavors
	c.Data["Subnets"] = subnets
	c.Data["SecurityGroups"] = secgroups
	c.Data["Keys"] = keys
	c.Data["Hypers"] = hypers
	c.Data["Zones"] = zones
	c.HTML(200, "instances_new")
}

func (v *InstanceView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instanceID, err := strconv.Atoi(id)

	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instanceID)}}
	if err = db.Preload("Interfaces").Take(instance).Error; err != nil {
		logger.Error("Instance query failed", err)
		return
	}
	if err = db.Where("instance_id = ?", instanceID).Find(&instance.FloatingIps).Error; err != nil {
		logger.Errorf("Failed to query floating ip(s), %v", err)
		return
	}
	_, subnets, err := subnetAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	for _, iface := range instance.Interfaces {
		for i, subnet := range subnets {
			if iface == nil || iface.Address == nil {
				continue
			}
			if subnet.ID == iface.Address.SubnetID {
				subnets = append(subnets[:i], subnets[i+1:]...)
				break
			}
		}
	}
	_, flavors, err := flavorAdmin.List(0, -1, "", "")
	if err := db.Find(&flavors).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Instance"] = instance
	c.Data["Subnets"] = subnets
	c.Data["Flavors"] = flavors

	flag := c.QueryTrim("flag")
	logger.Debugf("Edit instance %s with flag %s", id, flag)
	if flag == "ChangeHostname" {
		c.HTML(200, "instances_hostname")
	} else if flag == "ChangeStatus" {
		if c.QueryTrim("action") != "" {
			ctx := c.Req.Context()
			instanceID64, vmError := strconv.ParseInt(id, 10, 64)
			if vmError != nil {
				logger.Error("Change String to int64 failed", err)
				return
			}
			instance, vmError = instanceAdmin.Get(ctx, instanceID64)
			if vmError != nil {
				logger.Error("Get instance failed", err)
				return
			}
			vmError = instanceAdmin.ChangeInstanceStatus(ctx, instance, c.QueryTrim("action"))
			if vmError != nil {
				logger.Error("Instance action command execution failed", err)
				return
			}
			redirectTo := "../instances"
			c.Redirect(redirectTo)
		} else {
			c.HTML(200, "instances_status")
		}
	} else if flag == "MigrateInstance" {
		c.HTML(200, "instances_migrate")
	} else if flag == "ResizeInstance" {
		c.HTML(200, "instances_size")
	} else {
		c.HTML(200, "instances_patch")
	}
}

func (v *InstanceView) Patch(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "../instances"
	instanceID := c.ParamsInt64("id")
	flavorID := c.QueryInt64("flavor")
	hostname := c.QueryTrim("hostname")
	hyperID := c.QueryInt("hyper")
	action := c.QueryTrim("action")
	instance, err := instanceAdmin.Get(ctx, instanceID)
	if err != nil {
		logger.Error("Invalid instance", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	var flavor *model.Flavor
	if flavorID > 0 {
		flavor, err = flavorAdmin.Get(ctx, flavorID)
		if err != nil {
			logger.Error("Invalid flavor", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	err = instanceAdmin.Update(c.Req.Context(), instance, flavor, hostname, PowerAction(action), hyperID)
	if err != nil {
		logger.Errorf("update instance failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}

func (v *InstanceView) SetUserPassword(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "/instances"
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		logger.Error("Instance ID error ", err)
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instanceID)}}
	if err = db.Preload("Image").Take(instance).Error; err != nil {
		logger.Error("Instance query failed", err)
		c.Data["ErrorMsg"] = fmt.Sprintf("Instance query failed", err)
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if c.Req.Method == "GET" {
		if instance.Image.QAEnabled {
			c.Data["Instance"] = instance
			c.Data["Link"] = fmt.Sprintf("/instances/%d/set_user_password", instanceID)
			c.HTML(200, "instances_user_passwd")
		} else {
			c.Data["ErrorMsg"] = "Guest Agent is not enabled for the image of instance"
			c.HTML(http.StatusBadRequest, "error")
		}
		return
	} else if c.Req.Method == "POST" {
		user := c.QueryTrim("username")
		password := c.QueryTrim("password")
		err := instanceAdmin.SetUserPassword(ctx, int64(instanceID), user, password)
		if err != nil {
			logger.Error("Set user password failed", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		c.Redirect(redirectTo)

	}
}

func (v *InstanceView) checkNetparam(subnet *model.Subnet, IP, mac string) (macAddr string, err error) {
	_, inNet, err := net.ParseCIDR(subnet.Network)
	if err != nil {
		logger.Error("CIDR parsing failed ", err)
		return
	}
	if IP != "" && !inNet.Contains(net.ParseIP(IP)) {
		logger.Errorf("Primary IP %s not belonging to subnet %v\n", IP, subnet)
		err = fmt.Errorf("Primary IP not belonging to subnet")
		return
	}
	if mac != "" {
		macl := strings.Split(mac, ":")
		if len(macl) != 6 {
			logger.Errorf("Invalid mac address format: %s", mac)
			err = fmt.Errorf("Invalid mac address format: %s", mac)
			return
		}
		macAddr = strings.ToLower(mac)
		var tmp [6]int
		_, err = fmt.Sscanf(macAddr, "%02x:%02x:%02x:%02x:%02x:%02x", &tmp[0], &tmp[1], &tmp[2], &tmp[3], &tmp[4], &tmp[5])
		if err != nil {
			logger.Error("Failed to parse mac address")
			return
		}
		if tmp[0]%2 == 1 {
			logger.Error("Not a valid unicast mac address")
			err = fmt.Errorf("Not a valid unicast mac address")
			return
		}
	}
	return
}

func (v *InstanceView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Need Write permissions")
		c.Data["ErrorMsg"] = "Need Write permissions"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../instances"
	hostname := c.QueryTrim("hostname")
	rootPasswd := c.QueryTrim("rootpasswd")
	cnt := c.QueryTrim("count")
	count, err := strconv.Atoi(cnt)
	if err != nil {
		logger.Error("Invalid instance count", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	hyperID := c.QueryInt("hyper")
	if hyperID >= 0 {
		permit := memberShip.CheckPermission(model.Admin)
		if !permit {
			logger.Error("Need Admin permissions")
			c.Data["ErrorMsg"] = "Need Admin permissions"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	imageID := c.QueryInt64("image")
	if imageID <= 0 {
		logger.Error("No valid image ID", imageID)
		c.Data["ErrorMsg"] = "No valid image ID"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	image, err := imageAdmin.Get(ctx, imageID)
	if err != nil {
		c.Data["ErrorMsg"] = "No valid image"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavorID := c.QueryInt64("flavor")
	if flavorID <= 0 {
		logger.Error("Invalid flavor ID", flavorID)
		c.Data["ErrorMsg"] = "Invalid flavor ID"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor, err := flavorAdmin.Get(ctx, flavorID)
	if err != nil {
		logger.Error("No valid flavor", err)
		c.Data["ErrorMsg"] = "No valid flavor"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	zoneID := c.QueryInt64("zone")
	zone, err := zoneAdmin.Get(ctx, zoneID)
	if err != nil {
		logger.Error("No valid zone", err)
		c.Data["ErrorMsg"] = "No valid zone"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	primary := c.QueryTrim("primary")
	primaryID, err := strconv.Atoi(primary)
	if err != nil {
		logger.Error("Invalid primary subnet ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	primaryIP := c.QueryTrim("primaryip")
	ipAddr := strings.Split(primaryIP, "/")[0]
	primaryMac := c.QueryTrim("primarymac")
	primarySubnet, err := subnetAdmin.Get(ctx, int64(primaryID))
	if err != nil {
		logger.Error("Get primary subnet failed", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	macAddr, err := v.checkNetparam(primarySubnet, ipAddr, primaryMac)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroups := c.QueryTrim("secgroups")
	var securityGroups []*model.SecurityGroup
	if secgroups != "" {
		sg := strings.Split(secgroups, ",")
		for i := 0; i < len(sg); i++ {
			sgID, err := strconv.Atoi(sg[i])
			if err != nil {
				continue
			}
			var secgroup *model.SecurityGroup
			secgroup, err = secgroupAdmin.Get(ctx, int64(sgID))
			if err != nil {
				logger.Error("Get security groups failed", err)
				c.Data["ErrorMsg"] = "Get security groups failed"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			if secgroup.RouterID != primarySubnet.RouterID {
				logger.Error("Security group is not the same router with subnet")
				c.Data["ErrorMsg"] = "Security group is not in subnet vpc"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			securityGroups = append(securityGroups, secgroup)
		}
	} else {
		var sgID int64
		routerID := primarySubnet.RouterID
		if routerID > 0 {
			var router *model.Router
			router, err = routerAdmin.Get(ctx, routerID)
			if err != nil {
				logger.Error("Get router failed", err)
				c.Data["ErrorMsg"] = "Get router failed"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			sgID = router.DefaultSG
		}
		var secGroup *model.SecurityGroup
		secGroup, err = secgroupAdmin.Get(ctx, int64(sgID))
		if err != nil {
			logger.Error("Get security groups failed", err)
			c.Data["ErrorMsg"] = "Get security groups failed"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		securityGroups = append(securityGroups, secGroup)
	}
	primaryIface := &InterfaceInfo{
		Subnet:         primarySubnet,
		IpAddress:      ipAddr,
		MacAddress:     macAddr,
		SecurityGroups: securityGroups,
		Inbound:        1000,
		Outbound:       1000,
	}
	subnets := c.QueryTrim("subnets")
	var secondaryIfaces []*InterfaceInfo
	s := strings.Split(subnets, ",")
	for i := 0; i < len(s); i++ {
		sID, err := strconv.Atoi(s[i])
		if err != nil {
			logger.Error("Invalid secondary subnet ID", err)
			continue
		}
		var subnet *model.Subnet
		subnet, err = subnetAdmin.Get(ctx, int64(sID))
		if err != nil {
			logger.Error("Get secondary subnet failed", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		if subnet.RouterID != primarySubnet.RouterID {
			logger.Error("All subnets must be in the same vpc", err)
			c.Data["ErrorMsg"] = "All subnets must be in the same vpc"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		secondaryIfaces = append(secondaryIfaces, &InterfaceInfo{
			Subnet:         subnet,
			IpAddress:      "",
			MacAddress:     "",
			SecurityGroups: securityGroups,
			Inbound:        1000,
			Outbound:       1000,
		})
	}
	keys := c.QueryTrim("keys")
	k := strings.Split(keys, ",")
	var instKeys []*model.Key
	for i := 0; i < len(k); i++ {
		kID, err := strconv.Atoi(k[i])
		if err != nil {
			logger.Error("Invalid key ID", err)
			continue
		}
		var key *model.Key
		key, err = keyAdmin.Get(ctx, int64(kID))
		if err != nil {
			logger.Error("Failed to access key", err)
			c.Data["ErrorMsg"] = "Failed to access key"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		instKeys = append(instKeys, key)
	}
	userdata := c.QueryTrim("userdata")
	_, err = instanceAdmin.Create(ctx, count, hostname, userdata, image, flavor, zone, primarySubnet.RouterID, primaryIface, secondaryIfaces, instKeys, rootPasswd, hyperID)
	if err != nil {
		logger.Error("Create instance failed", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}
