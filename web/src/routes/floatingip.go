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
	"strings"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	floatingIpAdmin = &FloatingIpAdmin{}
	floatingIpView  = &FloatingIpView{}
)

type FloatingIps struct {
	Instance  int64  `json:"instance"`
	PublicIp  string `json:"public_ip"`
	PrivateIp string `json:"private_ip"`
}

type FloatingIpAdmin struct{}
type FloatingIpView struct{}

func (a *FloatingIpAdmin) Create(ctx context.Context, instance *model.Instance, pubSubnet *model.Subnet, publicIp string, name string, inbound, outbound int32) (floatingIp *model.FloatingIp, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized for this operation")
		return
	}
	if pubSubnet != nil && pubSubnet.Type != "public" {
		logger.Error("Subnet must be public", err)
		err = fmt.Errorf("Subnet must be public")
		return
	}
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	floatingIp = &model.FloatingIp{Model: model.Model{Creater: memberShip.UserID}, Owner: memberShip.OrgID, Name: name, Inbound: inbound, Outbound: outbound}
	err = db.Create(floatingIp).Error
	if err != nil {
		logger.Error("DB failed to create floating ip", err)
		return
	}
	fipIface, err := AllocateFloatingIp(ctx, floatingIp.ID, memberShip.OrgID, pubSubnet, publicIp)
	if err != nil {
		logger.Error("DB failed to allocate floating ip", err)
		return
	}
	floatingIp.FipAddress = fipIface.Address.Address
	floatingIp.IPAddress = strings.Split(floatingIp.FipAddress, "/")[0]
	floatingIp.Interface = fipIface
	if instance != nil {
		err = a.Attach(ctx, floatingIp, instance)
		if err != nil {
			logger.Error("Execute floating ip failed", err)
			return
		}
	}
	err = db.Model(floatingIp).Updates(floatingIp).Error
	if err != nil {
		logger.Error("DB failed to update floating ip", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) Attach(ctx context.Context, floatingIp *model.FloatingIp, instance *model.Instance) (err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized for this operation")
		return
	}
	ctx, db := GetContextDB(ctx)
	if instance == nil || instance.Status != "running" {
		logger.Error("Instance is not running")
		err = fmt.Errorf("Instance must be running")
		return
	}
	instID := instance.ID
	routerID := instance.RouterID
	if routerID == 0 {
		logger.Error("Instance has no router")
		err = fmt.Errorf("Instance has no router")
		return
	}
	router := &model.Router{Model: model.Model{ID: routerID}}
	err = db.Take(router).Error
	if err != nil {
		logger.Error("DB failed to query router", err)
		return
	}
	var primaryIface *model.Interface
	for i, iface := range instance.Interfaces {
		if iface.PrimaryIf {
			primaryIface = instance.Interfaces[i]
			break
		}
	}
	if primaryIface == nil {
		err = fmt.Errorf("No primary interface for the instance, %d", instID)
		return
	}
	floatingIp.IntAddress = primaryIface.Address.Address
	floatingIp.InstanceID = instance.ID
	floatingIp.RouterID = instance.RouterID
	err = db.Model(floatingIp).Updates(floatingIp).Error
	if err != nil {
		logger.Error("DB failed to update floating ip", err)
		return
	}
	pubSubnet := floatingIp.Interface.Address.Subnet
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_floating.sh '%d' '%s' '%s' '%d' '%s' '%d' '%d' '%d' '%d'", router.ID, floatingIp.FipAddress, pubSubnet.Gateway, pubSubnet.Vlan, primaryIface.Address.Address, primaryIface.Address.Subnet.Vlan, floatingIp.ID, floatingIp.Inbound, floatingIp.Outbound)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Execute floating ip failed", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) Get(ctx context.Context, id int64) (floatingIp *model.FloatingIp, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid floatingIp ID: %d", id)
		logger.Error(err)
		return
	}
	memberShip := GetMemberShip(ctx)
	db := DB()
	where := memberShip.GetWhere()
	floatingIp = &model.FloatingIp{Model: model.Model{ID: id}}
	err = db.Preload("Interface").Preload("Interface.Address").Preload("Interface.Address.Subnet").Where(where).Take(floatingIp).Error
	if err != nil {
		logger.Error("DB failed to query floatingIp ", err)
		return
	}
	if floatingIp.InstanceID > 0 {
		floatingIp.Instance = &model.Instance{Model: model.Model{ID: floatingIp.InstanceID}}
		err = db.Take(floatingIp.Instance).Error
		if err != nil {
			logger.Error("DB failed to query instance ", err)
			return
		}
		instance := floatingIp.Instance
		err = db.Preload("Address").Preload("Address.Subnet").Where("instance = ? and primary_if = true", instance.ID).Find(&instance.Interfaces).Error
		if err != nil {
			logger.Error("Failed to query interfaces %v", err)
			return
		}
	}
	if floatingIp.RouterID > 0 {
		floatingIp.Router = &model.Router{Model: model.Model{ID: floatingIp.RouterID}}
		err = db.Take(floatingIp.Router).Error
		if err != nil {
			logger.Error("DB failed to query instance ", err)
			return
		}
	}
	return
}

func (a *FloatingIpAdmin) GetFloatingIpByUUID(ctx context.Context, uuID string) (floatingIp *model.FloatingIp, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	floatingIp = &model.FloatingIp{}
	err = db.Preload("Interface").Preload("Interface.Address").Preload("Interface.Address.Subnet").Where(where).Where("uuid = ?", uuID).Take(floatingIp).Error
	if err != nil {
		logger.Error("Failed to query floatingIp, %v", err)
		return
	}
	if floatingIp.InstanceID > 0 {
		floatingIp.Instance = &model.Instance{Model: model.Model{ID: floatingIp.InstanceID}}
		err = db.Take(floatingIp.Instance).Error
		if err != nil {
			logger.Error("DB failed to query instance ", err)
			return
		}
		instance := floatingIp.Instance
		err = db.Preload("Address").Preload("Address.Subnet").Where("instance = ? and primary_if = true", instance.ID).Find(&instance.Interfaces).Error
		if err != nil {
			logger.Error("Failed to query interfaces %v", err)
			return
		}
	}
	if floatingIp.RouterID > 0 {
		floatingIp.Router = &model.Router{Model: model.Model{ID: floatingIp.RouterID}}
		err = db.Take(floatingIp.Router).Error
		if err != nil {
			logger.Error("DB failed to query instance ", err)
			return
		}
	}
	return
}

func (a *FloatingIpAdmin) Detach(ctx context.Context, floatingIp *model.FloatingIp) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	if floatingIp.Instance != nil {
		var primaryIface *model.Interface
		instance := floatingIp.Instance
		for i, iface := range instance.Interfaces {
			if iface.PrimaryIf {
				primaryIface = instance.Interfaces[i]
				break
			}
		}
		control := fmt.Sprintf("inter=%d", floatingIp.Instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_floating.sh '%d' '%s' '%s' '%d' '%d'", floatingIp.RouterID, floatingIp.FipAddress, floatingIp.IntAddress, primaryIface.Address.Subnet.Vlan, floatingIp.ID)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Detach floating ip failed", err)
			return
		}
	}
	logger.Errorf("Floating ip: %v\n", floatingIp)
	floatingIp.InstanceID = 0
	floatingIp.Instance = nil
	err = db.Model(floatingIp).Where("id = ?", floatingIp.ID).Update(map[string]interface{}{"instance_id": 0}).Error
	if err != nil {
		logger.Error("Failed to update instance ID for floating ip", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) Delete(ctx context.Context, floatingIp *model.FloatingIp) (err error) {
	ctx, _, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	if floatingIp.Instance != nil {
		err = a.Detach(ctx, floatingIp)
		if err != nil {
			logger.Error("Failed to detach floating ip", err)
			return
		}
	}
	err = DeallocateFloatingIp(ctx, floatingIp.ID)
	if err != nil {
		logger.Error("DB failed to deallocate floating ip", err)
		return
	}
	return
}

func (a *FloatingIpAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, floatingIps []*model.FloatingIp, err error) {
	memberShip := GetMemberShip(ctx)
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}
	if query != "" {
		query = fmt.Sprintf("fip_address like '%%%s%%' or int_address like '%%%s%%' or name like '%%%s%%'", query, query)
	}

	db := DB()
	where := memberShip.GetWhere()
	floatingIps = []*model.FloatingIp{}
	if err = db.Model(&model.FloatingIp{}).Where(where).Where(query).Count(&total).Error; err != nil {
		logger.Error("DB failed to count floating ip(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Instance").Preload("Instance.Zone").Where(where).Where(query).Find(&floatingIps).Error; err != nil {
		logger.Error("DB failed to query floating ip(s), %v", err)
		return
	}
	for _, fip := range floatingIps {
		if fip.InstanceID <= 0 {
			continue
		}
		fip.Instance = &model.Instance{Model: model.Model{ID: fip.InstanceID}}
		err = db.Take(fip.Instance).Error
		if err != nil {
			logger.Error("DB failed to query instance ", err)
		}
		instance := fip.Instance
		err = db.Preload("Address").Where("instance = ? and primary_if = true", instance.ID).Find(&instance.Interfaces).Error
		if err != nil {
			logger.Error("Failed to query interfaces ", err)
			continue
		}
		if fip.RouterID > 0 {
			fip.Router = &model.Router{Model: model.Model{ID: fip.RouterID}}
			err = db.Take(fip.Router).Error
			if err != nil {
				logger.Error("DB failed to query instance ", err)
				continue
			}
		}
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, fip := range floatingIps {
			fip.OwnerInfo = &model.Organization{Model: model.Model{ID: fip.Owner}}
			if err = db.Take(fip.OwnerInfo).Error; err != nil {
				logger.Error("Failed to query owner info", err)
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
		logger.Error("Not authorized for this operation")
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
	total, floatingIps, err := floatingIpAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		logger.Error("Failed to list floating ip(s), %v", err)
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
	c.Data["FloatingIps"] = floatingIps
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"floatingIps": floatingIps,
			"total":       total,
			"pages":       pages,
			"query":       query,
		})
		return
	}
	c.HTML(200, "floatingips")
}

func (v *FloatingIpView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "id does not exist"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	floatingIpID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid floating ip ID ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	floatingIp, err := floatingIpAdmin.Get(ctx, int64(floatingIpID))
	if err != nil {
		logger.Error("Failed to get floating ip ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = floatingIpAdmin.Delete(ctx, floatingIp)
	if err != nil {
		logger.Error("Failed to delete floating ip, %v", err)
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
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	where := memberShip.GetWhere()
	instances := []*model.Instance{}
	err := db.Where(where).Find(&instances).Error
	if err != nil {
		logger.Error("Failed to query instances %v", err)
		return
	}
	for _, instance := range instances {
		if err = db.Preload("Address").Preload("Address.Subnet").Where("instance = ? and primary_if = true", instance.ID).Find(&instance.Interfaces).Error; err != nil {
			logger.Error("Failed to query interfaces %v", err)
			return
		}
	}
	c.Data["Instances"] = instances
	c.HTML(200, "floatingips_new")
}

func (v *FloatingIpView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "../floatingips"
	instID := c.QueryInt64("instance")
	publicIp := c.QueryTrim("publicip")
	name := c.QueryTrim("name")
	instance, err := instanceAdmin.Get(ctx, int64(instID))
	if err != nil {
		logger.Error("Failed to get instance ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, err = floatingIpAdmin.Create(c.Req.Context(), instance, nil, publicIp, name, 0, 0)
	if err != nil {
		logger.Error("Failed to create floating ip", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Redirect(redirectTo)
}

func AllocateFloatingIp(ctx context.Context, floatingIpID, owner int64, pubSubnet *model.Subnet, address string) (fipIface *model.Interface, err error) {
	ctx, db := GetContextDB(ctx)
	subnets := []*model.Subnet{}
	if pubSubnet != nil {
		subnets = append(subnets, pubSubnet)
	} else {
		where := "type = 'public'"
		err = db.Where(where).Find(&subnets).Error
		if err != nil || len(subnets) == 0 {
			logger.Error("Failed to query subnets ", err)
			return
		}
	}
	name := "fip"
	logger.Errorf("Subnets: %v\n", subnets)
	for _, subnet := range subnets {
		fipIface, err = CreateInterface(ctx, subnet, floatingIpID, owner, -1, 0, 0, address, "", name, "floating", nil)
		if err == nil {
			logger.Errorf("FipIface: %v\n", fipIface)
			break
		}
	}
	return
}

func DeallocateFloatingIp(ctx context.Context, floatingIpID int64) (err error) {
	ctx, db := GetContextDB(ctx)
	DeleteInterfaces(ctx, floatingIpID, 0, "floating")
	floatingIp := &model.FloatingIp{Model: model.Model{ID: floatingIpID}}
	err = db.Delete(floatingIp).Error
	if err != nil {
		logger.Error("Failed to delete floating ip, %v", err)
		return
	}
	return
}
