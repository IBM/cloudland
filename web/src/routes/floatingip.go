/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	floatingipAdmin = &FloatingIpAdmin{}
	floatingipView  = &FloatingIpView{}
)

type FloatingIps struct {
	Instance  int64  `json:"instance"`
	PublicIp  string `json:"public_ip"`
	PrivateIp string `json:"private_ip"`
}

type FloatingIpAdmin struct{}
type FloatingIpView struct{}

func (a *FloatingIpAdmin) Create(ctx context.Context, instID, ifaceID int64, types []string, publicIp string) (floatingips []*model.FloatingIp, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	instance := &model.Instance{Model: model.Model{ID: instID}}
	err = db.Model(instance).Take(instance).Error
	if err != nil {
		log.Println("DB failed to query instance, %v", err)
		return
	}
	err = db.Preload("Address").Preload("Address.Subnet").Where("instance = ? and primary_if = ?", instID, true).Find(&instance.Interfaces).Error
	if err != nil {
		log.Println("DB failed to query interfaces, %v", err)
		return
	}
	err = db.Where("instance_id = ?", instID).Find(&instance.FloatingIps).Error
	if err == nil && instance.FloatingIps != nil && len(instance.FloatingIps) > 0 {
		log.Println("DB failed to query instance, %v", err)
		floatingips = instance.FloatingIps
		return
	}
	iface := instance.Interfaces[0]
	routerID := iface.Address.Subnet.RouterID
	if routerID <= 0 {
		err = fmt.Errorf("Floating IP can not be created without a router")
		log.Println("Floating IP can not be created without a router")
		return
	}
	router := &model.Router{Model: model.Model{ID: routerID}}
	err = db.Model(router).Take(router).Error
	if err != nil {
		log.Println("DB failed to query router, %v", err)
		return
	}
	for _, ftype := range types {
		floatingip := &model.FloatingIp{Model: model.Model{Creater: memberShip.UserID}, Owner: memberShip.OrgID, InstanceID: instID, RouterID: router.ID, Type: ftype}
		err = db.Create(floatingip).Error
		if err != nil {
			log.Println("DB failed to create floating ip", err)
			return
		}
		address := publicIp
		var fipIface *model.Interface
		var gateway string
		fipIface, gateway, err = AllocateFloatingIp(ctx, floatingip.ID, memberShip.OrgID, router, ftype, address)
		if err != nil {
			log.Println("DB failed to allocate floating ip", err)
			return
		}
		floatingip.FipAddress = fipIface.Address.Address
		floatingip.IntAddress = iface.Address.Address
		floatingip.IPAddress = strings.Split(floatingip.FipAddress, "/")[0]
		err = db.Save(floatingip).Error
		if err != nil {
			log.Println("DB failed to update floating ip", err)
			return
		}
		floatingips = append(floatingips, floatingip)
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_floating.sh '%d' '%s' '%s' '%s'", router.ID, floatingip.FipAddress, iface.Address.Address, gateway)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Create floating ip failed", err)
			return
		}
	}
	return
}

func (a *FloatingIpAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	ctx = SaveTXtoCtx(ctx, db)
	floatingip := &model.FloatingIp{Model: model.Model{ID: id}}
	if err = db.Preload("Instance").Find(floatingip).Error; err != nil {
		log.Println("Failed to query floating ip", err)
		return
	}
	control := fmt.Sprintf("inter=%d", floatingip.Instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_floating.sh '%d' '%s' '%s'", floatingip.RouterID, floatingip.FipAddress, floatingip.IntAddress)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create floating ip failed", err)
		return
	}
	err = DeallocateFloatingIp(ctx, id)
	if err != nil {
		log.Println("DB failed to deallocate floating ip", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, floatingips []*model.FloatingIp, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}
	if query != "" {
		query = fmt.Sprintf("fip_address like '%%%s%%' or int_address like '%%%s%%'", query, query)
	}

	where := memberShip.GetWhere()
	floatingips = []*model.FloatingIp{}
	if err = db.Model(&model.FloatingIp{}).Where(where).Where(query).Count(&total).Error; err != nil {
		log.Println("DB failed to count floating ip(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Instance").Preload("Instance.Zone").Where(where).Where(query).Find(&floatingips).Error; err != nil {
		log.Println("DB failed to query floating ip(s), %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, fip := range floatingips {
			fip.OwnerInfo = &model.Organization{Model: model.Model{ID: fip.Owner}}
			if err = db.Take(fip.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				return
			}
		}
	}

	return
}

func (v *FloatingIpView) List(c *macaron.Context, store session.Store) {
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
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, floatingips, err := floatingipAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list floating ip(s), %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	pages := GetPages(total, limit)
	c.Data["FloatingIps"] = floatingips
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"floatingips": floatingips,
			"total":       total,
			"pages":       pages,
			"query":       query,
		})
		return
	}
	c.HTML(200, "floatingips")
}

func (v *FloatingIpView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "id does not exist"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	floatingipID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid floating ip ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "floating_ips", int64(floatingipID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = floatingipAdmin.Delete(c.Req.Context(), int64(floatingipID))
	if err != nil {
		log.Println("Failed to delete floating ip, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "floatingips",
	})
	return
}

func (v *FloatingIpView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	where := memberShip.GetWhere()
	instances := []*model.Instance{}
	err := db.Where(where).Find(&instances).Error
	if err != nil {
		log.Println("Failed to query instances %v", err)
		return
	}
	for _, instance := range instances {
		if err = db.Preload("Address").Preload("Address.Subnet").Where("instance = ? and primary_if = true", instance.ID).Find(&instance.Interfaces).Error; err != nil {
			log.Println("Failed to query interfaces %v", err)
			return
		}
	}
	c.Data["Instances"] = instances
	c.HTML(200, "floatingips_new")
}

func (v *FloatingIpView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../floatingips"
	instID := c.QueryInt64("instance")
	ftype := c.QueryTrim("ftype")
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if ftype == "" {
		ftype = "public"
	}
	publicIp := c.QueryTrim("publicip")
	types := strings.Split(ftype, ",")
	floatingips, err := floatingipAdmin.Create(c.Req.Context(), int64(instID), 0, types, publicIp)
	if err != nil {
		log.Println("Failed to create floating ip", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, floatingips)
		return
	}
	c.Redirect(redirectTo)
}

func (v *FloatingIpView) Assign(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instID := c.QueryInt64("instance")
	floatingIP := c.QueryTrim("floatingIP")
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	types := []string{"public"}
	fipsData := &FloatingIps{Instance: instID}
	floatingips, err := floatingipAdmin.Create(c.Req.Context(), int64(instID), 0, types, floatingIP)
	if err != nil {
		log.Println("Failed to create floating ip", err)
		fipsData.PublicIp = ""
	} else {
		for _, fip := range floatingips {
			if fip.Type == "public" {
				fipsData.PublicIp = fip.FipAddress
			} else if fip.Type == "private" {
				fipsData.PrivateIp = fip.FipAddress
			}
		}
	}
	c.JSON(200, fipsData)
}

func AllocateFloatingIp(ctx context.Context, floatingipID, owner int64, router *model.Router, ftype, address string) (fipIface *model.Interface, gateway string, err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	subnets := []*model.Subnet{}
	err = db.Where("type = 'public' and vlan = ?", router.PublicLink).Find(&subnets).Error
	if err != nil {
		log.Println("Failed to query subnets ", err)
		return
	}
	name := ftype + "fip"
	log.Printf("Subnets: %v\n", subnets)
	for _, subnet := range subnets {
		fipIface, err = CreateInterface(ctx, subnet, floatingipID, owner, -1, address, "", name, "floating", nil)
		if err == nil {
			log.Printf("FipIface: %v\n", fipIface)
			gateway = subnet.Gateway
			return
		}
	}
	return
}

func DeallocateFloatingIp(ctx context.Context, floatingipID int64) (err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	DeleteInterfaces(ctx, floatingipID, 0, "floating")
	floatingip := &model.FloatingIp{Model: model.Model{ID: floatingipID}}
	err = db.Delete(floatingip).Error
	if err != nil {
		log.Println("Failed to delete floating ip, %v", err)
		return
	}
	return
}
