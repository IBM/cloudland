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
	MacAddr string `json:"mac_address"`
}

type InstanceData struct {
	Userdata string             `json:"userdata"`
	Vlans    []*VlanInfo        `json:"vlans"`
	Networks []*InstanceNetwork `json:"networks"`
	Links    []*NetworkLink     `json:"links"`
	Keys     []string           `json:"keys"`
}

func (a *InstanceAdmin) Create(ctx context.Context, count int, prefix, userdata string, imageID, flavorID, primaryID int64, subnetIDs, keyIDs []int64) (instance *model.Instance, err error) {
	db := DB()
	image := &model.Image{Model: model.Model{ID: imageID}}
	if err = db.Take(image).Error; err != nil {
		log.Println("Image query failed, %v", err)
		return
	}
	flavor := &model.Flavor{Model: model.Model{ID: flavorID}}
	if err = db.Find(flavor).Error; err != nil {
		log.Println("Flavor query failed, %v", err)
		return
	}
	primary := &model.Subnet{Model: model.Model{ID: primaryID}}
	if err = db.Find(primary).Error; err != nil {
		log.Println("Primary subnet query failed, %v", err)
		return
	}
	subnets := []*model.Subnet{}
	if err = db.Where(subnetIDs).Take(&subnets).Error; err != nil {
		log.Println("Secondary subnets query failed, %v", err)
		return
	}
	keys := []*model.Key{}
	if err = db.Where(keyIDs).Find(&keys).Error; err != nil {
		log.Println("Keys query failed, %v", err)
		return
	}
	i := 0
	hostname := prefix
	for i < count {
		if count > 1 {
			hostname = fmt.Sprintf("%s-%d", prefix, i+1)
		}
		instance = &model.Instance{Hostname: hostname, ImageID: imageID, Image: image, FlavorID: flavorID, Flavor: flavor, Keys: keys, Userdata: userdata, Status: "pending"}
		err = db.Create(instance).Error
		if err != nil {
			log.Println("DB create instance failed, %v", err)
			return
		}
		metadata := ""
		_, metadata, err = a.buildMetadata(primary, subnets, keys, instance, userdata)
		if err != nil {
			log.Println("Build instance metadata failed, %v", err)
			return
		}
		control := fmt.Sprintf("inter= cpu=%d memory=%d disk=%d network=%d", 0, 0, 0, 0)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/launch_vm.sh %d image-%d.%s %s %d %d %d <<EOF\n%s\nEOF", instance.ID, image.ID, image.Format, hostname, flavor.Cpu, flavor.Memory, flavor.Disk, metadata)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed, %v", err)
			return
		}
		i++
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

func (a *InstanceAdmin) buildMetadata(primary *model.Subnet, subnets []*model.Subnet, keys []*model.Key, instance *model.Instance, userdata string) (interfaces []*model.Interface, metadata string, err error) {
	vlans := []*VlanInfo{}
	instNetworks := []*InstanceNetwork{}
	instLinks := []*NetworkLink{}
	gateway := strings.Split(primary.Gateway, "/")[0]
	instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: gateway}
	iface, err := model.CreateInterface(primary.ID, instance.ID, "eth0", "instance")
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
	vlans = append(vlans, &VlanInfo{Device: "eth0", Vlan: primary.Vlan, MacAddr: iface.MacAddr})
	for i, subnet := range subnets {
		ifname := fmt.Sprintf("eth%d", i+1)
		iface, err = model.CreateInterface(subnet.ID, instance.ID, ifname, "instance")
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
		vlans = append(vlans, &VlanInfo{Device: ifname, Vlan: subnet.Vlan, MacAddr: iface.MacAddr})
	}
	var instKeys []string
	for _, key := range keys {
		instKeys = append(instKeys, key.PublicKey)
	}
	instData := &InstanceData{
		Userdata: userdata,
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
			fip.InstanceID = 0
			err = db.Save(fip).Error
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
	if err = model.DeleteInterfaces(id, "instance"); err != nil {
		log.Println("DB failed to delete interfaces, %v", err)
		return
	}
	if err = db.Delete(&model.Instance{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("Failed to delete instance, %v", err)
		return
	}
	return
}

func (a *InstanceAdmin) List(offset, limit int64, order string) (total int64, instances []*model.Instance, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	instances = []*model.Instance{}
	if err = db.Model(&model.Instance{}).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Find(&instances).Error; err != nil {
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
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, instances, err := instanceAdmin.List(offset, limit, order)
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
	subnets := []*model.Subnet{}
	if err := db.Find(&subnets).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	secgroups := []*model.SecurityGroup{}
	if err := db.Find(&secgroups).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	keys := []*model.Key{}
	if err := db.Find(&keys).Error; err != nil {
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

func (v *InstanceView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../instances"
	hostname := c.Query("hostname")
	cnt := c.Query("count")
	count, err := strconv.Atoi(cnt)
	if err != nil {
		log.Println("Invalid instance count", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	image := c.Query("image")
	imageID, err := strconv.Atoi(image)
	if err != nil {
		log.Println("Invalid image ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	flavor := c.Query("flavor")
	flavorID, err := strconv.Atoi(flavor)
	if err != nil {
		log.Println("Invalid flavor ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	primary := c.Query("primary")
	primaryID, err := strconv.Atoi(primary)
	if err != nil {
		log.Println("Invalid primary subnet ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	subnets := c.Query("subnets")
	s := strings.Split(subnets, ",")
	var subnetIDs []int64
	for i := 0; i < len(s); i++ {
		sID, err := strconv.Atoi(s[i])
		if err != nil {
			log.Println("Invalid secondary subnet ID, %v", err)
			continue
		}
		subnetIDs = append(subnetIDs, int64(sID))
	}
	keys := c.Query("keys")
	k := strings.Split(keys, ",")
	var keyIDs []int64
	for i := 0; i < len(k); i++ {
		kID, err := strconv.Atoi(k[i])
		if err != nil {
			log.Println("Invalid key ID, %v", err)
			continue
		}
		keyIDs = append(keyIDs, int64(kID))
	}
	userdata := c.Query("userdata")
	_, err = instanceAdmin.Create(c.Req.Context(), count, hostname, userdata, int64(imageID), int64(flavorID), int64(primaryID), subnetIDs, keyIDs)
	if err != nil {
		log.Println("Create instance failed, %v", err)
		c.HTML(http.StatusBadRequest, err.Error())
	}
	c.Redirect(redirectTo)
}
