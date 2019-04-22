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
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	gatewayAdmin = &GatewayAdmin{}
	gatewayView  = &GatewayView{}
)

type SubnetIface struct {
	Address string `json:"ip_address"`
	Vni     int64  `json:"vni"`
}

type GatewayAdmin struct{}
type GatewayView struct{}

func (a *GatewayAdmin) Create(ctx context.Context, name string, pubID, priID int64, subnetIDs []int64) (gateway *model.Gateway, err error) {
	db := DB()
	vni, err := getValidVni()
	if err != nil {
		log.Println("Failed to get valid vrrp vni %s, %v", vni, err)
		return
	}
	gateway = &model.Gateway{Name: name, VrrpVni: int64(vni), VrrpAddr: "169.254.169.250/24", PeerAddr: "169.254.169.251/24", Status: "pending"}
	err = db.Create(gateway).Error
	if err != nil {
		log.Println("DB failed to create gateway, %v", err)
		return
	}
	pubSubnet := &model.Subnet{Model: model.Model{ID: pubID}}
	if pubID == 0 {
		pubSubnet.Type = "public"
	}
	err = db.Model(pubSubnet).Where(pubSubnet).Take(pubSubnet).Error
	if err != nil {
		log.Println("DB failed to query public subnet, %v", err)
		return
	}
	pubIface, err := model.CreateInterface(pubSubnet.ID, gateway.ID, fmt.Sprintf("pub%d", pubSubnet.ID), "gateway_public")
	if err != nil {
		log.Println("DB failed to create public interface, %v", err)
		return
	}
	priSubnet := &model.Subnet{Model: model.Model{ID: priID}}
	if priID == 0 {
		priSubnet.Type = "private"
	}
	err = db.Model(priSubnet).Where(priSubnet).Take(priSubnet).Error
	if err != nil {
		log.Println("DB failed to query private subnet, %v", err)
		return
	}
	priIface, err := model.CreateInterface(priSubnet.ID, gateway.ID, fmt.Sprintf("pri%d", priSubnet.ID), "gateway_private")
	if err != nil {
		log.Println("DB failed to create private interface, %v", err)
		return
	}
	intIfaces := []*SubnetIface{}
	for _, sID := range subnetIDs {
		var subnet *model.Subnet
		subnet, err = model.SetGateway(sID, gateway.ID)
		if err != nil {
			log.Println("DB failed to set gateway, %v", err)
			return
		}
		intIfaces = append(intIfaces, &SubnetIface{Address: subnet.Gateway, Vni: subnet.Vlan})
	}
	jsonData, err := json.Marshal(intIfaces)
	if err != nil {
		log.Println("Failed to marshal gateway json data, %v", err)
		return
	}
	pubmask := net.IPMask(net.ParseIP(pubIface.Address.Netmask).To4())
	pubsize, _ := pubmask.Size()
	primask := net.IPMask(net.ParseIP(priIface.Address.Netmask).To4())
	prisize, _ := primask.Size()
	control := fmt.Sprintf("inter=")
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_router.sh %d %s/%d %s/%d %d %s 'MASTER' <<EOF\n%s\nEOF", gateway.ID, pubIface.Address.Address, pubsize, priIface.Address.Address, prisize, vni, gateway.VrrpAddr, jsonData)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create master router command execution failed, %v", err)
		return
	}
	control = fmt.Sprintf("inter=")
	command = fmt.Sprintf("/opt/cloudland/scripts/backend/create_router.sh %d %s/%d %s/%d %d %s 'SLAVE' <<EOF\n%s\nEOF", gateway.ID, pubIface.Address.Address, pubsize, priIface.Address.Address, prisize, vni, gateway.PeerAddr, jsonData)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create peer router command execution failed, %v", err)
		return
	}
	return
}

func (a *GatewayAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	gateway := &model.Gateway{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Find(gateway).Error; err != nil {
		log.Println("Failed to query gateway, %v", err)
		return
	}
	intIfaces := []*SubnetIface{}
	for _, subnet := range gateway.Subnets {
		intIfaces = append(intIfaces, &SubnetIface{Address: subnet.Gateway, Vni: subnet.Vlan})
	}
	jsonData, err := json.Marshal(intIfaces)
	if err != nil {
		log.Println("Failed to marshal gateway json data, %v", err)
		return
	}
	if err = db.Model(&model.Subnet{}).Where("router = ?", id).Update("router", 0).Error; err != nil {
		log.Println("DB failed to update router for subnet, %v", err)
		return
	}
	if err = model.DeleteInterfaces(id, "gateway"); err != nil {
		log.Println("DB failed to delete interfaces, %v", err)
		return
	}
	if err = db.Model(&model.Gateway{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("DB failed to delete gateway, %v", err)
		return
	}
	control := fmt.Sprintf("inter=%d", gateway.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_router.sh %d %d <<EOF\n%s\nEOF", gateway.ID, gateway.VrrpVni, jsonData)
	err = hyperExecute(ctx, control, command)
	control = fmt.Sprintf("inter=%d", gateway.Peer)
	command = fmt.Sprintf("/opt/cloudland/scripts/backend/clear_router.sh %d %d <<EOF\n%s\nEOF", gateway.ID, gateway.VrrpVni, jsonData)
	err = hyperExecute(ctx, control, command)
	return
}

func (a *GatewayAdmin) List(offset, limit int64, order string) (total int64, gateways []*model.Gateway, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	gateways = []*model.Gateway{}
	if err = db.Model(&model.Gateway{}).Count(&total).Error; err != nil {
		log.Println("DB failed to count gateway, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Find(&gateways).Error; err != nil {
		log.Println("DB failed to query gateways, %v", err)
		return
	}

	return
}

func (v *GatewayView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, gateways, err := gatewayAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list gateways, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Gateways"] = gateways
	c.Data["Total"] = total
	c.HTML(200, "gateways")
}

func (v *GatewayView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		log.Println("Id is empty")
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	gatewayID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid gateway id, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = gatewayAdmin.Delete(c.Req.Context(), int64(gatewayID))
	if err != nil {
		log.Println("Failed to delete gateway, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "gateways",
	})
	return
}

func (v *GatewayView) New(c *macaron.Context, store session.Store) {
	db := dbs.DB()
	subnets := []*model.Subnet{}
	if err := db.Find(&subnets).Error; err != nil {
		log.Println("DB failed to query subnets, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Subnets"] = subnets
	c.HTML(200, "gateways_new")
}

func (v *GatewayView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../gateways"
	name := c.Query("name")
	pubSubnet := c.Query("public")
	priSubnet := c.Query("private")
	subnets := c.Query("subnets")
	pubID, err := strconv.Atoi(pubSubnet)
	if err != nil {
		log.Println("Invalid public subnet id, %v", err)
		pubID = 0
	}
	priID, err := strconv.Atoi(priSubnet)
	if err != nil {
		log.Println("Invalid private subnet id, %v", err)
		priID = 0
	}
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
	_, err = gatewayAdmin.Create(c.Req.Context(), name, int64(pubID), int64(priID), subnetIDs)
	if err != nil {
		log.Println("Failed to create gateway, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
