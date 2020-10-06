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

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	secgroupAdmin = &SecgroupAdmin{}
	secgroupView  = &SecgroupView{}
)

type SecgroupAdmin struct{}
type SecgroupView struct{}

func (a *SecgroupAdmin) Switch(ctx context.Context, newSg *model.SecurityGroup, store session.Store) (err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	org := &model.Organization{Model: model.Model{ID: memberShip.OrgID}}
	err = db.Take(org).Error
	if err != nil {
		log.Println("Failed to query organization", err)
	}
	oldSg := &model.SecurityGroup{Model: model.Model{ID: org.DefaultSG}}
	err = db.Take(oldSg).Error
	if err != nil {
		log.Println("Failed to query default security group", err)
	}
	oldSg.IsDefault = false
	err = db.Save(oldSg).Error
	if err != nil {
		log.Println("Failed to save old security group", err)
	}
	org.DefaultSG = newSg.ID
	err = db.Save(org).Error
	if err != nil {
		log.Println("Failed to save organization", err)
	}
	newSg.IsDefault = true
	err = db.Save(newSg).Error
	if err != nil {
		log.Println("Failed to save new security group", err)
	}
	store.Set("defsg", org.DefaultSG)
	return
}

func (a *SecgroupAdmin) Update(ctx context.Context, sgID int64, name string, isDefault bool) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{ID: sgID}}
	err = db.Take(secgroup).Error
	if err != nil {
		log.Println("Failed to query security group", err)
		return
	}
	if name != "" && secgroup.Name != name {
		secgroup.Name = name
	}
	if isDefault && secgroup.IsDefault != isDefault {
		secgroup.IsDefault = isDefault
	}
	err = db.Save(secgroup).Error
	if err != nil {
		log.Println("Failed to save security group", err)
		return
	}
	return
}

func (a *SecgroupAdmin) Create(ctx context.Context, name string, isDefault bool, owner int64) (secgroup *model.SecurityGroup, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{Creater: memberShip.UserID, Owner: owner}, Name: name, IsDefault: isDefault}
	err = db.Create(secgroup).Error
	if err != nil {
		log.Println("DB failed to create security group, %v", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "egress", "tcp", 1, 65535)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "egress", "udp", 1, 65535)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 22, 22)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "udp", 68, 68)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "egress", "icmp", -1, -1)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "icmp", -1, -1)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	return
}

func (a *SecgroupAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	secgroup := &model.SecurityGroup{Model: model.Model{ID: id}}
	err = db.Model(secgroup).Related(&secgroup.Interfaces, "Interfaces").Error
	if err != nil {
		log.Println("DB failed to query security group", err)
		return
	}
	if len(secgroup.Interfaces) > 0 {
		log.Println("Security group has associated interfaces")
		err = fmt.Errorf("Security group has associated interfaces")
		return
	}
	err = db.Where("secgroup = ?", id).Delete(&model.SecurityRule{}).Error
	if err != nil {
		log.Println("DB failed to delete security group rules", err)
	}
	if err = db.Delete(&model.SecurityGroup{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("DB failed to delete security group", err)
		return
	}
	return
}

func (a *SecgroupAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, secgroups []*model.SecurityGroup, err error) {
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
	secgroups = []*model.SecurityGroup{}
	if err = db.Model(&model.SecurityGroup{}).Where(where).Where(query).Count(&total).Error; err != nil {
		log.Println("DB failed to count security group(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Where(query).Find(&secgroups).Error; err != nil {
		log.Println("DB failed to query security group(s), %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, sg := range secgroups {
			sg.OwnerInfo = &model.Organization{Model: model.Model{ID: sg.Owner}}
			if err = db.Take(sg.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				err = nil
				continue
			}
		}
	}

	return
}

func (v *SecgroupView) List(c *macaron.Context, store session.Store) {
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
	total, secgroups, err := secgroupAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list security group(s), %v", err)
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
	c.Data["SecurityGroups"] = secgroups
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"secgroups": secgroups,
			"total":     total,
			"pages":     pages,
			"query":     query,
		})
		return
	}
	c.HTML(200, "secgroups")
}

func (v *SecgroupView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroupID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid security group ID, %v", err)
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
	err = secgroupAdmin.Delete(int64(secgroupID))
	if err != nil {
		log.Println("Failed to delete security group, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "secgroups",
	})
	return
}

func (v *SecgroupView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "secgroups_new")
}

func (v *SecgroupView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params(":id")
	sgID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroup := &model.SecurityGroup{Model: model.Model{ID: int64(sgID)}}
	err = db.Take(secgroup).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["Secgroup"] = secgroup
	c.HTML(200, "secgroups_patch")
}

func (v *SecgroupView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../secgroups"
	id := c.Params(":id")
	name := c.QueryTrim("name")
	sgID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	isdefStr := c.QueryTrim("isdefault")
	isDef := false
	if isdefStr == "" || isdefStr == "no" {
		isDef = false
	} else if isdefStr == "yes" {
		isDef = true
	}
	secgroup, err := secgroupAdmin.Update(c.Req.Context(), int64(sgID), name, isDef)
	if err != nil {
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(500, err.Error())
		return
	}
	if isDef {
		err = secgroupAdmin.Switch(c.Req.Context(), secgroup, store)
		if err != nil {
			log.Println("Failed to switch security group", err)
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
	}
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, secgroup)
		return
	}
	c.Redirect(redirectTo)
	return
}

func (v *SecgroupView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../secgroups"
	name := c.QueryTrim("name")
	isdefStr := c.QueryTrim("isdefault")
	isDef := false
	if isdefStr == "" || isdefStr == "no" {
		isDef = false
	} else if isdefStr == "yes" {
		isDef = true
	}

	secgroup, err := secgroupAdmin.Create(c.Req.Context(), name, isDef, memberShip.OrgID)
	if err != nil {
		log.Println("Failed to create security group, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	if isDef {
		err = secgroupAdmin.Switch(c.Req.Context(), secgroup, store)
		if err != nil {
			log.Println("Failed to switch security group", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(500, "500")
			return
		}
	}
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, secgroup)
		return
	}
	c.Redirect(redirectTo)
}
