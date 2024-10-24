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

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	routerAdmin = &RouterAdmin{}
	routerView  = &RouterView{}
)

type StaticRoute struct {
	Destination string `json:"destination"`
	Nexthop     string `json:"nexthop"`
}

type SubnetIface struct {
	Address string         `json:"ip_address"`
	MacAddr string         `json:"mac_address"`
	Vni     int64          `json:"vni"`
	Routes  []*StaticRoute `json:"routes,omitempty"`
}

type RouterAdmin struct{}
type RouterView struct{}

func createRouterIface(ctx context.Context, rtype string, router *model.Router, owner, zoneID int64) (iface *model.Interface, subnet *model.Subnet, err error) {
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
		} else {
			continue
		}
		iface, err = CreateInterface(ctx, subnet.ID, router.ID, owner, zoneID, router.Hyper, "", "", name, ifType, nil)
		if err == nil {
			log.Println("Created gateway interface from subnet")
			break
		}
	}
	return
}

func (a *RouterAdmin) Create(ctx context.Context, name, stype string, pubID, owner, zoneID int64) (router *model.Router, err error) {
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
	router = &model.Router{Model: model.Model{Creater: memberShip.UserID, Owner: owner}, Name: name, Type: stype, VrrpVni: int64(vni), VrrpAddr: "169.254.169.250/24", PeerAddr: "169.254.169.251/24", Status: "pending", ZoneID: zoneID}
	err = db.Create(router).Error
	if err != nil {
		log.Println("DB failed to create router, %v", err)
		return
	}
	var pubIface *model.Interface
	var pubSubnet *model.Subnet
	if pubID == 0 {
		pubIface, pubSubnet, err = createRouterIface(ctx, "public", router, owner, zoneID)
		if err != nil || pubIface == nil {
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
		pubIface, err = CreateInterface(ctx, pubSubnet.ID, router.ID, owner, zoneID, router.Hyper, "", "", fmt.Sprintf("pub%d", pubSubnet.ID), "gateway_public", nil)
		if err != nil {
			log.Println("DB failed to create public interface, %v", err)
			return
		}
	}
	router.PublicID = pubSubnet.ID
	if err = db.Save(router).Error; err != nil {
		log.Println("Failed to save router", err)
		return
	}
	hyperGroup, err := instanceAdmin.getHyperGroup("", zoneID)
	if err != nil {
		log.Println("No valid hypervisor", err)
		return
	}
	control := "select=" + hyperGroup
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_router.sh '%d' '%s' '%s' '%d' '%s' 'MASTER'", router.ID, pubSubnet.Gateway, pubIface.Address.Address, vni, router.VrrpAddr)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create master router command execution failed, %v", err)
		return
	}
	return
}

func (a *RouterAdmin) Update(ctx context.Context, id int64, name string, pubID int64) (router *model.Router, err error) {
	db := DB()
	router = &model.Router{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Find(router).Error; err != nil {
		log.Println("Failed to query router", err)
		return
	}
	if router.Name != name {
		router.Name = name
		if err = db.Save(router).Error; err != nil {
			log.Println("Failed to save router", err)
			return
		}
	}
	return
}

func (a *RouterAdmin) Delete(ctx context.Context, id int64) (err error) {
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
	err = db.Model(&model.FloatingIp{}).Where("router_id = ?", id).Count(&count).Error
	if err != nil {
		log.Println("Failed to count floating ip")
		return
	}
	if count > 0 {
		log.Println("There are floating ips")
		return
	}
	count = 0
	err = db.Model(&model.Portmap{}).Where("router_id = ?", id).Count(&count).Error
	if err != nil {
		log.Println("Failed to count portmap")
		return
	}
	if count > 0 {
		log.Println("There are floating ips")
		return
	}
	router := &model.Router{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(router).Error; err != nil {
		log.Println("Failed to query router", err)
		return
	}
	intIfaces := []*SubnetIface{}
	for _, subnet := range router.Subnets {
		intIfaces = append(intIfaces, &SubnetIface{Address: subnet.Gateway, Vni: subnet.Vlan})
	}
	jsonData, err := json.Marshal(intIfaces)
	if err != nil {
		log.Println("Failed to marshal router json data, %v", err)
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
	if router.Hyper != -1 {
		control = fmt.Sprintf("inter=%d", router.Hyper)
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_router.sh '%d' '%d' <<EOF\n%s\nEOF", router.ID, router.VrrpVni, jsonData)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete master failed")
	}
	if control != "toall=" {
		control = "toall="
		if router.Peer != -1 {
			control = fmt.Sprintf("inter=%d", router.Peer)
		}
		command = fmt.Sprintf("/opt/cloudland/scripts/backend/clear_router.sh '%d' '%d' <<EOF\n%s\nEOF", router.ID, router.VrrpVni, jsonData)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete slave failed")
		}
	}
	if err = db.Delete(router).Error; err != nil {
		log.Println("DB failed to delete router", err)
		return
	}
	return
}

func (a *RouterAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, routers []*model.Router, err error) {
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
	routers = []*model.Router{}
	if err = db.Model(&model.Router{}).Where(where).Where(query).Count(&total).Error; err != nil {
		log.Println("DB failed to count router, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Where(where).Where(query).Find(&routers).Error; err != nil {
		log.Println("DB failed to query routers, %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, router := range routers {
			router.OwnerInfo = &model.Organization{Model: model.Model{ID: router.Owner}}
			if err = db.Take(router.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				return
			}
		}
	}
	return
}

func (v *RouterView) List(c *macaron.Context, store session.Store) {
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
	total, routers, err := routerAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list routers, %v", err)
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
	c.Data["Routers"] = routers
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"routers": routers,
			"total":   total,
			"pages":   pages,
			"query":   query,
		})
		return
	}
	c.HTML(200, "routers")
}

func (v *RouterView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		log.Println("Id is empty")
		c.Data["ErrorMsg"] = "Id is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	routerID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid router id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "routers", int64(routerID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = routerAdmin.Delete(c.Req.Context(), int64(routerID))
	if err != nil {
		log.Println("Failed to delete router, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "routers",
	})
	return
}

func (v *RouterView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	zones := []*model.Zone{}
	err := DB().Find(&zones).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	subnets := []*model.Subnet{}
	err = DB().Set("gorm:auto_preload", true).Where("type = 'public'").Find(&subnets).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	subnetList := []*model.Subnet{}
	for _, subnet := range subnets {
		subnetList = append(subnetList, subnet)
	}
	c.Data["Subnets"] = subnetList
	c.Data["Zones"] = zones
	c.HTML(200, "routers_new")
}

func (v *RouterView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := dbs.DB()
	id := c.Params("id")
	routerID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid router id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "routers", int64(routerID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	router := &model.Router{Model: model.Model{ID: int64(routerID)}}
	if err = db.Set("gorm:auto_preload", true).Find(router).Error; err != nil {
		log.Println("Failed to query router, %v", err)
		return
	}
	subnets := []*model.Subnet{}
	where := "type = 'internal'"
	for _, gsub := range router.Subnets {
		where = fmt.Sprintf("%s and id != %d", where, gsub.ID)
	}
	if err := db.Where(where).Where(memberShip.GetWhere()).Find(&subnets).Error; err != nil {
		log.Println("DB failed to query subnets, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Router"] = router
	c.Data["Subnets"] = subnets
	c.HTML(200, "routers_patch")
}

func (v *RouterView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../routers"
	id := c.Params("id")
	routerID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid router id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "routers", int64(routerID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	name := c.QueryTrim("name")
	pubSubnet := c.QueryTrim("public")
	pubID, err := strconv.Atoi(pubSubnet)
	if err != nil {
		log.Println("Invalid public subnet id, %v", err)
		pubID = 0
	}
	router, err := routerAdmin.Update(c.Req.Context(), int64(routerID), name, int64(pubID))
	if err != nil {
		log.Println("Failed to create router", err)
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
		c.JSON(200, router)
		return
	}
	c.Redirect(redirectTo)
}

func (v *RouterView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../routers"
	name := c.QueryTrim("name")
	zoneID := c.QueryInt64("zone")
	pubSubnet := c.QueryTrim("public")
	pubID, err := strconv.Atoi(pubSubnet)
	if err != nil {
		log.Println("Invalid public subnet id, %v", err)
		pubID = 0
	}
	router, err := routerAdmin.Create(c.Req.Context(), name, "", int64(pubID), memberShip.OrgID, zoneID)
	if err != nil {
		log.Println("Failed to create router, %v", err)
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
		c.JSON(200, router)
		return
	}
	c.Redirect(redirectTo)
}
