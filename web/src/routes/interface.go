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
	interfaceAdmin = &InterfaceAdmin{}
	interfaceView  = &InterfaceView{}
)

type InterfaceInfo struct {
	Subnet         *model.Subnet
	MacAddress     string
	IpAddress      string
	Inbound        int32
	Outbound       int32
	AllowSpoofing  bool
	SecurityGroups []*model.SecurityGroup
}

type InterfaceAdmin struct{}

type InterfaceView struct{}

func (a *InterfaceAdmin) Get(ctx context.Context, id int64) (iface *model.Interface, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid interface ID: %d", id)
		logger.Debug(err)
		return
	}
	memberShip := GetMemberShip(ctx)
	db := DB()
	iface = &model.Interface{Model: model.Model{ID: id}}
	err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Take(iface).Error
	if err != nil {
		logger.Debug("DB failed to query interface, %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, iface.Owner)
	if !permit {
		logger.Debug("Not authorized to read the subnet")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *InterfaceAdmin) GetInterfaceByUUID(ctx context.Context, uuID string) (iface *model.Interface, err error) {
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	db := DB()
	iface = &model.Interface{}
	err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where(where).Where("uuid = ?", uuID).Take(iface).Error
	if err != nil {
		logger.Debug("DB failed to query interface, %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, iface.Owner)
	if !permit {
		logger.Debug("Not authorized to read the subnet")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *InterfaceAdmin) List(ctx context.Context, offset, limit int64, order string, instance *model.Instance) (total int64, interfaces []*model.Interface, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Reader, instance.Owner)
	if !permit {
		logger.Debug("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	where := fmt.Sprintf("instance = %d", instance.ID)
	wm := memberShip.GetWhere()
	if wm != "" {
		where = fmt.Sprintf("%s and %s", where, wm)
	}
	interfaces = []*model.Interface{}
	if err = db.Model(&model.Interface{}).Where(where).Count(&total).Error; err != nil {
		logger.Debug("DB failed to count security rule(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("SecurityGroups").Preload("Address").Preload("Address.Subnet").Where(where).Find(&interfaces).Error; err != nil {
		logger.Debug("DB failed to query security rule(s), %v", err)
		return
	}

	return
}

func (a *InterfaceAdmin) Update(ctx context.Context, instance *model.Instance, iface *model.Interface, name string, inbound, outbound int32, secgroups []*model.SecurityGroup) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	needUpdate := false
	needRemoteUpdate := false
	if iface.Name != name {
		iface.Name = name
		needUpdate = true
	}
	if iface.Inbound != inbound {
		iface.Inbound = inbound
		needRemoteUpdate = true
	}
	if iface.Outbound != outbound {
		iface.Outbound = outbound
		needUpdate = true
		needRemoteUpdate = true
	}
	if len(secgroups) > 0 {
		iface.SecurityGroups = secgroups
		needRemoteUpdate = true
	}
	if needUpdate || needRemoteUpdate {
		if err = db.Model(iface).Save(iface).Error; err != nil {
			logger.Debug("Failed to save interface", err)
			return
		}
	}
	if needRemoteUpdate {
		err = ApplyInterface(ctx, instance, iface)
		if err != nil {
			logger.Error("Update vm nic command execution failed", err)
			return
		}
	}
	return
}

func (v *InterfaceView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	ifaceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", int64(ifaceID))
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	iface := &model.Interface{Model: model.Model{ID: int64(ifaceID)}}
	if err = db.Preload("Address").Preload("SecurityGroups").Take(iface).Error; err != nil {
		logger.Error("Image query failed", err)
		return
	}
	_, secgroups, err := secgroupAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Interface"] = iface
	c.Data["Secgroups"] = secgroups
	if len(iface.SecurityGroups) > 0 {
		c.Data["SecgroupID"] = iface.SecurityGroups[0].ID
	} else {
		c.Data["SecgroupID"] = -1
	}
	c.HTML(200, "interfaces_patch")
}

func (v *InterfaceView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	subnetID := c.QueryInt64("subnet")
	subnet, err := subnetAdmin.Get(ctx, subnetID)
	if err != nil {
		logger.Error("Get subnet failed", err)
		c.HTML(http.StatusBadRequest, err.Error())
		return
	}
	instID := c.QueryInt64("instance")
	if instID > 0 {
		permit, _ := memberShip.CheckOwner(model.Writer, "instances", int64(instID))
		if !permit {
			logger.Error("Not authorized to access instance")
			c.Data["ErrorMsg"] = "Not authorized to access instance"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	address := c.QueryTrim("address")
	mac := c.QueryTrim("mac")
	ifname := c.QueryTrim("ifname")
	sgList := c.QueryTrim("secgroups")
	var sgIDs []int64
	if sgList != "" {
		sg := strings.Split(sgList, ",")
		for i := 0; i < len(sg); i++ {
			sgID, err := strconv.Atoi(sg[i])
			if err != nil {
				logger.Error("Invalid security group ID", err)
				continue
			}
			permit, _ := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
			if !permit {
				logger.Error("Not authorized to access security group")
				c.Data["ErrorMsg"] = "Not authorized to access security group"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			sgIDs = append(sgIDs, int64(sgID))
		}
	} else {
		sgID := store.Get("defsg").(int64)
		permit, _ := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
		if !permit {
			logger.Error("Not authorized to access security group")
			c.Data["ErrorMsg"] = "Not authorized to access security group"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		sgIDs = append(sgIDs, sgID)
	}
	secgroups := []*model.SecurityGroup{}
	if err = DB().Where(sgIDs).Find(&secgroups).Error; err != nil {
		logger.Error("Security group query failed", err)
		return
	}
	iface, err := CreateInterface(ctx, subnet, instID, memberShip.OrgID, -1, 0, 0, address, mac, ifname, "instance", secgroups)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})
	}
	c.JSON(200, iface)
}

func (v *InterfaceView) Delete(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", id)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	iface := &model.Interface{Model: model.Model{ID: id}}
	err = DB().Take(iface).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = DeleteInterface(ctx, iface)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, "ok")
}

func (v *InterfaceView) Patch(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "../instances"
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	ifaceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	iface, err := interfaceAdmin.Get(ctx, int64(ifaceID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	instance, err := instanceAdmin.Get(ctx, iface.Instance)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	name := c.QueryTrim("name")
	sgs := c.QueryStrings("secgroups")
	secgroups := []*model.SecurityGroup{}
	if len(sgs) > 0 {
		for _, sg := range sgs {
			sgID, err := strconv.Atoi(sg)
			if err != nil {
				logger.Debug("Invalid security group ID, %v", err)
				c.Data["ErrorMsg"] = err.Error()
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			secgroup, err := secgroupAdmin.Get(ctx, int64(sgID))
			if err != nil {
				logger.Debug("Failed to query security group, %v", err)
				c.Data["ErrorMsg"] = err.Error()
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			secgroups = append(secgroups, secgroup)
		}
	}
	err = interfaceAdmin.Update(ctx, instance, iface, name, 1000, 1000, secgroups)
	if err != nil {
		logger.Debug("Failed to update interface", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
	}
	c.Redirect(redirectTo)
}
