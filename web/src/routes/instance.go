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
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"
	"web/src/rpcs"

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

type ExecutionCommand struct{
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

type InstanceData struct {
	Userdata string             `json:"userdata"`
	VirtType string             `json:"virt_type"`
	DNS      string             `json:"dns"`
	Vlans    []*VlanInfo        `json:"vlans"`
	Networks []*InstanceNetwork `json:"networks"`
	Links    []*NetworkLink     `json:"links"`
	Keys     []string           `json:"keys"`
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
		log.Println("Hypers query failed", err)
		return
	}
	if len(hypers) == 0 {
		log.Println("No qualified hypervisor")
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

func (a *InstanceAdmin) Create(ctx context.Context, count int, prefix, userdata string, image *model.Image, flavor *model.Flavor, zone *model.Zone, routerID int64, primaryIface *InterfaceInfo, secondaryIfaces []*InterfaceInfo, keys []*model.Key, hyperID int) (instances []*model.Instance, err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	if image.Status != "available" {
		err = fmt.Errorf("Image status not available")
		log.Println("Image status not available")
		return
	}
	zoneID := zone.ID
	if hyperID >= 0 {
		hyper := &model.Hyper{}
		err = db.Where("hostid = ?", hyperID).Take(hyper).Error
		if err != nil {
			log.Println("Failed to query hypervisor", err)
			return
		}
		if hyper.ZoneID != zone.ID {
			log.Printf("Hypervisor %v is not in zone %d, %v", hyper, zoneID, err)
			err = fmt.Errorf("Hypervisor is not in this zone")
			return
		}
	}
	hyperGroup, err := a.getHyperGroup(ctx, image.VirtType, zoneID)
	if err != nil {
		log.Println("No valid hypervisor", err)
		return
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
			log.Println("Failed to query total instances with the image", err)
			return
		}
		snapshot := total/MaxmumSnapshot + 1 // Same snapshot reference can not be over 128, so use 96 here
		instance := &model.Instance{Model: model.Model{Creater: memberShip.UserID}, Owner: memberShip.OrgID, Hostname: hostname, ImageID: image.ID, Snapshot: int64(snapshot), FlavorID: flavor.ID, Keys: keys, Userdata: userdata, Status: "pending", ZoneID: zoneID, RouterID: routerID}
		err = db.Create(instance).Error
		if err != nil {
			log.Println("DB create instance failed", err)
			return
		}
		instance.Image = image
		instance.Flavor = flavor
		instance.Zone = zone
		var bootVolume *model.Volume
		imagePrefix := fmt.Sprintf("image-%d-%s", image.ID, strings.Split(image.UUID, "-")[0])
		// boot volume name format: instance-15-image-2-3c0cca59-boot-volume-10
		bootVolume, err = volumeAdmin.CreateVolume(ctx, fmt.Sprintf("instance-%d-%s-boot-volume", instance.ID, imagePrefix), flavor.Disk, instance.ID, true, 0, 0, 0, 0, "")
		if err != nil {
			log.Println("Failed to create boot volume", err)
			return
		}
		metadata := ""
		var ifaces []*model.Interface
		ifaces, metadata, err = a.buildMetadata(ctx, primaryIface, secondaryIfaces, keys, instance, userdata, routerID, zoneID, "")
		if err != nil {
			log.Println("Build instance metadata failed", err)
			return
		}
		instance.Interfaces = ifaces
		rcNeeded := fmt.Sprintf("cpu=%d memory=%d disk=%d network=%d", flavor.Cpu, flavor.Memory*1024, (flavor.Disk+flavor.Swap+flavor.Ephemeral)*1024*1024, 0)
		control := "select=" + hyperGroup + " " + rcNeeded
		if i == 0 && hyperID >= 0 {
			control = fmt.Sprintf("inter=%d %s", hyperID, rcNeeded)
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/launch_vm.sh '%d' '%s.%s' '%d' '%s' '%d' '%d' '%d' '%d'<<EOF\n%s\nEOF", instance.ID, imagePrefix, image.Format, snapshot, hostname, flavor.Cpu, flavor.Memory, flavor.Disk, bootVolume.ID, base64.StdEncoding.EncodeToString([]byte(metadata)))
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
		err = hyperExecute(ctx, cmd.Control, cmd.Command)
		if err != nil {
			log.Println("Command execution failed", err)
		}
	}
	return
}

func (a *InstanceAdmin) ChangeInstanceStatus(ctx context.Context, id int64, action string) (instance *model.Instance, err error) {
	ctx, db := GetContextDB(ctx)
	instance = &model.Instance{Model: model.Model{ID: id}}
	if err = db.Take(instance).Error; err != nil {
		log.Println("Failed to query instance ", err)
		return
	}
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/action_vm.sh '%d' '%s'", instance.ID, action)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete vm command execution failed", err)
		return
	}
	return
}

func (a *InstanceAdmin) Update(ctx context.Context, instance *model.Instance, flavor *model.Flavor, hostname string, action PowerAction, hyperID int) (err error) {
	memberShip := GetMemberShip(ctx)
	permit, err := memberShip.CheckOwner(model.Writer, "instances", instance.ID)
	if err != nil {
		log.Println("Failed to check owner")
		return
	}
	if !permit {
		log.Println("Not authorized to delete the instance")
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
			log.Println("Not authorized to migrate VM")
			err = fmt.Errorf("Not authorized to migrate VM")
			return
		}
		// TODO: migrate VM
	}
	if flavor != nil && flavor.ID != instance.FlavorID {
		if instance.Status == "running" {
			err = fmt.Errorf("Instance must be shutdown first before resize")
			log.Println("Instance must be shutdown first before resize", err)
			return
		}
		if flavor.Disk < instance.Flavor.Disk {
			err = fmt.Errorf("Disk(s) can not be resized to smaller size")
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
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Resize vm command execution failed", err)
			return
		}
		instance.FlavorID = flavor.ID
		instance.Flavor = flavor
	}
	if instance.Hostname != hostname {
		instance.Hostname = hostname
	}
	if err = db.Model(instance).Updates(instance).Error; err != nil {
		log.Println("Failed to save instance", err)
		return
	}
	if string(action) != "" {
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/action_vm.sh '%d' '%s'", instance.ID, string(action))
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("action vm command execution failed", err)
			return
		}
	}
	return
}

func hyperExecute(ctx context.Context, control, command string) (err error) {
	return rpcs.HyperExecute(ctx, control, command)
}

func (a *InstanceAdmin) deleteInterfaces(ctx context.Context, instance *model.Instance) (err error) {
	for _, iface := range instance.Interfaces {
		err = a.deleteInterface(ctx, iface)
		if err != nil {
			log.Println("Failed to delete interface", err)
			continue
		}
	}
	return
}

func (a *InstanceAdmin) deleteInterface(ctx context.Context, iface *model.Interface) (err error) {
	err = DeleteInterface(ctx, iface)
	if err != nil {
		log.Println("Failed to create interface")
		return
	}
	vlan := iface.Address.Subnet.Vlan
	control := ""
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/del_host.sh '%d' '%s' '%s'", vlan, iface.MacAddr, iface.Address.Address)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete interface failed")
		return
	}
	return
}

func (a *InstanceAdmin) createInterface(ctx context.Context, subnet *model.Subnet, address, mac string, instance *model.Instance, ifname string, secGroups []*model.SecurityGroup, zoneID int64) (iface *model.Interface, err error) {
	memberShip := GetMemberShip(ctx)
	if subnet.Type == "public" {
		permit := memberShip.CheckPermission(model.Admin)
		if !permit {
			log.Println("Not authorized to create interface in public subnet")
			err = fmt.Errorf("Not authorized")
			return
		}
	}
	iface, err = CreateInterface(ctx, subnet, instance.ID, memberShip.OrgID, instance.Hyper, address, mac, ifname, "instance", secGroups)
	if err != nil {
		log.Println("Failed to create interface")
		return
	}
	return
}

func (a *InstanceAdmin) buildMetadata(ctx context.Context, primaryIface *InterfaceInfo, secondaryIfaces []*InterfaceInfo, keys []*model.Key, instance *model.Instance, userdata string, routerID, zoneID int64, service string) (interfaces []*model.Interface, metadata string, err error) {
	vlans := []*VlanInfo{}
	instNetworks := []*InstanceNetwork{}
	instLinks := []*NetworkLink{}
	primary := primaryIface.Subnet
	primaryIP := primaryIface.IpAddress
	primaryMac := primaryIface.MacAddress
	gateway := strings.Split(primary.Gateway, "/")[0]
	instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: gateway}
	iface, err := a.createInterface(ctx, primary, primaryIP, primaryMac, instance, "eth0", primaryIface.SecurityGroups, zoneID)
	if err != nil {
		log.Println("Allocate address for primary subnet %s--%s/%s failed, %v", primary.Name, primary.Network, primary.Netmask, err)
		return
	}
	interfaces = append(interfaces, iface)
	address := strings.Split(iface.Address.Address, "/")[0]
	instNetwork := &InstanceNetwork{Address: address, Netmask: primary.Netmask, Type: "ipv4", Link: iface.Name, ID: "network0"}
	instNetwork.Routes = append(instNetwork.Routes, instRoute)
	instNetworks = append(instNetworks, instNetwork)
	instLinks = append(instLinks, &NetworkLink{MacAddr: iface.MacAddr, Mtu: uint(iface.Mtu), ID: iface.Name, Type: "phy"})
	securityData, err := GetSecurityData(ctx, primaryIface.SecurityGroups)
	if err != nil {
		log.Println("Get security data for primary interface failed, %v", err)
		return
	}
	vlans = append(vlans, &VlanInfo{Device: "eth0", Vlan: primary.Vlan, Gateway: primary.Gateway, Router: primary.RouterID, IpAddr: address, MacAddr: iface.MacAddr, SecRules: securityData})
	for i, ifaceInfo := range secondaryIfaces {
		subnet := ifaceInfo.Subnet
		ifname := fmt.Sprintf("eth%d", i+1)
		iface, err = a.createInterface(ctx, subnet, ifaceInfo.IpAddress, ifaceInfo.MacAddress, instance, ifname, ifaceInfo.SecurityGroups, zoneID)
		if err != nil {
			log.Println("Allocate address for secondary subnet %s--%s/%s failed, %v", subnet.Name, subnet.Network, subnet.Netmask, err)
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
		securityData, err = GetSecurityData(ctx, ifaceInfo.SecurityGroups)
		if err != nil {
			log.Println("Get security data for secondary interface failed, %v", err)
			return
		}
		instLinks = append(instLinks, &NetworkLink{MacAddr: iface.MacAddr, Mtu: uint(iface.Mtu), ID: iface.Name, Type: "phy"})
		vlans = append(vlans, &VlanInfo{Device: ifname, Vlan: subnet.Vlan, Gateway: subnet.Gateway, Router: subnet.RouterID, IpAddr: address, MacAddr: iface.MacAddr, SecRules: securityData})
	}
	var instKeys []string
	for _, key := range keys {
		instKeys = append(instKeys, key.PublicKey)
	}
	image := &model.Image{Model: model.Model{ID: instance.ImageID}}
	err = DB().Take(image).Error
	if err != nil {
		log.Println("Invalid image ", instance.ImageID)
		return
	}
	virtType := image.VirtType
	dns := primary.NameServer
	if dns == primaryIP {
		dns = ""
	}
	instData := &InstanceData{
		Userdata: userdata,
		VirtType: virtType,
		DNS:      dns,
		Vlans:    vlans,
		Networks: instNetworks,
		Links:    instLinks,
		Keys:     instKeys,
	}
	jsonData, err := json.Marshal(instData)
	if err != nil {
		log.Println("Failed to marshal instance json data, %v", err)
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
		log.Println("Not authorized to delete the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	if instance.FloatingIps != nil {
		for _, fip := range instance.FloatingIps {
			fip.Instance = instance
			err = floatingIpAdmin.Detach(ctx, fip)
			if err != nil {
				log.Println("Failed to detach floating ip, %v", err)
				return
			}
		}
		instance.FloatingIps = nil
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.Volumes).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	var bootVolume *model.Volume
	if instance.Volumes != nil {
		for i, vol := range instance.Volumes {
			if vol.Booting {
				bootVolume = instance.Volumes[i]
				continue
			}
			_, err = volumeAdmin.Update(ctx, vol.ID, "", 0)
			if err != nil {
				log.Println("Failed to detach volume, %v", err)
				return
			}
		}
		instance.Volumes = nil
	}
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	if instance.Hyper == -1 {
		control = "toall="
	}
	bootName := ""
	if bootVolume != nil {
		bootName = fmt.Sprintf("%s-%d", bootVolume.Name, bootVolume.ID)
		err = volumeAdmin.Delete(ctx, bootVolume)
		if err != nil {
			log.Println("Failed to delete volume, %v", err)
			return
		}
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_vm.sh '%d' '%s' '%d'", instance.ID, bootName, instance.RouterID)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete vm command execution failed ", err)
		return
	}
	instance.Status = "deleting"
	err = db.Model(instance).Updates(instance).Error
	if err != nil {
		log.Println("Failed to mark vm as deleting ", err)
		return
	}
	return
}

func (a *InstanceAdmin) Get(ctx context.Context, id int64) (instance *model.Instance, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	instance = &model.Instance{Model: model.Model{ID: id}}
	if err = db.Preload("Image").Preload("Zone").Preload("Flavor").Preload("Keys").Where(where).Take(instance).Error; err != nil {
		log.Println("Failed to query instance, %v", err)
		return
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where("instance = ?", instance.ID).Find(&instance.Interfaces).Error; err != nil {
		log.Println("Failed to query interfaces %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, instance.Owner)
	if !permit {
		log.Println("Not authorized to read the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	permit = memberShip.CheckPermission(model.Admin)
	if permit {
		instance.OwnerInfo = &model.Organization{Model: model.Model{ID: instance.Owner}}
		if err = db.Take(instance.OwnerInfo).Error; err != nil {
			log.Println("Failed to query owner info", err)
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
	if err = db.Preload("Image").Preload("Zone").Preload("Flavor").Preload("Keys").Where(where).Where("uuid = ?", uuID).Take(instance).Error; err != nil {
		log.Println("Failed to query instance, %v", err)
		return
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where("instance = ?", instance.ID).Find(&instance.Interfaces).Error; err != nil {
		log.Println("Failed to query interfaces %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, instance.Owner)
	if !permit {
		log.Println("Not authorized to read the instance")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *InstanceAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, instances []*model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
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

	if query != "" {
		query = fmt.Sprintf("hostname like '%%%s%%'", query)
	}
	where := memberShip.GetWhere()
	instances = []*model.Instance{}
	if err = db.Model(&model.Instance{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Image").Preload("Zone").Preload("Flavor").Preload("Keys").Where(where).Where(query).Find(&instances).Error; err != nil {
		log.Println("Failed to query instance(s), %v", err)
		return
	}
	db = db.Offset(0).Limit(-1)
	for _, instance := range instances {
		if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where("instance = ?", instance.ID).Find(&instance.Interfaces).Error; err != nil {
			log.Println("Failed to query interfaces %v", err)
			return
		}
		if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
			log.Println("Failed to query floating ip(s), %v", err)
			return
		}
		permit := memberShip.CheckPermission(model.Admin)
		if permit {
			instance.OwnerInfo = &model.Organization{Model: model.Model{ID: instance.Owner}}
			if err = db.Take(instance.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
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
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
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
		log.Println("Not authorized for this operation")
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
		log.Println("Not authorized for this operation")
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
	log.Printf("Instance status %+v", instance)
	c.HTML(200, "instances_status")
}

func (v *InstanceView) Delete(c *macaron.Context, store session.Store) (err error) {
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
	err = instanceAdmin.Delete(ctx, instance)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
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
		log.Println("Not authorized for this operation")
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
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instanceID)}}
	if err = db.Preload("Interfaces").Take(instance).Error; err != nil {
		log.Println("Image query failed", err)
		return
	}
	if err = db.Where("instance_id = ?", instanceID).Find(&instance.FloatingIps).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
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
	if flag == "ChangeHostname" {
		c.HTML(200, "instances_hostname")
	} else if flag == "ChangeStatus" {
		if c.QueryTrim("action") != "" {
			instanceID64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Println("Change String to int64 failed", err)
				return
			}
			_, vmError := instanceAdmin.ChangeInstanceStatus(c.Req.Context(), instanceID64, c.QueryTrim("action"))
			if vmError != nil {
				log.Println("Launch vm command execution failed", err)
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
		log.Println("Invalid instance", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	var flavor *model.Flavor
	if flavorID > 0 {
		flavor, err = flavorAdmin.Get(ctx, flavorID)
		if err != nil {
			log.Println("Invalid flavor", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	err = instanceAdmin.Update(c.Req.Context(), instance, flavor, hostname, PowerAction(action), hyperID)
	if err != nil {
		log.Println("update instance failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}

func (v *InstanceView) checkNetparam(subnet *model.Subnet, IP, mac string) (macAddr string, err error) {
	_, inNet, err := net.ParseCIDR(subnet.Network)
	if err != nil {
		log.Println("CIDR parsing failed ", err)
		return
	}
	if IP != "" && !inNet.Contains(net.ParseIP(IP)) {
		log.Printf("Primary IP %s not belonging to subnet %v\n", IP, subnet)
		err = fmt.Errorf("Primary IP not belonging to subnet")
		return
	}
	if mac != "" {
		macl := strings.Split(mac, ":")
		if len(macl) != 6 {
			log.Println("Invalid mac address format")
			err = fmt.Errorf("Invalid mac address format")
			return
		}
		macAddr = strings.ToLower(mac)
		var tmp [6]int
		_, err = fmt.Sscanf(macAddr, "%02x:%02x:%02x:%02x:%02x:%02x", &tmp[0], &tmp[1], &tmp[2], &tmp[3], &tmp[4], &tmp[5])
		if err != nil {
			log.Println("Failed to parse mac address")
			return
		}
		if tmp[0]%2 == 1 {
			log.Println("Not a valid unicast mac address")
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
		log.Println("Need Write permissions")
		c.Data["ErrorMsg"] = "Need Write permissions"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../instances"
	hostname := c.QueryTrim("hostname")
	cnt := c.QueryTrim("count")
	count, err := strconv.Atoi(cnt)
	if err != nil {
		log.Println("Invalid instance count", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	hyperID := c.QueryInt("hyper")
	if hyperID >= 0 {
		permit := memberShip.CheckPermission(model.Admin)
		if !permit {
			log.Println("Need Admin permissions")
			c.Data["ErrorMsg"] = "Need Admin permissions"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	imageID := c.QueryInt64("image")
	if imageID <= 0 {
		log.Println("No valid image ID", imageID)
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
		log.Println("Invalid flavor ID", flavorID)
		c.Data["ErrorMsg"] = "Invalid flavor ID"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor, err := flavorAdmin.Get(ctx, flavorID)
	if err != nil {
		log.Println("No valid flavor", err)
		c.Data["ErrorMsg"] = "No valid flavor"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	zoneID := c.QueryInt64("zone")
	zone, err := zoneAdmin.Get(ctx, zoneID)
	if err != nil {
		log.Println("No valid zone", err)
		c.Data["ErrorMsg"] = "No valid zone"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	primary := c.QueryTrim("primary")
	primaryID, err := strconv.Atoi(primary)
	if err != nil {
		log.Println("Invalid primary subnet ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	primaryIP := c.QueryTrim("primaryip")
	ipAddr := strings.Split(primaryIP, "/")[0]
	primaryMac := c.QueryTrim("primarymac")
	primarySubnet, err := subnetAdmin.Get(ctx, int64(primaryID))
	if err != nil {
		log.Println("Get primary subnet failed", err)
		c.HTML(http.StatusBadRequest, err.Error())
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
				log.Println("Get security groups failed", err)
				c.Data["ErrorMsg"] = "Get security groups failed"
				c.HTML(http.StatusBadRequest, err.Error())
				return
			}
			if secgroup.RouterID != primarySubnet.RouterID {
				log.Println("Security group is not the same router with subnet")
				c.Data["ErrorMsg"] = "Security group is not in subnet vpc"
				c.HTML(http.StatusBadRequest, "Security group not in subnet vpc")
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
				log.Println("Get router failed", err)
				c.Data["ErrorMsg"] = "Get router failed"
				c.HTML(http.StatusBadRequest, err.Error())
				return
			}
			sgID = router.DefaultSG
		}
		var secGroup *model.SecurityGroup
		secGroup, err = secgroupAdmin.Get(ctx, int64(sgID))
		if err != nil {
			log.Println("Get security groups failed", err)
			c.Data["ErrorMsg"] = "Get security groups failed"
			c.HTML(http.StatusBadRequest, err.Error())
			return
		}
		securityGroups = append(securityGroups, secGroup)
	}
	primaryIface := &InterfaceInfo{
		Subnet:         primarySubnet,
		IpAddress:      ipAddr,
		MacAddress:     macAddr,
		SecurityGroups: securityGroups,
	}
	subnets := c.QueryTrim("subnets")
	var secondaryIfaces []*InterfaceInfo
	s := strings.Split(subnets, ",")
	for i := 0; i < len(s); i++ {
		sID, err := strconv.Atoi(s[i])
		if err != nil {
			log.Println("Invalid secondary subnet ID", err)
			continue
		}
		var subnet *model.Subnet
		subnet, err = subnetAdmin.Get(ctx, int64(sID))
		if err != nil {
			log.Println("Get secondary subnet failed", err)
			c.HTML(http.StatusBadRequest, err.Error())
			return
		}
		if subnet.RouterID != primarySubnet.RouterID {
			log.Println("All subnets must be in the same vpc", err)
			c.Data["ErrorMsg"] = "All subnets must be in the same vpc"
			c.HTML(http.StatusBadRequest, "All subnets must be in the same vpc")
			return
		}
		secondaryIfaces = append(secondaryIfaces, &InterfaceInfo{
			Subnet:         subnet,
			IpAddress:      "",
			MacAddress:     "",
			SecurityGroups: securityGroups,
		})
	}
	keys := c.QueryTrim("keys")
	k := strings.Split(keys, ",")
	var instKeys []*model.Key
	for i := 0; i < len(k); i++ {
		kID, err := strconv.Atoi(k[i])
		if err != nil {
			log.Println("Invalid key ID", err)
			continue
		}
		var key *model.Key
		key, err = keyAdmin.Get(ctx, int64(kID))
		if err != nil {
			log.Println("Failed to access key", err)
			c.Data["ErrorMsg"] = "Failed to access key"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		instKeys = append(instKeys, key)
	}
	userdata := c.QueryTrim("userdata")
	_, err = instanceAdmin.Create(ctx, count, hostname, userdata, image, flavor, zone, primarySubnet.RouterID, primaryIface, secondaryIfaces, instKeys, hyperID)
	if err != nil {
		log.Println("Create instance failed", err)
		c.HTML(http.StatusBadRequest, err.Error())
		return
	}
	c.Redirect(redirectTo)
}
