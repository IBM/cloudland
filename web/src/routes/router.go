/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

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

func createRouterIface(ctx context.Context, rtype string, router *model.Router, owner int64) (iface *model.Interface, subnet *model.Subnet, err error) {
	db := DB()
	subnets := []*model.Subnet{}
	err = db.Where("type = ?", rtype).Find(&subnets).Error
	if err != nil {
		logger.Error("Failed to query subnets", err)
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
		iface, err = CreateInterface(ctx, subnet, router.ID, owner, router.Hyper, 0, 0, "", "", name, ifType, nil)
		if err == nil {
			logger.Error("Created gateway interface from subnet")
			break
		}
	}
	return
}

func (a *RouterAdmin) Create(ctx context.Context, name string) (router *model.Router, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized to create routers")
		err = fmt.Errorf("Not authorized")
		return
	}
	owner := memberShip.OrgID
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	router = &model.Router{Model: model.Model{Creater: memberShip.UserID}, Owner: owner, Name: name, Status: "available"}
	err = db.Create(router).Error
	if err != nil {
		logger.Error("DB failed to create router ", err)
		return
	}
	secGroup, err := secgroupAdmin.Create(ctx, name+"-native", true, router)
	if err != nil {
		logger.Error("Failed to create security group", err)
		return
	}
	router.DefaultSG = secGroup.ID
	if err = db.Model(router).Update("default_sg", router.DefaultSG).Error; err != nil {
		logger.Error("Failed to save router", err)
		return
	}
	return
}

func (a *RouterAdmin) Get(ctx context.Context, id int64) (router *model.Router, err error) {
	if id <= 0 {
		logger.Error("returning nil router")
		return
	}
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	router = &model.Router{Model: model.Model{ID: id}}
	if err = db.Preload("Subnets").Where(where).Take(router).Error; err != nil {
		logger.Error("Failed to query router", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, router.Owner)
	if !permit {
		logger.Error("Not authorized to read the router")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *RouterAdmin) GetRouterByUUID(ctx context.Context, uuID string) (router *model.Router, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	router = &model.Router{}
	err = db.Preload("Subnets").Where(where).Where("uuid = ?", uuID).Take(router).Error
	if err != nil {
		logger.Error("Failed to query router, %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, router.Owner)
	if !permit {
		logger.Error("Not authorized to read the router")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *RouterAdmin) GetRouterByName(ctx context.Context, name string) (router *model.Router, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	router = &model.Router{}
	err = db.Preload("Subnets").Where(where).Where("name = ?", name).Take(router).Error
	if err != nil {
		logger.Error("Failed to query router, %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, router.Owner)
	if !permit {
		logger.Error("Not authorized to read the router")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *RouterAdmin) GetRouter(ctx context.Context, reference *BaseReference) (router *model.Router, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Router base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		router, err = a.GetRouterByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		router, err = a.GetRouterByName(ctx, reference.Name)
		return
	}
	return
}

func (a *RouterAdmin) Update(ctx context.Context, id int64, name string, pubID int64) (router *model.Router, err error) {
	db := DB()
	router = &model.Router{Model: model.Model{ID: id}}
	if err = db.Find(router).Error; err != nil {
		logger.Error("Failed to query router", err)
		return
	}
	if router.Name != name {
		router.Name = name
		if err = db.Model(router).Update("name", router.Name).Error; err != nil {
			logger.Error("Failed to save router", err)
			return
		}
	}
	return
}

func (a *RouterAdmin) Delete(ctx context.Context, router *model.Router) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, router.Owner)
	if !permit {
		logger.Error("Not authorized to delete the router")
		err = fmt.Errorf("Not authorized")
		return
	}
	count := 0
	err = db.Model(&model.FloatingIp{}).Where("router_id = ?", router.ID).Count(&count).Error
	if err != nil {
		logger.Error("Failed to count floating ip")
		err = fmt.Errorf("Failed to count floating ip")
		return
	}
	if count > 0 {
		logger.Error("There are floating ips")
		err = fmt.Errorf("There areassociated floating ips")
		return
	}
	count = 0
	err = db.Model(&model.Subnet{}).Where("router_id = ?", router.ID).Count(&count).Error
	if err != nil {
		logger.Error("Failed to count subnet")
		err = fmt.Errorf("Failed to count subnet")
		return
	}
	if count > 0 {
		logger.Error("There are associated subnets")
		err = fmt.Errorf("There are associated subnets")
		return
	}
	err = db.Model(&model.Portmap{}).Where("router_id = ?", router.ID).Count(&count).Error
	if err != nil {
		logger.Error("Failed to count portmap")
		return
	}
	if count > 0 {
		logger.Error("There are associated portmaps")
		err = fmt.Errorf("There are associated portmaps")
		return
	}
	control := "toall="
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_local_router.sh '%d'", router.ID)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Delete master failed")
		return
	}
	router.Name = fmt.Sprintf("%s-%d", router.Name, router.CreatedAt.Unix())
	err = db.Model(router).Update("name", router.Name).Error
	if err != nil {
		logger.Error("DB failed to update router name", err)
		return
	}
	if err = db.Delete(router).Error; err != nil {
		logger.Error("DB failed to delete router", err)
		return
	}
	secgroups := []*model.SecurityGroup{}
	err = db.Where("router_id = ?", router.ID).Find(&secgroups).Error
	if err != nil {
		logger.Error("DB failed to query security groups", err)
		return
	}
	for _, sg := range secgroups {
		err = secgroupAdmin.Delete(ctx, sg)
		if err != nil {
			logger.Error("Can not delete security group", err)
			return
		}
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
		logger.Error("DB failed to count router, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Subnets").Where(where).Where(query).Find(&routers).Error; err != nil {
		logger.Error("DB failed to query routers, %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, router := range routers {
			router.OwnerInfo = &model.Organization{Model: model.Model{ID: router.Owner}}
			if err = db.Take(router.OwnerInfo).Error; err != nil {
				logger.Error("Failed to query owner info", err)
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
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.Error(http.StatusBadRequest)
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
		logger.Error("Failed to list routers, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["Routers"] = routers
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.HTML(200, "routers")
}

func (v *RouterView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		logger.Error("Id is empty")
		c.Data["ErrorMsg"] = "Id is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	routerID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid router id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	router, err := routerAdmin.Get(ctx, int64(routerID))
	if err != nil {
		logger.Error("Not able to get vpc")
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = routerAdmin.Delete(ctx, router)
	if err != nil {
		logger.Error("Failed to delete router, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
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
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.Error(http.StatusBadRequest)
		return
	}
	c.HTML(200, "routers_new")
}

func (v *RouterView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	routerID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid router id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "routers", int64(routerID))
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.Error(http.StatusBadRequest)
		return
	}
	router := &model.Router{Model: model.Model{ID: int64(routerID)}}
	if err = db.Find(router).Error; err != nil {
		logger.Error("Failed to query router, %v", err)
		return
	}
	c.Data["Router"] = router
	c.HTML(200, "routers_patch")
}

func (v *RouterView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../routers"
	id := c.Params("id")
	routerID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid router id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "routers", int64(routerID))
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.Error(http.StatusBadRequest)
		return
	}
	name := c.QueryTrim("name")
	pubSubnet := c.QueryTrim("public")
	pubID, err := strconv.Atoi(pubSubnet)
	if err != nil {
		logger.Error("Invalid public subnet id, %v", err)
		pubID = 0
	}
	_, err = routerAdmin.Update(c.Req.Context(), int64(routerID), name, int64(pubID))
	if err != nil {
		logger.Error("Failed to create router", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.Redirect(redirectTo)
}

func (v *RouterView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../routers"
	name := c.QueryTrim("name")
	_, err := routerAdmin.Create(c.Req.Context(), name)
	if err != nil {
		logger.Error("Failed to create router, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.Redirect(redirectTo)
}
