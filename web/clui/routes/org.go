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
	db := DB()
	user := &model.User{Username: owner}
	err = db.Model(user).Take(user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		return
	}
	sgName := name + "-default"
	secGroup, err := secgroupAdmin.Create(ctx, sgName, true)
	if err != nil {
		log.Println("Failed to create security group", err)
	}
	org = &model.Organization{
		Model:     model.Model{Owner: user.ID},
		Name:      name,
		DefaultSG: secGroup.ID,
	}
	err = db.Create(org).Error
	if err != nil {
		log.Println("DB failed to create organization ", err)
		return
	}
	member := &model.Member{UserID: user.ID, UserName: owner, OrgID: org.ID, OrgName: name, Role: model.Owner}
	err = db.Create(member).Error
	if err != nil {
		log.Println("DB failed to create organization member ", err)
		return
	}
	return
}

func (a *OrgAdmin) Update(ctx context.Context, orgID int64, members, users, roles []string) (org *model.Organization, err error) {
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
		err = db.Model(user).Take(user).Error
		if err != nil {
			log.Println("Failed to query user", err)
			continue
		}
		if user.ID <= 0 {
			continue
		}
		member := &model.Member{
			Model:    model.Model{Owner: orgID},
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
		role, err := strconv.Atoi(roles[i])
		if err != nil {
			log.Println("Failed to convert role", err)
			continue
		}
		err = db.Model(&model.Member{}).Where("user_name = ? and org_id = ?", user, orgID).Update("role", role).Error
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

func (a *OrgAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = db.Delete(&model.Member{}, `org_id = ?`, id).Error
	if err != nil {
		log.Println("DB failed to delete member, %v", err)
		return
	}
	err = db.Delete(&model.Organization{Model: model.Model{ID: id}}).Error
	if err != nil {
		log.Println("DB failed to delete organization, %v", err)
		return
	}
	return
}

func (a *OrgAdmin) List(offset, limit int64, order string, owner string) (total int64, orgs []*model.Organization, err error) {
	db := DB()
	user := &model.User{Username: owner}
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	err = db.Model(user).Where(user).Take(user).Error
	if err != nil {
		log.Println("DB failed to query user, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	where := ""
	if user.Username != "admin" {
		where = fmt.Sprintf("owner = %d", user.ID)
	}
	if err = db.Model(&model.Organization{}).Where(where).Count(&total).Error; err != nil {
		return
	}
	err = db.Where(where).Find(&orgs).Error
	if err != nil {
		log.Println("DB failed to query organizations, %v", err)
		return
	}
	return
}

func (v *OrgView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	owner := store.Get("login").(string)
	total, orgs, err := orgAdmin.List(offset, limit, order, owner)
	if err != nil {
		log.Println("Failed to list organizations, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Organizations"] = orgs
	c.Data["Total"] = total
	c.HTML(200, "orgs")
}

func (v *OrgView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		log.Println("ID is empty, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	orgID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid organization ID, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = orgAdmin.Delete(int64(orgID))
	if err != nil {
		log.Println("Failed to delete organization, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "orgs",
	})
	return
}

func (v *OrgView) New(c *macaron.Context, store session.Store) {
	c.HTML(200, "orgs_new")
}

func (v *OrgView) Edit(c *macaron.Context, store session.Store) {
	db := DB()
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	orgID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	org := &model.Organization{Model: model.Model{ID: int64(orgID)}}
	if err = db.Set("gorm:auto_preload", true).Take(org).Error; err != nil {
		log.Println("Image query failed", err)
		return
	}
	c.Data["Org"] = org
	c.HTML(200, "orgs_patch")
}

func (v *OrgView) Patch(c *macaron.Context, store session.Store) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	orgID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../orgs/" + id
	members := c.Query("members")
	memberList := strings.Split(members, " ")
	userList := c.QueryStrings("names")
	roleList := c.QueryStrings("roles")
	_, err = orgAdmin.Update(c.Req.Context(), int64(orgID), memberList, userList, roleList)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}

func (v *OrgView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../orgs"
	name := c.Query("name")
	owner := c.Query("owner")
	_, err := orgAdmin.Create(c.Req.Context(), name, owner)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
