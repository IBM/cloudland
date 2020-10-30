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
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/cloudland/web/clui/grpcs"
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	instanceAdmin = &InstanceAdmin{}
	instanceView  = &InstanceView{}
)

type InstanceAdmin struct{}

type InstanceView struct{}

type NetworkRoute struct {
	Network string `json:"network"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
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

type VlanInfo struct {
	Device  string `json:"device"`
	Vlan    int64  `json:"vlan"`
	IpAddr  string `json:"ip_address"`
	MacAddr string `json:"mac_address"`
}

type SecurityData struct {
	Secgroup    int64
	RemoteIp    string `json:"remote_ip"`
	RemoteGroup string `json:"remote_group"`
	Direction   string `json:"direction"`
	IpVersion   string `json:"ip_version"`
	Protocol    string `json:"protocol"`
	PortMin     int32  `json:"port_min"`
	PortMax     int32  `json:"port_max"`
}

type InstanceData struct {
	Userdata string             `json:"userdata"`
	Vlans    []*VlanInfo        `json:"vlans"`
	Networks []*InstanceNetwork `json:"networks"`
	Links    []*NetworkLink     `json:"links"`
	Keys     []string           `json:"keys"`
	SecRules []*SecurityData    `json:"security"`
}

type InstancesData struct {
	Instances []*model.Instance  `json:"instancedata"`
}

func (a *InstanceAdmin) Create(ctx context.Context, count int, prefix, userdata string, imageID, flavorID, primaryID, clusterID int64, primaryIP, primaryMac string, subnetIDs, keyIDs []int64, sgIDs []int64, hyper int) (instance *model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	image := &model.Image{Model: model.Model{ID: imageID}}
	if imageID > 0 {
		if err = db.Take(image).Error; err != nil {
			log.Println("Image query failed", err)
			return
		}
		if image.Status != "available" {
			err = fmt.Errorf("Image status not available")
			log.Println("Image status not available")
			return
		}
	}
	flavor := &model.Flavor{Model: model.Model{ID: flavorID}}
	if err = db.Find(flavor).Error; err != nil {
		log.Println("Flavor query failed", err)
		return
	}
	primary := &model.Subnet{Model: model.Model{ID: primaryID}}
	if err = db.Preload("Netlink").Take(primary).Error; err != nil {
		log.Println("Primary subnet query failed", err)
		return
	}
	subnets := []*model.Subnet{}
	if err = db.Where(subnetIDs).Preload("Netlink").Find(&subnets).Error; err != nil {
		log.Println("Secondary subnets query failed", err)
		return
	}
	keys := []*model.Key{}
	if err = db.Where(keyIDs).Find(&keys).Error; err != nil {
		log.Println("Keys query failed", err)
		return
	}
	secGroups := []*model.SecurityGroup{}
	if err = db.Where(sgIDs).Find(&secGroups).Error; err != nil {
		log.Println("Security group query failed", err)
		return
	}
	i := 0
	hostname := prefix
	for i < count {
		if count > 1 {
			hostname = fmt.Sprintf("%s-%d", prefix, i+1)
		}
		instance = &model.Instance{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Hostname: hostname, ImageID: imageID, FlavorID: flavorID, Userdata: userdata, Status: "pending", ClusterID: clusterID}
		err = db.Create(instance).Error
		if err != nil {
			log.Println("DB create instance failed", err)
			return
		}
		metadata := ""
		_, metadata, err = a.buildMetadata(ctx, primary, primaryIP, primaryMac, subnets, keys, instance, userdata, secGroups)
		if err != nil {
			log.Println("Build instance metadata failed", err)
			return
		}
		rcNeeded := fmt.Sprintf("cpu=%d memory=%d disk=%d network=%d", flavor.Cpu, flavor.Memory*1024, (flavor.Disk+flavor.Swap+flavor.Ephemeral)*1024*1024, 0)
		control := "inter= " + rcNeeded
		if i == 0 && hyper >= 0 {
			control = fmt.Sprintf("inter=%d %s", hyper, rcNeeded)
		}
		if primary.DomainSearch != "" {
			hostname = hostname + "." + primary.DomainSearch
		}
		command := ""
		if imageID > 0 {
			command = fmt.Sprintf("/opt/cloudland/scripts/backend/launch_vm.sh '%d' 'image-%d.%s' '%s' '%d' '%d' '%d' '%d' '%d'<<EOF\n%s\nEOF", instance.ID, image.ID, image.Format, hostname, flavor.Cpu, flavor.Memory, flavor.Disk, flavor.Swap, flavor.Ephemeral, base64.StdEncoding.EncodeToString([]byte(metadata)))
		} else if clusterID > 0 {
			command = fmt.Sprintf("/opt/cloudland/scripts/backend/oc_vm.sh '%d' '%d' '%d' '%d' '%s'<<EOF\n%s\nEOF", instance.ID, flavor.Cpu, flavor.Memory, flavor.Disk, hostname, metadata)
			openshift := &model.Openshift{Model: model.Model{ID: clusterID}}
			err = db.Model(openshift).Update("worker_num", gorm.Expr("worker_num + 1")).Error
			if err != nil {
				log.Println("Failed to update openshift cluster")
				return
			}
		}
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed", err)
			return
		}
		i++
	}
	return
}

func (a *InstanceAdmin) Update(ctx context.Context, id, flavorID int64, hostname, action string, subnetIDs, sgIDs []int64, hyper int) (instance *model.Instance, err error) {
	db := DB()
	instance = &model.Instance{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(instance).Error; err != nil {
		log.Println("Failed to query instance ", err)
		return
	}
	if hyper != int(instance.Hyper) {
		if instance.Status != "shut_off" {
			log.Println("Instance must be shutdown before migration")
			err = fmt.Errorf("Instance must be shutdown before migration")
			return
		}
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/migrate_vm.sh '%d' '%d'", instance.ID, hyper)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Migrate vm command execution failed", err)
			return
		}
	}
	if flavorID != instance.FlavorID {
		if instance.Status == "running" {
			err = fmt.Errorf("Instance must be shutdown first before resize")
			log.Println("Instance must be shutdown first before resize", err)
			return
		}
		flavor := &model.Flavor{Model: model.Model{ID: flavorID}}
		if err = db.Take(flavor).Error; err != nil {
			log.Println("Failed to query flavor", err)
			return
		}
		if flavor.Disk < instance.Flavor.Disk || flavor.Ephemeral < instance.Flavor.Ephemeral {
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
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/resize_vm.sh '%d' '%d' '%d' '%d' '%d' '%d'", instance.ID, flavor.Cpu, flavor.Memory, flavor.Disk, flavor.Swap, flavor.Ephemeral)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Resize vm command execution failed", err)
			return
		}
		instance.FlavorID = flavorID
		instance.Flavor = flavor
		if err = db.Save(instance).Error; err != nil {
			log.Println("Failed to save instance", err)
			return
		}
	}
	if instance.Hostname != hostname {
		instance.Hostname = hostname
		if err = db.Save(instance).Error; err != nil {
			log.Println("Failed to save instance", err)
			return
		}
	}
	if action == "shutdown" || action == "destroy" || action == "start" || action == "suspend" || action == "resume" {
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/action_vm.sh '%d' '%s'", instance.ID, action)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete vm command execution failed", err)
			return
		}
	}
	secGroups := []*model.SecurityGroup{}
	if err = db.Where(sgIDs).Find(&secGroups).Error; err != nil {
		log.Println("Security group query failed", err)
		return
	}
	secRules, err := model.GetSecurityRules(secGroups)
	if err != nil {
		log.Println("Failed to get security rules", err)
		return
	}
	securityData := []*SecurityData{}
	for _, rule := range secRules {
		sgr := &SecurityData{
			Secgroup:    rule.Secgroup,
			RemoteIp:    rule.RemoteIp,
			RemoteGroup: rule.RemoteGroup,
			Direction:   rule.Direction,
			IpVersion:   rule.IpVersion,
			Protocol:    rule.Protocol,
			PortMin:     rule.PortMin,
			PortMax:     rule.PortMax,
		}
		securityData = append(securityData, sgr)
	}
	jsonData, err := json.Marshal(securityData)
	if err != nil {
		log.Println("Failed to marshal security json data, %v", err)
		return
	}
	for _, iface := range instance.Interfaces {
		found := false
		for _, sID := range subnetIDs {
			if iface.Address.SubnetID == sID {
				found = true
				log.Println("Found SID ", sID)
				break
			}
		}
		if found == false {
			control := fmt.Sprintf("inter=%d", instance.Hyper)
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/detach_nic.sh '%d' '%d' '%s' '%s'", instance.ID, iface.Address.Subnet.Vlan, iface.Address.Address, iface.MacAddr)
			err = hyperExecute(ctx, control, command)
			if err != nil {
				log.Println("Delete vm command execution failed", err)
				return
			}
			err = a.deleteInterface(ctx, iface)
			if err != nil {
				log.Println("Failed to delete interface", err)
				return
			}
		}
	}
	index := len(instance.Interfaces)
	for i, sID := range subnetIDs {
		found := false
		for _, iface := range instance.Interfaces {
			if iface.Address.SubnetID == sID {
				found = true
				log.Println("Found SID ", sID)
				break
			}
		}
		if found == false {
			var iface *model.Interface
			ifname := fmt.Sprintf("eth%d", i+index)
			subnet := &model.Subnet{Model: model.Model{ID: sID}}
			err = db.Take(subnet).Error
			if err != nil {
				log.Println("Failed to query subnet", err)
				return
			}
			iface, err = a.createInterface(ctx, subnet, "", "", instance, ifname, secGroups)
			control := fmt.Sprintf("inter=%d", instance.Hyper)
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/attach_nic.sh '%d' '%d' '%s' '%s' <<EOF\n%s\nEOF", instance.ID, iface.Address.Subnet.Vlan, iface.Address.Address, iface.MacAddr, jsonData)
			err = hyperExecute(ctx, control, command)
			if err != nil {
				log.Println("Delete vm command execution failed", err)
				return
			}
		}
	}
	return
}

func hyperExecute(ctx context.Context, control, command string) (err error) {
	if control == "" {
		return
	}
	sciClient := grpcs.RemoteExecClient()
	sciReq := &scripts.ExecuteRequest{
		Id:      100,
		Extra:   0,
		Control: control,
		Command: command,
	}
	_, err = sciClient.Execute(ctx, sciReq)
	if err != nil {
		log.Println("SCI client execution failed, %v", err)
		return
	}
	return
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
	netlink := iface.Address.Subnet.Netlink
	if netlink == nil {
		log.Println("Subnet doesn't have network")
		return
	}
	control := ""
	if netlink.Hyper >= 0 {
		control = fmt.Sprintf("toall=vlan-%d:%d", iface.Address.Subnet.Vlan, netlink.Hyper)
		if netlink.Peer >= 0 {
			control = fmt.Sprintf("%s,%d", control, netlink.Peer)
		}
	} else if netlink.Peer >= 0 {
		control = fmt.Sprintf("inter=%d", netlink.Peer)
	} else {
		log.Println("Network has no valid hypers")
		return
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/del_host.sh '%d' '%s' '%s'", vlan, iface.MacAddr, iface.Address.Address)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete interface failed")
		return
	}
	return
}

func (a *InstanceAdmin) createInterface(ctx context.Context, subnet *model.Subnet, address, mac string, instance *model.Instance, ifname string, secGroups []*model.SecurityGroup) (iface *model.Interface, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	iface, err = CreateInterface(ctx, subnet.ID, instance.ID, memberShip.OrgID, instance.Hyper, address, mac, ifname, "instance", secGroups)
	if err != nil {
		log.Println("Failed to create interface")
		return
	}
	netlink := subnet.Netlink
	if netlink == nil {
		netlink = &model.Network{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Vlan: subnet.Vlan}
		err = db.Create(netlink).Error
		if err != nil {
			log.Println("Failed to query network")
			return
		}
	}
	if err != nil {
		log.Println("Failed to execute network creation")
		return
	}
	control := ""
	if netlink.Hyper >= 0 {
		control = fmt.Sprintf("toall=vlan-%d:%d", subnet.Vlan, netlink.Hyper)
		if netlink.Peer >= 0 {
			control = fmt.Sprintf("%s,%d", control, netlink.Peer)
		}
	} else if netlink.Peer >= 0 {
		control = fmt.Sprintf("inter=%d", netlink.Peer)
	} else {
		log.Println("Network has no valid hypers")
		return
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_host.sh '%d' '%s' '%s' '%s' '%s'", subnet.Vlan, iface.MacAddr, instance.Hostname, iface.Address.Address, subnet.DomainSearch)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete slave failed")
	}
	return
}

func (a *InstanceAdmin) buildMetadata(ctx context.Context, primary *model.Subnet, primaryIP, primaryMac string, subnets []*model.Subnet, keys []*model.Key, instance *model.Instance, userdata string, secGroups []*model.SecurityGroup) (interfaces []*model.Interface, metadata string, err error) {
	vlans := []*VlanInfo{}
	instNetworks := []*InstanceNetwork{}
	instLinks := []*NetworkLink{}
	gateway := strings.Split(primary.Gateway, "/")[0]
	instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: gateway}
	iface, err := a.createInterface(ctx, primary, primaryIP, primaryMac, instance, "eth0", secGroups)
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
	vlans = append(vlans, &VlanInfo{Device: "eth0", Vlan: primary.Vlan, IpAddr: address, MacAddr: iface.MacAddr})
	for i, subnet := range subnets {
		ifname := fmt.Sprintf("eth%d", i+1)
		iface, err = a.createInterface(ctx, subnet, "", "", instance, ifname, secGroups)
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
		instLinks = append(instLinks, &NetworkLink{MacAddr: iface.MacAddr, Mtu: uint(iface.Mtu), ID: iface.Name, Type: "phy"})
		vlans = append(vlans, &VlanInfo{Device: ifname, Vlan: subnet.Vlan, IpAddr: address, MacAddr: iface.MacAddr})
	}
	var instKeys []string
	for _, key := range keys {
		instKeys = append(instKeys, key.PublicKey)
	}
	secRules, err := model.GetSecurityRules(secGroups)
	if err != nil {
		log.Println("Failed to get security rules", err)
		return
	}
	securityData := []*SecurityData{}
	for _, rule := range secRules {
		sgr := &SecurityData{
			Secgroup:    rule.Secgroup,
			RemoteIp:    rule.RemoteIp,
			RemoteGroup: rule.RemoteGroup,
			Direction:   rule.Direction,
			IpVersion:   rule.IpVersion,
			Protocol:    rule.Protocol,
			PortMin:     rule.PortMin,
			PortMax:     rule.PortMax,
		}
		securityData = append(securityData, sgr)
	}
	instData := &InstanceData{
		Userdata: userdata,
		Vlans:    vlans,
		Networks: instNetworks,
		Links:    instLinks,
		Keys:     instKeys,
		SecRules: securityData,
	}
	jsonData, err := json.Marshal(instData)
	if err != nil {
		log.Println("Failed to marshal instance json data, %v", err)
		return
	}
	return interfaces, string(jsonData), nil
}

func (a *InstanceAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	instance := &model.Instance{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(instance).Error; err != nil {
		log.Println("Failed to query instance, %v", err)
		return
	}
	if instance.ClusterID > 0 && strings.Index(instance.Hostname, "worker-") == 0 {
		openshift := &model.Openshift{Model: model.Model{ID: instance.ClusterID}}
		err = db.Model(openshift).Update("worker_num", gorm.Expr("worker_num - 1")).Error
		if err != nil {
			log.Println("Failed to update openshift cluster")
			return
		}
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	if instance.FloatingIps != nil {
		for _, fip := range instance.FloatingIps {
			err = floatingipAdmin.Delete(ctx, fip.ID)
			if err != nil {
				log.Println("Failed to delete floating ip, %v", err)
				return
			}
		}
	}
	if err = db.Where("instance_id = ?", instance.ID).Find(&instance.Volumes).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	if instance.Volumes != nil {
		for _, vol := range instance.Volumes {
			_, err = volumeAdmin.Update(ctx, vol.ID, "", 0)
			if err != nil {
				log.Println("Failed to delete floating ip, %v", err)
				return
			}
		}
	}
	if instance.Hyper != -1 {
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_vm.sh '%d'", instance.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete vm command execution failed, %v", err)
			return
		}
	}
	if err = a.deleteInterfaces(ctx, instance); err != nil {
		log.Println("DB failed to delete interfaces, %v", err)
		return
	}
	if err = db.Delete(&model.Instance{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("Failed to delete instance, %v", err)
		return
	}
	return
}

func (a *InstanceAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, instances []*model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
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
	if err = db.Model(&model.Instance{}).Where(where).Where(query).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Where(where).Where(query).Where(query).Find(&instances).Error; err != nil {
		log.Println("Failed to query instance(s), %v", err)
		return
	}
	db = db.Offset(0).Limit(-1)
	for _, instance := range instances {
		if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
			log.Println("Failed to query floating ip(s), %v", err)
			return
		}
		if instance.ClusterID > 0 {
			instance.Cluster = &model.Openshift{Model: model.Model{ID: instance.ClusterID}}
			if err = db.Take(instance.Cluster).Error; err != nil {
				log.Println("Failed to query openshift cluster info", err)
				instance.ClusterID = 0
			}
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

func (a *InstanceAdmin) enableVnc(ctx context.Context, instance *model.Instance) (vnc *model.Vnc, err error) {
	db := DB()
	vnc = &model.Vnc{InstanceID: int64(instance.ID)}
	err = db.Where(vnc).Take(vnc).Error
	if err != nil {
		log.Println("VNC query failed", err)
	}
	expired := true
	if vnc.ExpiredAt != nil {
		expired = !vnc.ExpiredAt.After(time.Now())
	}
	if vnc.AccessAddress != "" && vnc.AccessPort > 0 && !expired {
		log.Println("Vnc uri is still valid")
		return
	}
	gateway := &model.Gateway{}
	routerID := instance.Interfaces[0].Address.Subnet.Router
	if routerID > 0 {
		gateway.ID = routerID
		err = db.Preload("Interfaces", "type = 'gateway_public'").Preload("Interfaces.Address").Take(gateway).Error
		if err != nil || gateway.Interfaces == nil || len(gateway.Interfaces) == 0 {
			log.Println("Failed to query instance gateway", err)
			return
		}
	} else {
		err = db.Preload("Interfaces", "type = 'gateway_public'").Preload("Interfaces.Address").Where("type = ?", "system").Take(gateway).Error
		if err != nil && gorm.IsRecordNotFoundError(err) {
			log.Println("Creating new system router")
			_, err = gatewayAdmin.Create(ctx, "System-Router", "system", 0, 0, nil, 1)
			if err != nil {
				log.Println("Failed to create system router", err)
			}
			return
		}
	}
	if vnc.LocalPort == 0 || expired {
		vnc.Passwd = ""
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/replace_vnc_passwd.sh '%d'", instance.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Replace vnc password command execution failed", err)
		}
	}
	if vnc.LocalPort != 0 && vnc.LocalAddress != "" && (vnc.AccessAddress == "" || vnc.AccessPort == 0 || expired) {
		raddress := strings.Split(gateway.Interfaces[0].Address.Address, "/")[0]
		count := 1
		rport := 0
		for count > 0 {
			rport = rand.Intn(remoteMax-remoteMin) + remoteMin
			if err = db.Model(&model.Vnc{}).Where("access_port = ? and router = ?", rport, routerID).Count(&count).Error; err != nil {
				log.Println("Failed to query existing remote port", err)
				return
			}
		}
		control := fmt.Sprintf("toall=router-%d:%d,%d", gateway.ID, gateway.Hyper, gateway.Peer)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/enable_vnc_portmap.sh '%d' '%d' '%s' '%d' '%s' '%d'", instance.ID, gateway.ID, vnc.LocalAddress, vnc.LocalPort, raddress, rport)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Enable vnc portmap command execution failed", err)
		}
		vnc.Router = routerID
		err = db.Where("instance_id = ?", instance.ID).Assign(vnc).FirstOrCreate(&model.Vnc{}).Error
		if err != nil {
			log.Println("Failed to update vnc", err)
		}
		vnc.AccessAddress = raddress
		vnc.AccessPort = int32(rport)
		expireAt := time.Now().Add(time.Minute * 30).UTC()
		vnc.ExpiredAt = &expireAt
	}
	return
}

func (v *InstanceView) List(c *macaron.Context, store session.Store) {
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
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
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
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"instances": instances,
			"total":     total,
			"pages":     pages,
			"query":     query,
		})
		return
	}
	c.HTML(200, "instances")
}

func (v *InstanceView) UpdateTable(c *macaron.Context, store session.Store){
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
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	var jsonData *InstancesData
	jsonData = &InstancesData{
		Instances: instances,
	}
	
	c.JSON(200, jsonData)
	return 
}

func (v *InstanceView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
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
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = instanceAdmin.Delete(c.Req.Context(), int64(instanceID))
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
	ctx := c.Req.Context()
	_, subnets, err := subnetAdmin.List(ctx, 0, -1, "", "", "")
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
	_, openshifts, err := openshiftAdmin.List(ctx, 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Images"] = images
	c.Data["Flavors"] = flavors
	c.Data["Subnets"] = subnets
	c.Data["Openshifts"] = openshifts
	c.Data["Secgroups"] = secgroups
	c.Data["Keys"] = keys
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
	if err = db.Set("gorm:auto_preload", true).Take(instance).Error; err != nil {
		log.Println("Image query failed", err)
		return
	}
	if err = db.Where("instance_id = ?", instanceID).Find(&instance.FloatingIps).Error; err != nil {
		log.Println("Failed to query floating ip(s), %v", err)
		return
	}
	_, subnets, err := subnetAdmin.List(c.Req.Context(), 0, -1, "", "", "")
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
	if flag == "ChangeHostname"{
		c.HTML(200, "instances_hostname")
	}else if flag == "ChangeStatus"{
		c.HTML(200, "instances_status")
	}else if flag == "MigrateInstance"{
		c.HTML(200, "instances_migrate")
	}else if flag == "ResizeInstance"{
		c.HTML(200, "instances_size")
	}else{
		c.HTML(200, "instances_patch")
	}
}

func (v *InstanceView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../instances"
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Writer, "instances", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor := c.QueryInt64("flavor")
	hostname := c.QueryTrim("hostname")                                             
	hyperID := c.QueryInt("hyper")
	action := c.QueryTrim("action")
	ifaces := c.QueryStrings("ifaces")
	instance := &model.Instance{Model: model.Model{ID: id}}
	err = DB().Take(instance).Error
	if err != nil {
		log.Println("Invalid instance", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if hyperID != int(instance.Hyper) {
		permit, err = memberShip.CheckAdmin(model.Admin, "instances", id)
		if !permit {
			log.Println("Not authorized to migrate VM")
			err = fmt.Errorf("Not authorized to migrate VM")
			return
		}
	}
	hyper := &model.Hyper{Hostid: int32(hyperID)}
	err = DB().Take(hyper).Error
	if err != nil {
		log.Println("Invalid hypervisor", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	var subnetIDs []int64
	for _, s := range ifaces {
		sID, err := strconv.Atoi(s)
		if err != nil {
			log.Println("Invalid secondary subnet ID, %v", err)
			continue
		}
		permit, err = memberShip.CheckOwner(model.Writer, "subnets", int64(sID))
		if !permit {
			log.Println("Not authorized for this operation")
			c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
			return
		}
		subnetIDs = append(subnetIDs, int64(sID))
	}
	var sgIDs []int64
	sgIDs = append(sgIDs, store.Get("defsg").(int64))
	instance, err = instanceAdmin.Update(c.Req.Context(), id, flavor, hostname, action, subnetIDs, sgIDs, hyperID)
	if err != nil {
		log.Println("Create instance failed, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, instance)
		return
	}
	c.Redirect(redirectTo)
}

func (v *InstanceView) checkNetparam(subnetID int64, IP, mac string) (macAddr string, err error) {
	subnet := &model.Subnet{Model: model.Model{ID: subnetID}}
	err = DB().Take(subnet).Error
	if err != nil {
		log.Println("DB failed to query subnet ", err)
		return
	}
	inNet := &net.IPNet{
		IP:   net.ParseIP(subnet.Network),
		Mask: net.IPMask(net.ParseIP(subnet.Netmask).To4()),
	}
	if IP != "" && !inNet.Contains(net.ParseIP(IP)) {
		log.Println("Primary IP not belonging to subnet")
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
	hyper := &model.Hyper{Hostid: int32(hyperID)}
	err = DB().Take(hyper).Error
	if err != nil {
		log.Println("Invalid hypervisor", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	cluster := c.QueryInt64("cluster")
	if cluster < 0 {
		log.Println("Invalid cluster ID", err)
		c.Data["ErrorMsg"] = "Invalid cluster ID"
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if cluster > 0 {
		permit, err = memberShip.CheckAdmin(model.Writer, "openshifts", cluster)
		if !permit {
			log.Println("Not authorized to access openshift cluster")
			c.Data["ErrorMsg"] = "Not authorized to access openshift cluster"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	image := c.QueryInt64("image")
	if image <= 0 && cluster <= 0 {
		log.Println("No valid image ID or cluster ID", err)
		c.Data["ErrorMsg"] = "No valid image ID or cluster ID"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor := c.QueryInt64("flavor")
	if flavor <= 0 {
		log.Println("Invalid flavor ID", err)
		c.Data["ErrorMsg"] = "Invalid flavor ID"
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
	permit, err = memberShip.CheckAdmin(model.Writer, "subnets", int64(primaryID))
	if !permit {
		log.Println("Not authorized to access subnet")
		c.Data["ErrorMsg"] = "Need Write permissions"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	primaryIP := c.QueryTrim("primaryip")
	ipAddr := strings.Split(primaryIP, "/")[0]
	primaryMac := c.QueryTrim("primarymac")
	macAddr, err := v.checkNetparam(int64(primaryID), ipAddr, primaryMac)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	subnets := c.QueryTrim("subnets")
	s := strings.Split(subnets, ",")
	var subnetIDs []int64
	for i := 0; i < len(s); i++ {
		sID, err := strconv.Atoi(s[i])
		if err != nil {
			log.Println("Invalid secondary subnet ID, %v", err)
			continue
		}
		permit, err = memberShip.CheckAdmin(model.Writer, "subnets", int64(sID))
		if !permit {
			log.Println("Not authorized to access subnet")
			c.Data["ErrorMsg"] = "Not authorized to access subnet"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		subnetIDs = append(subnetIDs, int64(sID))
	}
	keys := c.QueryTrim("keys")
	k := strings.Split(keys, ",")
	var keyIDs []int64
	for i := 0; i < len(k); i++ {
		kID, err := strconv.Atoi(k[i])
		if err != nil {
			log.Println("Invalid key ID, %v", err)
			continue
		}
		permit, err = memberShip.CheckOwner(model.Writer, "keys", int64(kID))
		if !permit {
			log.Println("Not authorized to access key")
			c.Data["ErrorMsg"] = "Not authorized to access key"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		keyIDs = append(keyIDs, int64(kID))
	}
	secgroups := c.QueryTrim("secgroups")
	var sgIDs []int64
	if secgroups != "" {
		sg := strings.Split(secgroups, ",")
		for i := 0; i < len(sg); i++ {
			sgID, err := strconv.Atoi(sg[i])
			if err != nil {
				log.Println("Invalid security group ID", err)
				continue
			}
			permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
			if !permit {
				log.Println("Not authorized to access security group")
				c.Data["ErrorMsg"] = "Not authorized to access security group"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			sgIDs = append(sgIDs, int64(sgID))
		}
	} else {
		sgID := store.Get("defsg").(int64)
		permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
		if !permit {
			log.Println("Not authorized to access security group")
			c.Data["ErrorMsg"] = "Not authorized to access security group"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		sgIDs = append(sgIDs, sgID)
	}
	userdata := c.QueryTrim("userdata")
	instance, err := instanceAdmin.Create(c.Req.Context(), count, hostname, userdata, image, flavor, int64(primaryID), cluster, ipAddr, macAddr, subnetIDs, keyIDs, sgIDs, hyperID)
	if err != nil {
		log.Println("Create instance failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, err.Error())
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, instance)
		return
	}
	c.Redirect(redirectTo)
}
