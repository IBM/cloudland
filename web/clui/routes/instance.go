/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/clui/grpcs"
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
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

func (a *InstanceAdmin) Create(ctx context.Context, count int, prefix, userdata string, imageID, flavorID, primaryID int64, primaryIP string, subnetIDs, keyIDs []int64, sgIDs []int64, hyper int) (instance *model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	image := &model.Image{Model: model.Model{ID: imageID}}
	if err = db.Take(image).Error; err != nil {
		log.Println("Image query failed", err)
		return
	}
	if image.Status != "available" {
		err = fmt.Errorf("Image status not available")
		log.Println("Image status not available")
		return
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
		instance = &model.Instance{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Hostname: hostname, ImageID: imageID, FlavorID: flavorID, Userdata: userdata, Status: "pending"}
		err = db.Create(instance).Error
		if err != nil {
			log.Println("DB create instance failed", err)
			return
		}
		metadata := ""
		_, metadata, err = a.buildMetadata(ctx, primary, primaryIP, subnets, keys, instance, userdata, secGroups)
		if err != nil {
			log.Println("Build instance metadata failed", err)
			return
		}
		control := fmt.Sprintf("inter= cpu=%d memory=%d disk=%d network=%d", flavor.Cpu, flavor.Memory*1024, flavor.Disk*1024*1024, 0)
		if i == 0 && hyper >= 0 {
			control = fmt.Sprintf("inter=%d cpu=%d memory=%d disk=%d network=%d", hyper, 0, 0, 0, 0)
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/launch_vm.sh %d image-%d.%s %s %d %d %d <<EOF\n%s\nEOF", instance.ID, image.ID, image.Format, hostname, flavor.Cpu, flavor.Memory, flavor.Disk, metadata)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed", err)
			return
		}
		i++
	}
	return
}

func (a *InstanceAdmin) Update(ctx context.Context, id int64, hostname, action string, subnetIDs, sgIDs []int64) (instance *model.Instance, err error) {
	db := DB()
	instance = &model.Instance{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(instance).Error; err != nil {
		log.Println("Failed to query instance ", err)
		return
	}
	if instance.Hostname != hostname {
		instance.Hostname = hostname
		if err = db.Save(instance).Error; err != nil {
			log.Println("Failed to save instance", err)
			return
		}
	}
	if action == "shutdown" || action == "start" || action == "suspend" || action == "resume" {
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/action_vm.sh %d %s", instance.ID, action)
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
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/detach_nic.sh %d %d %s %s", instance.ID, iface.Address.Subnet.Vlan, iface.Address.Address, iface.MacAddr)
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
			iface, err = a.createInterface(ctx, subnet, "", instance, ifname, secGroups)
			control := fmt.Sprintf("inter=%d", instance.Hyper)
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/attach_nic.sh %d %d %s %s <<EOF\n%s\nEOF", instance.ID, iface.Address.Subnet.Vlan, iface.Address.Address, iface.MacAddr, jsonData)
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
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/del_host.sh %d %s %s", vlan, iface.MacAddr, iface.Address.Address)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete interface failed")
		return
	}
	return
}

func (a *InstanceAdmin) createInterface(ctx context.Context, subnet *model.Subnet, address string, instance *model.Instance, ifname string, secGroups []*model.SecurityGroup) (iface *model.Interface, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	iface, err = CreateInterface(ctx, subnet.ID, instance.ID, memberShip.OrgID, address, ifname, "instance", secGroups)
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
	err = execNetwork(ctx, netlink, subnet, memberShip.OrgID)
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
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_host.sh %d %s %s %s", subnet.Vlan, iface.MacAddr, instance.Hostname, iface.Address.Address)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete slave failed")
	}
	return
}

func (a *InstanceAdmin) buildMetadata(ctx context.Context, primary *model.Subnet, primaryIP string, subnets []*model.Subnet, keys []*model.Key, instance *model.Instance, userdata string, secGroups []*model.SecurityGroup) (interfaces []*model.Interface, metadata string, err error) {
	vlans := []*VlanInfo{}
	instNetworks := []*InstanceNetwork{}
	instLinks := []*NetworkLink{}
	gateway := strings.Split(primary.Gateway, "/")[0]
	instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: gateway}
	iface, err := a.createInterface(ctx, primary, primaryIP, instance, "eth0", secGroups)
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
		iface, err = a.createInterface(ctx, subnet, "", instance, ifname, secGroups)
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
	if err = db.Set("gorm:auto_preload", true).Find(instance).Error; err != nil {
		log.Println("Failed to query instance, %v", err)
		return
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
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_vm.sh %d", instance.ID)
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

func (a *InstanceAdmin) List(ctx context.Context, offset, limit int64, order string) (total int64, instances []*model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	where := memberShip.GetWhere()
	instances = []*model.Instance{}
	if err = db.Model(&model.Instance{}).Where(where).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Where(where).Find(&instances).Error; err != nil {
		log.Println("Failed to query instance(s), %v", err)
		return
	}
	for _, instance := range instances {
		if err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error; err != nil {
			log.Println("Failed to query floating ip(s), %v", err)
			return
		}
	}

	return
}

func (v *InstanceView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	total, instances, err := instanceAdmin.List(c.Req.Context(), offset, limit, order)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Instances"] = instances
	c.Data["Total"] = total
	c.HTML(200, "instances")
}

func (v *InstanceView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = instanceAdmin.Delete(c.Req.Context(), int64(instanceID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
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
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
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
	_, subnets, err := subnetAdmin.List(ctx, 0, 0, "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, secgroups, err := secgroupAdmin.List(ctx, 0, 0, "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, keys, err := keyAdmin.List(ctx, 0, 0, "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Images"] = images
	c.Data["Flavors"] = flavors
	c.Data["Subnets"] = subnets
	c.Data["Secgroups"] = secgroups
	c.Data["Keys"] = keys
	c.HTML(200, "instances_new")
}

func (v *InstanceView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
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
	subnets := []*model.Subnet{}
	where := ""
	for i, iface := range instance.Interfaces {
		if i == 0 {
			where = fmt.Sprintf("id != %d", iface.Address.Subnet.ID)
		} else {
			where = fmt.Sprintf("%s and id != %d", where, iface.Address.Subnet.ID)
		}
	}
	if err := db.Where(where).Find(&subnets).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Instance"] = instance
	c.Data["Subnets"] = subnets
	c.HTML(200, "instances_patch")
}

func (v *InstanceView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../instances"
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	hostname := c.QueryTrim("hostname")
	action := c.QueryTrim("action")
	ifaces := c.QueryStrings("ifaces")
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
			code := http.StatusUnauthorized
			c.Error(code, http.StatusText(code))
			return
		}
		subnetIDs = append(subnetIDs, int64(sID))
	}
	var sgIDs []int64
	sgIDs = append(sgIDs, store.Get("defsg").(int64))
	_, err = instanceAdmin.Update(c.Req.Context(), int64(instanceID), hostname, action, subnetIDs, sgIDs)
	if err != nil {
		log.Println("Create instance failed, %v", err)
		c.HTML(http.StatusBadRequest, err.Error())
	}
	c.Redirect(redirectTo)
}

func (v *InstanceView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Need Write permissions")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../instances"
	hostname := c.QueryTrim("hostname")
	cnt := c.QueryTrim("count")
	count, err := strconv.Atoi(cnt)
	if err != nil {
		log.Println("Invalid instance count", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	hyper := c.QueryTrim("hyper")
	hyperID := -1
	if hyper != "" {
		hyperID, err = strconv.Atoi(hyper)
		if err != nil {
			log.Println("Invalid image ID, %v", err)
			code := http.StatusBadRequest
			c.Error(code, http.StatusText(code))
			return
		}
	}
	if hyperID >= 0 {
		permit := memberShip.CheckPermission(model.Admin)
		if !permit {
			log.Println("Need Admin permissions")
			code := http.StatusUnauthorized
			c.Error(code, http.StatusText(code))
			return
		}
	}
	image := c.QueryTrim("image")
	imageID, err := strconv.Atoi(image)
	if err != nil {
		log.Println("Invalid image ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	flavor := c.QueryTrim("flavor")
	flavorID, err := strconv.Atoi(flavor)
	if err != nil {
		log.Println("Invalid flavor ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	primary := c.QueryTrim("primary")
	primaryID, err := strconv.Atoi(primary)
	if err != nil {
		log.Println("Invalid primary subnet ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err = memberShip.CheckAdmin(model.Writer, "subnets", int64(primaryID))
	if !permit {
		log.Println("Not authorized to access subnet")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	primaryIP := c.QueryTrim("primaryip")
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
			code := http.StatusUnauthorized
			c.Error(code, http.StatusText(code))
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
			code := http.StatusUnauthorized
			c.Error(code, http.StatusText(code))
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
				code := http.StatusUnauthorized
				c.Error(code, http.StatusText(code))
				return
			}
			sgIDs = append(sgIDs, int64(sgID))
		}
	} else {
		sgID := store.Get("defsg").(int64)
		permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
		if !permit {
			log.Println("Not authorized to access security group")
			code := http.StatusUnauthorized
			c.Error(code, http.StatusText(code))
			return
		}
		sgIDs = append(sgIDs, sgID)
	}
	userdata := c.QueryTrim("userdata")
	_, err = instanceAdmin.Create(c.Req.Context(), count, hostname, userdata, int64(imageID), int64(flavorID), int64(primaryID), primaryIP, subnetIDs, keyIDs, sgIDs, hyperID)
	if err != nil {
		log.Println("Create instance failed, %v", err)
		c.HTML(http.StatusBadRequest, err.Error())
	}
	c.Redirect(redirectTo)
}
