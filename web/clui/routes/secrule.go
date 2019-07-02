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
	secruleAdmin = &SecruleAdmin{}
	secruleView  = &SecruleView{}
)

type SecruleAdmin struct{}
type SecruleView struct{}

func (a *SecruleAdmin) ApplySecgroup(ctx context.Context, secgroup *model.SecurityGroup, ruleID int64) (err error) {
	db := DB()
	if len(secgroup.Interfaces) <= 0 {
		return
	}
	securityData := []*SecurityData{}
	secrules := []*model.SecurityRule{}
	err = db.Model(&model.SecurityRule{}).Where("secgroup = ?", secgroup.ID).Find(&secrules).Error
	if err != nil {
		log.Println("DB failed to query security rules", err)
		return
	}
	for _, rule := range secrules {
		if rule.ID == ruleID {
			continue
		}
		sgr := &SecurityData{
			Secgroup:    rule.Secgroup,
			RemoteIp:    rule.RemoteIp,
			RemoteGroup: rule.RemoteGroup,
			Direction:   rule.Direction,
			IpVersion:   rule.IpVersion,
			Protocol:    rule.Protocol,
			PortMin:     rule.PortMin,
			PortMax:     rule.PortMax,
		}
		securityData = append(securityData, sgr)
	}
	jsonData, err := json.Marshal(securityData)
	if err != nil {
		log.Println("Failed to marshal instance json data, %v", err)
		return
	}
	for _, iface := range secgroup.Interfaces {
		inst := &model.Instance{Model: model.Model{ID: iface.Instance}}
		err = db.Take(inst).Error
		if err != nil {
			log.Println("DB failed to create security rule", err)
			continue
		}
		control := fmt.Sprintf("inter=%d", inst.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/reapply_secgroup.sh %s %s <<EOF\n%s\nEOF", iface.Address.Address, iface.MacAddr, jsonData)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed", err)
			return
		}
	}
	return
}

func (a *SecruleAdmin) Create(ctx context.Context, sgID int64, remoteIp, direction, protocol string, portMin, portMax int) (secrule *model.SecurityRule, err error) {
	db := DB()
	secgroup := &model.SecurityGroup{Model: model.Model{ID: sgID}}
	err = db.Model(secgroup).Preload("Address").Related(&secgroup.Interfaces, "Interfaces").Error
	if err != nil {
		log.Println("DB failed to query security group", err)
		return
	}
	secrule = &model.SecurityRule{
		Secgroup:  sgID,
		RemoteIp:  remoteIp,
		Direction: direction,
		IpVersion: "ipv4",
		Protocol:  protocol,
		PortMin:   int32(portMin),
		PortMax:   int32(portMax),
	}
	err = db.Create(secrule).Error
	if err != nil {
		log.Println("DB failed to create security rule", err)
		return
	}
	err = a.ApplySecgroup(ctx, secgroup, 0)
	if err != nil {
		log.Println("Failed to apply security rule", err)
		return
	}
	return
}

func (a *SecruleAdmin) Delete(ctx context.Context, sgID, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	secgroup := &model.SecurityGroup{Model: model.Model{ID: sgID}}
	err = db.Model(secgroup).Preload("Address").Related(&secgroup.Interfaces, "Interfaces").Error
	if err != nil {
		log.Println("DB failed to query security group", err)
		return
	}
	if err = db.Delete(&model.SecurityRule{Model: model.Model{ID: id}, Secgroup: sgID}).Error; err != nil {
		log.Println("DB failed to delete security rule, %v", err)
		return
	}
	err = a.ApplySecgroup(ctx, secgroup, id)
	if err != nil {
		log.Println("Failed to apply security rule", err)
		return
	}
	return
}

func (a *SecruleAdmin) List(offset, limit int64, order string, secgroupID int64) (total int64, secrules []*model.SecurityRule, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	secrules = []*model.SecurityRule{}
	if err = db.Model(&model.SecurityRule{}).Where("secgroup = ?", secgroupID).Count(&total).Error; err != nil {
		log.Println("DB failed to count security rule(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where("secgroup = ?", secgroupID).Find(&secrules).Error; err != nil {
		log.Println("DB failed to query security rule(s), %v", err)
		return
	}

	return
}

func (v *SecruleView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	sgid := c.Params("sgid")
	if sgid == "" {
		log.Println("Security group ID is empty")
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		log.Println("Invalid security group ID", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	total, secrules, err := secruleAdmin.List(offset, limit, order, int64(secgroupID))
	if err != nil {
		log.Println("Failed to list security rule(s)", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["SecurityRules"] = secrules
	c.Data["Total"] = total
	c.HTML(200, "secrules")
}

func (v *SecruleView) Delete(c *macaron.Context, store session.Store) (err error) {
	sgid := c.Params("sgid")
	if sgid == "" {
		log.Println("Security group ID is empty")
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		log.Println("Invalid security group ID", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	id := c.Params("id")
	if id == "" {
		log.Println("ID is empty, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid security rule ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = secruleAdmin.Delete(c.Req.Context(), int64(secgroupID), int64(secruleID))
	if err != nil {
		log.Println("Failed to delete security rule, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "secrules",
	})
	return
}

func (v *SecruleView) New(c *macaron.Context, store session.Store) {
	c.HTML(200, "secrules_new")
}

func (v *SecruleView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../secrules"
	remoteIp := c.Query("remoteip")
	sgid := c.Params("sgid")
	if sgid == "" {
		log.Println("Security group ID is empty")
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		log.Println("Invalid security group ID", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	direction := c.Query("direction")
	protocol := c.Query("protocol")
	min := c.Query("portmin")
	max := c.Query("portmax")
	portMin, err := strconv.Atoi(min)
	portMax, err := strconv.Atoi(max)
	_, err = secruleAdmin.Create(c.Req.Context(), int64(secgroupID), remoteIp, direction, protocol, portMin, portMax)
	if err != nil {
		log.Println("Failed to create security rule, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
