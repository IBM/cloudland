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

func (a *SecgroupAdmin) Create(ctx context.Context, name string, isDefault bool) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Name: name, IsDefault: isDefault}
	err = db.Create(secgroup).Error
	if err != nil {
		log.Println("DB failed to create security group, %v", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, "0.0.0.0/0", "egress", "tcp", 1, 65535)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, "0.0.0.0/0", "egress", "udp", 1, 65535)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, "0.0.0.0/0", "ingress", "tcp", 22, 22)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, "0.0.0.0/0", "egress", "icmp", -1, -1)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, "0.0.0.0/0", "ingress", "icmp", -1, -1)
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

func (a *SecgroupAdmin) List(offset, limit int64, order string) (total int64, secgroups []*model.SecurityGroup, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	where := memberShip.GetWhere()
	secgroups = []*model.SecurityGroup{}
	if err = db.Model(&model.SecurityGroup{}).Where(where).Count(&total).Error; err != nil {
		log.Println("DB failed to count security group(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Find(&secgroups).Error; err != nil {
		log.Println("DB failed to query security group(s), %v", err)
		return
	}

	return
}

func (v *SecgroupView) List(c *macaron.Context, store session.Store) {
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
	total, secgroups, err := secgroupAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list security group(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["SecurityGroups"] = secgroups
	c.Data["Total"] = total
	c.HTML(200, "secgroups")
}

func (v *SecgroupView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	secgroupID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid security group ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "security_groups", int64(secgroupID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = secgroupAdmin.Delete(int64(secgroupID))
	if err != nil {
		log.Println("Failed to delete security group, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "secgroups",
	})
	return
}

func (v *SecgroupView) New(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "secgroups_new")
}

func (v *SecgroupView) Create(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../secgroups"
	name := c.Query("name")
	isdefStr := c.Query("isdefault")
	isDef := false
	if isdefStr == "" || isdefStr == "no" {
		isDef = false
	} else if isdefStr == "yes" {
		isDef = true
	}

	_, err := secgroupAdmin.Create(c.Req.Context(), name, isDef)
	if err != nil {
		log.Println("Failed to create security group, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
