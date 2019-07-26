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

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	floatingipAdmin = &FloatingIpAdmin{}
	floatingipView  = &FloatingIpView{}
)

type FloatingIpAdmin struct{}
type FloatingIpView struct{}

func (a *FloatingIpAdmin) Create(ctx context.Context, instID int64, types []string) (floatingips []*model.FloatingIp, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	instance := &model.Instance{Model: model.Model{ID: instID}}
	err = db.Set("gorm:auto_preload", true).Preload("Interfaces", "primary_if = ?", true).Model(instance).Take(instance).Error
	if err != nil {
		log.Println("DB failed to query instance, %v", err)
		return
	}
	iface := instance.Interfaces[0]
	if iface.Address.Subnet.Router == 0 {
		err = fmt.Errorf("Floating IP can not be created without a gateway")
		log.Println("Floating IP can not be created without a gateway")
		return
	}
	gateway := &model.Gateway{Model: model.Model{ID: iface.Address.Subnet.Router}}
	err = db.Model(gateway).Set("gorm:auto_preload", true).Take(gateway).Error
	if err != nil {
		log.Println("DB failed to query gateway", err)
		return
	}
	for _, ftype := range types {
		if ftype != "private" && ftype != "public" {
			log.Println("Invalid floating ip type", err)
			return
		}
		floatingip := &model.FloatingIp{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, GatewayID: gateway.ID, InstanceID: instance.ID, Type: ftype}
		err = db.Create(floatingip).Error
		if err != nil {
			log.Println("DB failed to create floating ip", err)
			return
		}
		var fipIface *model.Interface
		fipIface, err = AllocateFloatingIp(ctx, floatingip.ID, memberShip.OrgID, gateway, ftype)
		if err != nil {
			log.Println("DB failed to allocate floating ip", err)
			return
		}
		floatingip.FipAddress = fipIface.Address.Address
		floatingip.IntAddress = iface.Address.Address
		err = db.Save(floatingip).Error
		if err != nil {
			log.Println("DB failed to update floating ip", err)
			return
		}
		floatingips = append(floatingips, floatingip)
		control := fmt.Sprintf("toall=router-%d:%d,%d", gateway.ID, gateway.Hyper, gateway.Peer)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_floating.sh %d %s %s %s", gateway.ID, ftype, floatingip.FipAddress, iface.Address.Address)
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
	ctx = saveTXtoCtx(ctx, db)
	floatingip := &model.FloatingIp{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Find(floatingip).Error; err != nil {
		log.Println("Failed to query floating ip", err)
		return
	}
	if floatingip.Gateway != nil {
		control := fmt.Sprintf("toall=router-%d:%d,%d", floatingip.Gateway.ID, floatingip.Gateway.Hyper, floatingip.Gateway.Peer)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_floating.sh %d %s %s %s", floatingip.GatewayID, floatingip.Type, floatingip.FipAddress, floatingip.IntAddress)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Create floating ip failed", err)
			return
		}
	}
	err = DeallocateFloatingIp(ctx, id)
	if err != nil {
		log.Println("DB failed to deallocate floating ip", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) List(ctx context.Context, offset, limit int64, order string) (total int64, floatingips []*model.FloatingIp, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	where := memberShip.GetWhere()
	floatingips = []*model.FloatingIp{}
	if err = db.Model(&model.FloatingIp{}).Where(where).Count(&total).Error; err != nil {
		log.Println("DB failed to count floating ip(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Set("gorm:auto_preload", true).Where(where).Find(&floatingips).Error; err != nil {
		log.Println("DB failed to query floating ip(s), %v", err)
		return
	}

	return
}

func (v *FloatingIpView) List(c *macaron.Context, store session.Store) {
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
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, floatingips, err := floatingipAdmin.List(c.Req.Context(), offset, limit, order)
	if err != nil {
		log.Println("Failed to list floating ip(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["FloatingIps"] = floatingips
	c.Data["Total"] = total
	c.HTML(200, "floatingips")
}

func (v *FloatingIpView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	floatingipID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid floating ip ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "floating_ips", int64(floatingipID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = floatingipAdmin.Delete(c.Req.Context(), int64(floatingipID))
	if err != nil {
		log.Println("Failed to delete floating ip, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
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
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	db := DB()
	where := memberShip.GetWhere()
	instances := []*model.Instance{}
	if err := db.Preload("Interfaces", "primary_if = ?", true).Preload("Interfaces.Address").Where(where).Find(&instances).Error; err != nil {
		return
	}
	c.Data["Instances"] = instances
	c.HTML(200, "floatingips_new")
}

func (v *FloatingIpView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../floatingips"
	instance := c.Query("instance")
	ftype := c.Query("ftype")
	instID, err := strconv.Atoi(instance)
	if err != nil {
		log.Println("Invalid interface ID", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err = memberShip.CheckOwner(model.Writer, "instances", int64(instID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	if ftype == "" {
		ftype = "public,private"
	}
	types := strings.Split(ftype, ",")
	_, err = floatingipAdmin.Create(c.Req.Context(), int64(instID), types)
	if err != nil {
		log.Println("Failed to create floating ip", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}

func AllocateFloatingIp(ctx context.Context, floatingipID, owner int64, gateway *model.Gateway, ftype string) (fipIface *model.Interface, err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
	var subnet *model.Subnet
	for _, iface := range gateway.Interfaces {
		if strings.Contains(iface.Type, ftype) {
			subnet = iface.Address.Subnet
			break
		}
	}
	if subnet == nil {
		err = fmt.Errorf("Invalid gateway subnet")
		return
	}
	name := ftype + "fip"
	fipIface, err = CreateInterface(ctx, subnet.ID, floatingipID, owner, "", name, "floating", nil)
	if err != nil {
		subnets := []*model.Subnet{}
		err = db.Where("vlan = ? and id <> ?", subnet.Vlan, subnet.ID).Find(&subnets).Error
		if err == nil && len(subnets) > 0 {
			for _, s := range subnets {
				fipIface, err = CreateInterface(ctx, s.ID, floatingipID, owner, "", name, "floating", nil)
				if err == nil {
					break
				}
			}
		} else {
			err = fmt.Errorf("No valid external subnets")
		}
	}
	return
}

func DeallocateFloatingIp(ctx context.Context, floatingipID int64) (err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
	DeleteInterfaces(ctx, floatingipID, 0, "floating")
	floatingip := &model.FloatingIp{Model: model.Model{ID: floatingipID}}
	err = db.Delete(floatingip).Error
	if err != nil {
		log.Println("Failed to delete floating ip, %v", err)
		return
	}
	return
}
