/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"encoding/json"
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
	secruleAdmin = &SecruleAdmin{}
	secruleView  = &SecruleView{}
)

type SecruleAdmin struct{}
type SecruleView struct{}

func (a *SecruleAdmin) ApplySecgroup(ctx context.Context, secgroup *model.SecurityGroup) (err error) {
	ctx, db := GetContextDB(ctx)
	err = secgroupAdmin.GetSecgroupInterfaces(ctx, secgroup)
	if err != nil {
		logger.Error("DB failed to get security group related interfaces", err)
		return
	}
	for _, iface := range secgroup.Interfaces {
		var securityData []*SecurityData
		err = secgroupAdmin.GetInterfaceSecgroups(ctx, iface)
		if err != nil {
			logger.Error("DB failed to get interface related security groups, %v", err)
			continue
		}
		securityData, err = GetSecurityData(ctx, iface.SecurityGroups)
		if err != nil {
			logger.Error("DB failed to get security data, %v", err)
			continue
		}
		var jsonData []byte
		jsonData, err = json.Marshal(securityData)
		if err != nil {
			logger.Error("Failed to marshal security json data, %v", err)
			continue
		}
		logger.Debugf("iface: %+v", iface)
		instance := &model.Instance{Model: model.Model{ID: iface.Instance}}
		err = db.Take(instance).Error
		if err != nil {
			logger.Error("DB failed to get instance, %v", err)
			continue
		}
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/reapply_secgroup.sh '%s' '%s' '%t'<<EOF\n%s\nEOF", iface.Address.Address, iface.MacAddr, iface.AllowSpoofing, jsonData)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Reapply security groups execution failed, %v", err)
			return
		}
	}
	return
}

func (a *SecruleAdmin) Update(ctx context.Context, id int64, remoteIp, direction, protocol string, portMin, portMax int) (secrule *model.SecurityRule, err error) {
	db := DB()
	//secrule = &model.SecurityRule{Model: model.Model{ID: id}}
	secrules := &model.SecurityRule{Model: model.Model{ID: id}}
	err = db.Take(secrules).Error
	if err != nil {
		logger.Error("DB failed to query security rules ", err)
		return
	}
	//remoteip
	if remoteIp != "" {
		netLen := strings.Split(remoteIp, "/")
		NetLen, _ := strconv.Atoi(netLen[1])
		if NetLen < 0 || NetLen > 32 {
			logger.Error("Invalid Netmask,fill in valid one")
			err = fmt.Errorf("Invalid Netmask for RemoteIp, please fill a valid one")
			return
		}
		secrules.RemoteIp = remoteIp
	}
	//direction
	if direction != "" {
		secrules.Direction = direction
	}
	if protocol != "" {
		secrules.Protocol = protocol
	}
	if portMin <= portMax {
		if portMin > 0 && portMin < 65536 {
			secrules.PortMin = int32(portMin)
			if portMax > 0 && portMax < 65536 {
				secrules.PortMax = int32(portMax)
			} else if portMax > 65535 {
				logger.Error("it's out of range, please input less than 65536")
				err = fmt.Errorf("it's invalid port for PortMax, please fill a valid port")
				return
			} else {
				secrules.PortMax = -1
			}
		} else if portMin < -1 || portMin == 0 {
			logger.Error("it's out of range,please fill a valid port")
			err = fmt.Errorf("it's invalid port for PortMin, please fill a valid port")
			return
		} else if portMin > 65535 {
			logger.Error("it's out of range, please input less than 65537")
			err = fmt.Errorf("it's invalid port for PortMin, please fill a valid port")
			return
		} else {
			secrules.PortMin = -1
		}

	} else {
		logger.Error("PortMax should be greater than PortMin")
		err = fmt.Errorf("PortMax should be greater than PortMin")
		return
	}
	err = db.Model(secrule).Updates(secrules).Error
	if err != nil {
		logger.Error("DB failed to save sucurity rule ", err)
		return
	}
	return

}

func (a *SecruleAdmin) Create(ctx context.Context, remoteIp, direction, protocol string, portMin, portMax int32, secgroup *model.SecurityGroup) (secrule *model.SecurityRule, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, secgroup.Owner)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	if protocol == "icmp" {
		portMin = -1
		portMax = -1
	}
	secrule = &model.SecurityRule{
		Model:     model.Model{Creater: memberShip.UserID},
		Owner:     memberShip.OrgID,
		Secgroup:  secgroup.ID,
		RemoteIp:  remoteIp,
		Direction: direction,
		IpVersion: "ipv4",
		Protocol:  protocol,
		PortMin:   portMin,
		PortMax:   portMax,
	}
	err = db.Where(secrule).Take(secrule).Error
	if err == nil {
		logger.Error("Security rule already exists")
		return
	}
	err = db.Create(secrule).Error
	if err != nil {
		logger.Error("DB failed to create security rule", err)
		return
	}
	err = a.ApplySecgroup(ctx, secgroup)
	if err != nil {
		logger.Error("Failed to apply security rule", err)
		return
	}
	return
}

func (a *SecruleAdmin) DeleteRule(ctx context.Context, remoteIp, direction, protocol string, portMin, portMax int32, secgroup *model.SecurityGroup) (secrule *model.SecurityRule, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, secgroup.Owner)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	secrule = &model.SecurityRule{
		Secgroup:  secgroup.ID,
		RemoteIp:  remoteIp,
		Direction: direction,
		IpVersion: "ipv4",
		Protocol:  protocol,
		PortMin:   portMin,
		PortMax:   portMax,
	}
	err = db.Where(secrule).Take(secrule).Error
	if err != nil {
		logger.Error("Failed to query secrule", err)
		return
	}
	err = db.Delete(secrule).Error
	if err != nil {
		logger.Error("DB failed to delete security rule", err)
		return
	}
	err = a.ApplySecgroup(ctx, secgroup)
	if err != nil {
		logger.Error("Failed to apply security rule", err)
		return
	}
	return
}

func (a *SecruleAdmin) Delete(ctx context.Context, secrule *model.SecurityRule, secgroup *model.SecurityGroup) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, secrule.Owner)
	if !permit {
		logger.Error("Not authorized to delete the router")
		err = fmt.Errorf("Not authorized")
		return
	}
	if err = db.Delete(secrule).Error; err != nil {
		logger.Error("DB failed to delete security rule, %v", err)
		return
	}
	err = a.ApplySecgroup(ctx, secgroup)
	if err != nil {
		logger.Error("Failed to apply security rule", err)
		return
	}
	return
}

func (a *SecruleAdmin) List(ctx context.Context, offset, limit int64, order string, secgroup *model.SecurityGroup) (total int64, secrules []*model.SecurityRule, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Reader, secgroup.Owner)
	if !permit {
		logger.Error("Not authorized for this operation")
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

	where := fmt.Sprintf("secgroup = %d", secgroup.ID)
	wm := memberShip.GetWhere()
	if wm != "" {
		where = fmt.Sprintf("%s and %s", where, wm)
	}
	secrules = []*model.SecurityRule{}
	if err = db.Model(&model.SecurityRule{}).Where(where).Count(&total).Error; err != nil {
		logger.Error("DB failed to count security rule(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Find(&secrules).Error; err != nil {
		logger.Error("DB failed to query security rule(s), %v", err)
		return
	}

	return
}

func (v *SecruleView) List(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	sgid := c.Params("sgid")
	if sgid == "" {
		logger.Error("Security group ID is empty")
		c.Data["ErrorMsg"] = "Security group ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		logger.Error("Invalid security group ID", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroup, err := secgroupAdmin.Get(ctx, int64(secgroupID))
	if err != nil {
		logger.Error("Failed to get security group", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	total, secrules, err := secruleAdmin.List(ctx, offset, limit, order, secgroup)
	if err != nil {
		logger.Error("Failed to list security rule(s)", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["SecurityRules"] = secrules
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.HTML(200, "secrules")
}

func (v *SecruleView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	sgid := c.Params("sgid")
	if sgid == "" {
		logger.Error("Security group ID is empty")
		c.Data["ErrorMsg"] = "Security group ID is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		logger.Error("Invalid security group ID", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	secgroup, err := secgroupAdmin.Get(ctx, int64(secgroupID))
	if err != nil {
		logger.Error("Failed to get security group", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	id := c.Params("id")
	if id == "" {
		logger.Error("ID is empty, %v", err)
		c.Data["ErrorMsg"] = "ID is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid security rule ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	secrule, err := secruleAdmin.Get(ctx, int64(secruleID), secgroup)
	if err != nil {
		logger.Error("Failed to get security rule", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = secruleAdmin.Delete(c.Req.Context(), secrule, secgroup)
	if err != nil {
		logger.Error("Failed to delete security rule, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "secrules",
	})
	return
}

func (v *SecruleView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "secrules_new")
}

func (v *SecruleView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "../secrules"
	remoteIp := c.QueryTrim("remoteip")
	sgid := c.Params("sgid")
	if sgid == "" {
		logger.Error("Security group ID is empty")
		c.Data["ErrorMsg"] = "Security group ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		logger.Error("Invalid security group ID", err)
		c.Data["Error:Msg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	direction := c.QueryTrim("direction")
	protocol := c.QueryTrim("protocol")
	min := c.QueryTrim("portmin")
	max := c.QueryTrim("portmax")
	portMin, _ := strconv.Atoi(min)
	portMax, _ := strconv.Atoi(max)
	secgroup, err := secgroupAdmin.Get(ctx, int64(secgroupID))
	if err != nil {
		logger.Error("Failed to get security group", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, err = secruleAdmin.Create(ctx, remoteIp, direction, protocol, int32(portMin), int32(portMax), secgroup)
	if err != nil {
		logger.Error("Failed to create security rule, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}

func (a *SecruleAdmin) Get(ctx context.Context, id int64, secgroup *model.SecurityGroup) (secrule *model.SecurityRule, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid security rule ID: %d", id)
		logger.Error(err)
		return
	}
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	db := DB()
	secrule = &model.SecurityRule{Model: model.Model{ID: id}}
	err = db.Where(where).Take(secrule).Error
	if err != nil {
		logger.Error("Failed to query secrule", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, secrule.Owner)
	if !permit {
		logger.Error("Not authorized to get security group")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *SecruleAdmin) GetSecruleByUUID(ctx context.Context, uuID string, secgroup *model.SecurityGroup) (secrule *model.SecurityRule, err error) {
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	db := DB()
	secrule = &model.SecurityRule{}
	err = db.Where(where).Where("uuid = ? and secgroup = ?", uuID, secgroup.ID).Take(secrule).Error
	if err != nil {
		logger.Error("Failed to query secrule", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, secrule.Owner)
	if !permit {
		logger.Error("Not authorized to get security group")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (v *SecruleView) Edit(c *macaron.Context, store session.Store) {
	db := DB()
	id := c.Params("id")
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Security Rule ID is empty")
		c.Data["ErrorMsg"] = "Security Rule ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secrules := &model.SecurityRule{Model: model.Model{ID: int64(secruleID)}}
	err = db.Take(secrules).Error
	if err != nil {
		logger.Error("Database failed to query security rules", err)
		return
	}
	c.Data["Secrules"] = secrules
	logger.Debugf("Edit security rules: %+v", secrules)
	c.HTML(200, "secrules_patch")
}

func (v *SecruleView) Patch(c *macaron.Context, store session.Store) {
	redirectTo := "../secrules"
	id := c.Params("id")
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid secure rule ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	remoteIp := c.QueryTrim("remoteip")
	direction := c.QueryTrim("direction")
	protocol := c.QueryTrim("protocol")
	min := c.QueryTrim("portmin")
	max := c.QueryTrim("portmax")
	portMin, err := strconv.Atoi(min)
	portMax, err := strconv.Atoi(max)
	_, err = secruleAdmin.Update(c.Req.Context(), int64(secruleID), remoteIp, direction, protocol, portMin, portMax)
	if err != nil {
		logger.Error("Create Security Rules failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)

}
