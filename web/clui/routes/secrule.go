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
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/reapply_secgroup.sh '%s' '%s' <<EOF\n%s\nEOF", iface.Address.Address, iface.MacAddr, jsonData)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed", err)
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
		log.Println("DB failed to query security rules ", err)
		return
	}
	//remoteip
	if remoteIp != "" {
		netLen := strings.Split(remoteIp, "/")
		NetLen, _ := strconv.Atoi(netLen[1])
		if NetLen < 0 || NetLen > 32 {
			log.Println("Invalid Netmask,fill in valid one")
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
				log.Println("it's out of range, please input less than 65536")
				err = fmt.Errorf("it's invalid port for PortMax, please fill a valid port")
				return
			} else {
				secrules.PortMax = -1
			}
		} else if portMin < -1 || portMin == 0 {
			log.Println("it's out of range,please fill a valid port")
			err = fmt.Errorf("it's invalid port for PortMin, please fill a valid port")
			return
		} else if portMin > 65535 {
			log.Println("it's out of range, please input less than 65537")
			err = fmt.Errorf("it's invalid port for PortMin, please fill a valid port")
			return
		} else {
			secrules.PortMin = -1
		}

	} else {
		log.Println("PortMax should be greater than PortMin")
		err = fmt.Errorf("PortMax should be greater than PortMin")
		return
	}
	err = db.Save(secrules).Error
	if err != nil {
		log.Println("DB failed to save sucurity rule ", err)
		return
	}
	return

}

func (a *SecruleAdmin) Create(ctx context.Context, sgID, owner int64, remoteIp, direction, protocol string, portMin, portMax int) (secrule *model.SecurityRule, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	secgroup := &model.SecurityGroup{Model: model.Model{ID: sgID}}
	err = db.Model(secgroup).Preload("Address").Related(&secgroup.Interfaces, "Interfaces").Error
	if err != nil {
		log.Println("DB failed to query security group", err)
		return
	}
	secrule = &model.SecurityRule{
		Model:     model.Model{Creater: memberShip.UserID, Owner: owner},
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

func (a *SecruleAdmin) List(ctx context.Context, offset, limit int64, order string, secgroupID int64) (total int64, secrules []*model.SecurityRule, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	where := fmt.Sprintf("secgroup = %d", secgroupID)
	wm := memberShip.GetWhere()
	if wm != "" {
		where = fmt.Sprintf("%s and %s", where, wm)
	}
	secrules = []*model.SecurityRule{}
	if err = db.Model(&model.SecurityRule{}).Where(where).Count(&total).Error; err != nil {
		log.Println("DB failed to count security rule(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Find(&secrules).Error; err != nil {
		log.Println("DB failed to query security rule(s), %v", err)
		return
	}

	return
}

func (v *SecruleView) List(c *macaron.Context, store session.Store) {
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
	sgid := c.Params("sgid")
	if sgid == "" {
		log.Println("Security group ID is empty")
		c.Data["ErrorMsg"] = "Security group ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		log.Println("Invalid security group ID", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	total, secrules, err := secruleAdmin.List(c.Req.Context(), offset, limit, order, int64(secgroupID))
	if err != nil {
		log.Println("Failed to list security rule(s)", err)
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
	c.Data["SecurityRules"] = secrules
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"secrules": secrules,
			"total":    total,
			"pages":    pages,
		})
		return
	}
	c.HTML(200, "secrules")
}

func (v *SecruleView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	sgid := c.Params("sgid")
	if sgid == "" {
		log.Println("Security group ID is empty")
		c.Data["ErrorMsg"] = "Security group ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		log.Println("Invalid security group ID", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	id := c.Params("id")
	if id == "" {
		log.Println("ID is empty, %v", err)
		c.Data["ErrorMsg"] = "ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid security rule ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "security_groups", int64(secgroupID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err = memberShip.CheckOwner(model.Writer, "security_rules", int64(secruleID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = secruleAdmin.Delete(c.Req.Context(), int64(secgroupID), int64(secruleID))
	if err != nil {
		log.Println("Failed to delete security rule, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
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
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "secrules_new")
}

func (v *SecruleView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../secrules"
	remoteIp := c.QueryTrim("remoteip")
	sgid := c.Params("sgid")
	if sgid == "" {
		log.Println("Security group ID is empty")
		c.Data["ErrorMsg"] = "Security group ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroupID, err := strconv.Atoi(sgid)
	if err != nil {
		log.Println("Invalid security group ID", err)
		c.Data["Error:Msg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	direction := c.QueryTrim("direction")
	protocol := c.QueryTrim("protocol")
	min := c.QueryTrim("portmin")
	max := c.QueryTrim("portmax")
	portMin, err := strconv.Atoi(min)
	portMax, err := strconv.Atoi(max)
	secrule, err := secruleAdmin.Create(c.Req.Context(), int64(secgroupID), memberShip.OrgID, remoteIp, direction, protocol, portMin, portMax)
	if err != nil {
		log.Println("Failed to create security rule, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, err.Error())
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, secrule)
		return
	}
	c.Redirect(redirectTo)
}
func (v *SecruleView) Edit(c *macaron.Context, store session.Store) {
	db := DB()
	id := c.Params("id")
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Security Rule ID is empty")
		c.Data["ErrorMsg"] = "Security Rule ID is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secrules := &model.SecurityRule{Model: model.Model{ID: int64(secruleID)}}
	err = db.Take(secrules).Error
	if err != nil {
		log.Println("Database failed to query security rules", err)
		return
	}
	c.Data["Secrules"] = secrules
	log.Println(secrules)
	c.HTML(200, "secrules_patch")
}
func (v *SecruleView) Patch(c *macaron.Context, store session.Store) {
	redirectTo := "../secrules"
	id := c.Params("id")
	secruleID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid secure rule ID, %v", err)
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
	secrule, err := secruleAdmin.Update(c.Req.Context(), int64(secruleID), remoteIp, direction, protocol, portMin, portMax)
	if err != nil {
		log.Println("Create Security Rules failed, %v", err)
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
		c.JSON(200, secrule)
		return
	}
	c.Redirect(redirectTo)

}
