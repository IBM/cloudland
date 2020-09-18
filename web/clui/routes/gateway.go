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

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	gatewayAdmin = &GatewayAdmin{}
	gatewayView  = &GatewayView{}
)

type StaticRoute struct {
	Destination string `json:"destination"`
	Nexthop     string `json:"nexthop"`
}

type SubnetIface struct {
	Address string         `json:"ip_address"`
	Vni     int64          `json:"vni"`
	Routes  []*StaticRoute `json:"routes,omitempty"`
}

type GatewayAdmin struct{}
type GatewayView struct{}

func createGatewayIface(ctx context.Context, rtype string, gateway *model.Gateway, owner int64) (iface *model.Interface, subnet *model.Subnet, err error) {
	db := DB()
	subnets := []*model.Subnet{}
	err = db.Where("type = ?", rtype).Find(&subnets).Error
	if err != nil {
		log.Println("Failed to query subnets", err)
		return
	}
	name := ""
	ifType := ""
	for _, subnet = range subnets {
		if rtype == "public" {
			name = fmt.Sprintf("pub%d", subnet.ID)
			ifType = "gateway_public"
		} else if rtype == "private" {
			name = fmt.Sprintf("pub%d", subnet.ID)
			ifType = "gateway_private"
		}
		iface, err = CreateInterface(ctx, subnet.ID, gateway.ID, owner, gateway.Hyper, "", "", name, ifType, nil)
		if err == nil {
			log.Println("Created gateway interface from subnet", err)
			break
		}
	}
	return
}

func (a *GatewayAdmin) Create(ctx context.Context, name, stype string, pubID, priID int64, subnetIDs []int64, owner int64) (gateway *model.Gateway, err error) {
	memberShip := GetMemberShip(ctx)
	if owner == 0 {
		owner = memberShip.OrgID
	}
	db := DB()
	vni, err := getValidVni()
	if err != nil {
		log.Println("Failed to get valid vrrp vni %s, %v", vni, err)
		return
	}
	gateway = &model.Gateway{Model: model.Model{Creater: memberShip.UserID, Owner: owner}, Name: name, Type: stype, VrrpVni: int64(vni), VrrpAddr: "169.254.169.250/24", PeerAddr: "169.254.169.251/24", Status: "pending"}
	err = db.Create(gateway).Error
	if err != nil {
		log.Println("DB failed to create gateway, %v", err)
		return
	}
	var pubIface *model.Interface
	var pubSubnet *model.Subnet
	if pubID == 0 {
		pubIface, pubSubnet, err = createGatewayIface(ctx, "public", gateway, owner)
		if err != nil {
			log.Println("DB failed to create public interface", err)
			return
		}
	} else {
		pubSubnet = &model.Subnet{Model: model.Model{ID: pubID}}
		err = db.Model(pubSubnet).Where(pubSubnet).Take(pubSubnet).Error
		if err != nil {
			log.Println("DB failed to query public subnet, %v", err)
			return
		}
		pubIface, err = CreateInterface(ctx, pubSubnet.ID, gateway.ID, owner, gateway.Hyper, "", "", fmt.Sprintf("pub%d", pubSubnet.ID), "gateway_public", nil)
		if err != nil {
			log.Println("DB failed to create public interface, %v", err)
			return
		}
	}
	var priIface *model.Interface
	var priSubnet *model.Subnet
	if priID == 0 {
		priIface, priSubnet, err = createGatewayIface(ctx, "private", gateway, owner)
		if err != nil {
			log.Println("DB failed to create private interface", err)
			return
		}
	} else {
		priSubnet := &model.Subnet{Model: model.Model{ID: priID}}
		err = db.Model(priSubnet).Where(priSubnet).Take(priSubnet).Error
		if err != nil {
			log.Println("DB failed to query private subnet, %v", err)
			return
		}
		priIface, err = CreateInterface(ctx, priSubnet.ID, gateway.ID, owner, gateway.Hyper, "", "", fmt.Sprintf("pri%d", priSubnet.ID), "gateway_private", nil)
		if err != nil {
			log.Println("DB failed to create private interface, %v", err)
			return
		}
	}
	intIfaces := []*SubnetIface{}
	if subnetIDs != nil && len(subnetIDs) > 0 {
		for _, sID := range subnetIDs {
			var subnet *model.Subnet
			subnet, err = SetGateway(ctx, sID, gateway.ID)
			if err != nil {
				log.Println("DB failed to set gateway, %v", err)
				return
			}
			routes := []*StaticRoute{}
			err = json.Unmarshal([]byte(subnet.Routes), &routes)
			if err != nil {
				log.Println("Failed to unmarshal routes", err)
			}
			intIfaces = append(intIfaces, &SubnetIface{Address: subnet.Gateway, Vni: subnet.Vlan, Routes: routes})
		}
	}
	jsonData, err := json.Marshal(intIfaces)
	if err != nil {
		log.Println("Failed to marshal gateway json data, %v", err)
		return
	}
	control := fmt.Sprintf("inter=")
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_router.sh '%d' '%s' '%s' '%s' '%s' '%d' '%s' 'MASTER' <<EOF\n%s\nEOF", gateway.ID, pubSubnet.Gateway, priSubnet.Gateway, pubIface.Address.Address, priIface.Address.Address, vni, gateway.VrrpAddr, jsonData)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create master router command execution failed, %v", err)
		return
	}
	control = fmt.Sprintf("inter=")
	command = fmt.Sprintf("/opt/cloudland/scripts/backend/create_router.sh '%d' '%s' '%s' '%s' '%s' '%d' '%s' 'SLAVE' <<EOF\n%s\nEOF", gateway.ID, pubSubnet.Gateway, priSubnet.Gateway, pubIface.Address.Address, priIface.Address.Address, vni, gateway.PeerAddr, jsonData)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create peer router command execution failed, %v", err)
		return
	}
	return
}

func (a *GatewayAdmin) Update(ctx context.Context, id int64, name string, pubID, priID int64, subnetIDs []int64) (gateway *model.Gateway, err error) {
	db := DB()
	gateway = &model.Gateway{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Find(gateway).Error; err != nil {
		log.Println("Failed to query gateway ", err)
		return
	}
	if gateway.Name != name {
		gateway.Name = name
		if err = db.Save(gateway).Error; err != nil {
			log.Println("Failed to save gateway", err)
			return
		}
	}
	for _, gsub := range gateway.Subnets {
		found := false
		for _, sID := range subnetIDs {
			if gsub.ID == sID {
				found = true
				log.Println("Found SID ", sID)
				break
			}
		}
		if found == false {
			control := fmt.Sprintf("toall=router-%d:%d,%d", gateway.ID, gateway.Hyper, gateway.Peer)
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_gateway.sh '%d' '%s' '%d'", gateway.ID, gsub.Gateway, gsub.Vlan)
			err = hyperExecute(ctx, control, command)
			if err != nil {
				log.Println("Clear gateway failed")
				continue
			}
			err = UnsetGateway(ctx, gsub)
			if err != nil {
				log.Println("DB failed to update router for subnet", err)
				continue
			}
		}
	}
	for _, sID := range subnetIDs {
		found := false
		for _, gsub := range gateway.Subnets {
			if gsub.ID == sID {
				found = true
				break
			}
		}
		if found == false {
			sub := &model.Subnet{Model: model.Model{ID: sID}}
			err = db.Model(sub).Take(sub).Error
			if err != nil {
				log.Println("DB failed to query subnet, %v", err)
				continue
			}
			control := fmt.Sprintf("toall=router-%d:%d,%d", gateway.ID, gateway.Hyper, gateway.Peer)
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_gw_route.sh '%d' '%s' '%d' 'soft' <<EOF\n%s\nEOF", gateway.ID, sub.Gateway, sub.Vlan, sub.Routes)
			err = hyperExecute(ctx, control, command)
			if err != nil {
				log.Println("Set gateway failed")
				continue
			}
			_, err = SetGateway(ctx, sub.ID, gateway.ID)
			if err != nil {
				log.Println("DB failed to update router for subnet", err)
				continue
			}
		}
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
	ctx = saveTXtoCtx(ctx, db)
	count := 0
	err = db.Model(&model.FloatingIp{}).Where("gateway_id = ?", id).Count(&count).Error
	if err != nil {
		log.Println("Failed to count floating ip")
		return
	}
	if count > 0 {
		log.Println("There are floating ips")
		return
	}
	count = 0
	err = db.Model(&model.Portmap{}).Where("gateway_id = ?", id).Count(&count).Error
	if err != nil {
		log.Println("Failed to count portmap")
		return
	}
	if count > 0 {
		log.Println("There are floating ips")
		return
	}
	gateway := &model.Gateway{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(gateway).Error; err != nil {
		log.Println("Failed to query gateway", err)
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
	if err = DeleteInterfaces(ctx, id, 0, "gateway"); err != nil {
		log.Println("DB failed to delete interfaces, %v", err)
		return
	}
	control := "toall="
	if gateway.Hyper != -1 {
		control = fmt.Sprintf("inter=%d", gateway.Hyper)
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_router.sh '%d' '%d' <<EOF\n%s\nEOF", gateway.ID, gateway.VrrpVni, jsonData)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete master failed")
	}
	if control != "toall=" {
		control = "toall="
		if gateway.Peer != -1 {
			control = fmt.Sprintf("inter=%d", gateway.Peer)
		}
		command = fmt.Sprintf("/opt/cloudland/scripts/backend/clear_router.sh '%d' '%d' <<EOF\n%s\nEOF", gateway.ID, gateway.VrrpVni, jsonData)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete slave failed")
		}
	}
	if err = db.Delete(gateway).Error; err != nil {
		log.Println("DB failed to delete gateway", err)
		return
	}
	return
}

func (a *GatewayAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, gateways []*model.Gateway, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}
	where := memberShip.GetWhere()
	gateways = []*model.Gateway{}
	if err = db.Model(&model.Gateway{}).Where(where).Where(query).Count(&total).Error; err != nil {
		log.Println("DB failed to count gateway, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Where(where).Where(query).Find(&gateways).Error; err != nil {
		log.Println("DB failed to query gateways, %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, gateway := range gateways {
			gateway.OwnerInfo = &model.Organization{Model: model.Model{ID: gateway.Owner}}
			if err = db.Take(gateway.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				return
			}
		}
	}
	return
}

func (v *GatewayView) List(c *macaron.Context, store session.Store) {
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
	total, gateways, err := gatewayAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list gateways, %v", err)
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
	c.Data["Gateways"] = gateways
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"gateways": gateways,
			"total":    total,
			"pages":    pages,
			"query":    query,
		})
		return
	}
	c.HTML(200, "gateways")
}

func (v *GatewayView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		log.Println("Id is empty")
		c.Data["ErrorMsg"] = "Id is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	gatewayID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid gateway id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "gateways", int64(gatewayID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = gatewayAdmin.Delete(c.Req.Context(), int64(gatewayID))
	if err != nil {
		log.Println("Failed to delete gateway, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "gateways",
	})
	return
}

func (v *GatewayView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, subnets, err := subnetAdmin.List(c.Req.Context(), 0, -1, "", "", "router = 0")
	if err != nil {
		log.Println("DB failed to query subnets, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Subnets"] = subnets
	c.HTML(200, "gateways_new")
}

func (v *GatewayView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := dbs.DB()
	id := c.Params("id")
	gatewayID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid gateway id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "gateways", int64(gatewayID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	gateway := &model.Gateway{Model: model.Model{ID: int64(gatewayID)}}
	if err = db.Set("gorm:auto_preload", true).Find(gateway).Error; err != nil {
		log.Println("Failed to query gateway, %v", err)
		return
	}
	subnets := []*model.Subnet{}
	where := "type = 'internal'"
	for _, gsub := range gateway.Subnets {
		where = fmt.Sprintf("%s and id != %d", where, gsub.ID)
	}
	if err := db.Where(where).Where(memberShip.GetWhere()).Find(&subnets).Error; err != nil {
		log.Println("DB failed to query subnets, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Gateway"] = gateway
	c.Data["Subnets"] = subnets
	c.HTML(200, "gateways_patch")
}

func (v *GatewayView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../gateways"
	id := c.Params("id")
	gatewayID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid gateway id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "gateways", int64(gatewayID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	name := c.QueryTrim("name")
	pubSubnet := c.QueryTrim("public")
	priSubnet := c.QueryTrim("private")
	subnets := c.QueryStrings("subnets")
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
	var subnetIDs []int64
	for _, s := range subnets {
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
	gateway, err := gatewayAdmin.Update(c.Req.Context(), int64(gatewayID), name, int64(pubID), int64(priID), subnetIDs)
	if err != nil {
		log.Println("Failed to create gateway", err)
		c.Data["ErrorMsg"] = err.Error()
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, gateway)
		return
	}
	c.Redirect(redirectTo)
}

func (v *GatewayView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../gateways"
	name := c.QueryTrim("name")
	pubSubnet := c.QueryTrim("public")
	priSubnet := c.QueryTrim("private")
	subnets := c.QueryTrim("subnets")
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
		permit, err = memberShip.CheckOwner(model.Writer, "subnets", int64(sID))
		if !permit {
			log.Println("Not authorized for this operation")
			c.Data["ErrorMsg"] = "Not authorized for this operation"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		subnetIDs = append(subnetIDs, int64(sID))
	}
	_, err = gatewayAdmin.Create(c.Req.Context(), name, "", int64(pubID), int64(priID), subnetIDs, memberShip.OrgID)
	if err != nil {
		log.Println("Failed to create gateway, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
