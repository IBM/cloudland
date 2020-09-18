/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

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
	macaron "gopkg.in/macaron.v1"
)

var (
	orgView  = &OrgView{}
	orgAdmin = &OrgAdmin{}
)

type OrgAdmin struct {
}

type OrgView struct{}

func (a *OrgAdmin) Create(ctx context.Context, name, owner string) (org *model.Organization, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	user := &model.User{Username: owner}
	err = db.Where(user).Take(user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		return
	}
	org = &model.Organization{
		Model: model.Model{Owner: user.ID, Creater: memberShip.UserID},
		Name:  name,
	}
	err = db.Create(org).Error
	if err != nil {
		log.Println("DB failed to create organization ", err)
		return
	}
	secGroup, err := secgroupAdmin.Create(ctx, name, true, org.ID)
	if err != nil {
		log.Println("Failed to create security group", err)
	}
	member := &model.Member{UserID: user.ID, UserName: owner, OrgID: org.ID, OrgName: name, Role: model.Owner}
	err = db.Create(member).Error
	if err != nil {
		log.Println("DB failed to create organization member ", err)
		return
	}
	user.Owner = org.ID
	err = db.Save(user).Error
	if err != nil {
		log.Println("DB failed to update user owner", err)
		return
	}
	org.DefaultSG = secGroup.ID
	err = db.Save(org).Error
	if err != nil {
		log.Println("DB failed to update orgabization default security group", err)
		return
	}
	_, err = subnetAdmin.Create(ctx, name, "", "192.168.127.0", "255.255.255.0", "", "", "", "", "", "", "yes", "", 0, org.ID)
	if err != nil {
		log.Println("Failed to create demo subnet", err)
		err = nil
	}
	return
}

func (a *OrgAdmin) Update(ctx context.Context, orgID int64, members, users []string, roles []model.Role) (org *model.Organization, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	org = &model.Organization{Model: model.Model{ID: orgID}}
	err = db.Set("gorm:auto_preload", true).Take(org).Take(org).Error
	if err != nil {
		log.Println("Failed to query organization", err)
		return
	}
	for _, name := range members {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		found := false
		for _, em := range org.Members {
			if name == em.UserName {
				found = true
				break
			}
		}
		if found == true {
			continue
		}
		user := &model.User{Username: name}
		err = db.Model(user).Where(user).Take(user).Error
		if err != nil || user.ID <= 0 {
			log.Println("Failed to query user", err)
			continue
		}
		member := &model.Member{
			Model:    model.Model{Creater: memberShip.UserID, Owner: orgID},
			UserName: name,
			UserID:   user.ID,
			OrgName:  org.Name,
			OrgID:    orgID,
			Role:     model.Reader,
		}
		err = db.Create(member).Error
		if err != nil {
			log.Println("Failed to create member", err)
			continue
		}
	}
	for _, em := range org.Members {
		found := false
		for _, name := range members {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if name == em.UserName {
				found = true
				break
			}
		}
		if found == true {
			continue
		}
		member := &model.Member{
			UserName: em.UserName,
			OrgID:    orgID,
		}
		err = db.Where(member).Delete(member).Error
		if err != nil {
			log.Println("Failed to delete member", err)
			continue
		}
	}
	for i, user := range users {
		err = db.Model(&model.Member{}).Where("user_name = ? and org_id = ?", user, orgID).Update("role", roles[i]).Error
		if err != nil {
			log.Println("Failed to update member", err)
			continue
		}
	}
	return
}

func (a *OrgAdmin) Get(name string) (org *model.Organization, err error) {
	org = &model.Organization{}
	db := DB()
	err = db.Take(org, &model.Organization{Name: name}).Error
	return
}

func (a *OrgAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	count := 0
	err = db.Model(&model.Interface{}).Where("owner = ?", id).Count(&count).Error
	if err != nil {
		log.Println("DB failed to query interfaces, %v", err)
		return
	}
	if count > 0 {
		log.Println("There are resources in this org", err)
		err = fmt.Errorf("There are resources in this org")
		return
	}
	err = db.Delete(&model.Member{}, `org_id = ?`, id).Error
	if err != nil {
		log.Println("DB failed to delete member, %v", err)
		return
	}
	keys := []*model.Key{}
	err = db.Where("owner = ?", id).Find(&keys).Error
	if err != nil {
		log.Println("DB failed to query keys", err)
		return
	}
	for _, key := range keys {
		err = keyAdmin.Delete(key.ID)
		if err != nil {
			log.Println("Can not delete key", err)
			return
		}
	}
	secgroups := []*model.SecurityGroup{}
	err = db.Where("owner = ?", id).Find(&secgroups).Error
	if err != nil {
		log.Println("DB failed to query security groups", err)
		return
	}
	for _, sg := range secgroups {
		err = secgroupAdmin.Delete(sg.ID)
		if err != nil {
			log.Println("Can not delete security group", err)
			return
		}
	}
	err = db.Delete(&model.Organization{Model: model.Model{ID: id}}).Error
	if err != nil {
		log.Println("DB failed to delete organization, %v", err)
		return
	}
	return
}

func (a *OrgAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, orgs []*model.Organization, err error) {
	memberShip := GetMemberShip(ctx)
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}
	db := DB()
	user := &model.User{Model: model.Model{ID: memberShip.UserID}}
	err = db.Take(user).Error
	if err != nil {
		log.Println("DB failed to query user, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	where := ""
	if memberShip.Role != model.Admin {
		where = fmt.Sprintf("owner = %d", user.ID)
	}
	if err = db.Model(&model.Organization{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	err = db.Where(where).Where(query).Find(&orgs).Error
	if err != nil {
		log.Println("DB failed to query organizations, %v", err)
		return
	}
	return
}

func (v *OrgView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	query := c.QueryTrim("q")
	total, orgs, err := orgAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list organizations, %v", err)
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
	c.Data["Organizations"] = orgs
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"orgs":  orgs,
			"total": total,
			"pages": pages,
			"query": query,
		})
		return
	}
	c.HTML(200, "orgs")
}

func (v *OrgView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
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
	orgID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid organization ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = orgAdmin.Delete(c.Req.Context(), int64(orgID))
	if err != nil {
		log.Println("Failed to delete organization, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "orgs",
	})
	return
}

func (v *OrgView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "orgs_new")
}

func (v *OrgView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	orgID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if memberShip.Role != model.Admin && (memberShip.Role < model.Owner || memberShip.OrgID != int64(orgID)) {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	org := &model.Organization{Model: model.Model{ID: int64(orgID)}}
	if err = db.Preload("Members").Take(org).Error; err != nil {
		log.Println("Organization query failed", err)
		return
	}
	org.OwnerUser = &model.User{Model: model.Model{ID: org.Owner}}
	if err = db.Take(org.OwnerUser).Error; err != nil {
		log.Println("Owner user query failed", err)
		return
	}
	c.Data["Org"] = org
	c.HTML(200, "orgs_patch")
}

func (v *OrgView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	orgID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if memberShip.Role != model.Admin && (memberShip.Role < model.Owner || memberShip.OrgID != int64(orgID)) {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../orgs/" + id
	members := c.QueryTrim("members")
	memberList := strings.Split(members, " ")
	userList := c.QueryStrings("names")
	roles := c.QueryStrings("roles")
	var roleList []model.Role
	for _, r := range roles {
		role, err := strconv.Atoi(r)
		if err != nil {
			log.Println("Failed to convert role", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		if memberShip.Role < model.Role(role) {
			log.Println("Not authorized for this operation")
			c.Data["ErrorMsg"] = "Not authorized for this operation"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		roleList = append(roleList, model.Role(role))
	}
	org, err := orgAdmin.Update(c.Req.Context(), int64(orgID), memberList, userList, roleList)
	if err != nil {
		log.Println("Failed to update organization, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, org)
		return
	}
	c.Redirect(redirectTo)
}

func (v *OrgView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../orgs"
	name := c.QueryTrim("orgname")
	owner := c.QueryTrim("owner")
	organization, err := orgAdmin.Create(c.Req.Context(), name, owner)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, err.Error())
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, organization)
		return
	}
	c.Redirect(redirectTo)
}
