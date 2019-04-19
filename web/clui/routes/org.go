package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
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

func (a *OrgAdmin) Create(name, owner string) (org *model.Organization, err error) {
	db := DB()
	user := &model.User{Username: owner}
	err = db.Model(user).Take(user).Error
	org = &model.Organization{
		Name:  name,
		Owner: user.ID,
	}
	err = db.Create(org).Error
	if err != nil {
		log.Println("DB failed to create organization, %v", err)
		return
	}
	member := &model.Member{UserID: user.ID, UserName: owner, OrgID: org.ID, OrgName: name, Role: model.Owner}
	err = db.Create(member).Error
	if err != nil {
		log.Println("DB failed to create organization member, %v", err)
		return
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

func (a *OrgAdmin) List(owner string) (orgs []*model.Organization, err error) {
	db := DB()
	user := &model.User{Username: owner}
	err = db.Model(user).Take(user).Error
	if err != nil {
		log.Println("DB failed to query user, %v", err)
		return
	}
	members := []*model.Member{}
	err = db.Find(&members, &model.Member{UserID: user.ID}).Error
	if err != nil {
		log.Println("DB failed to query members, %v", err)
		return
	}
	orgs = []*model.Organization{}
	where := ""
	for i, member := range members {
		if i == 0 {
			where = fmt.Sprintf("id = %d", member.OrgID)
		} else {
			where = fmt.Sprintf("%s or id = %d", where, member.OrgID)
		}
	}
	err = db.Where(where).Find(&orgs).Error
	if err != nil {
		log.Println("DB failed to query organizations, %v", err)
		return
	}
	return
}

func (v *OrgView) List(c *macaron.Context, store session.Store) {
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	orgs, err := orgAdmin.List(order)
	if err != nil {
		log.Println("Failed to list organizations, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Organizations"] = orgs
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

func (v *OrgView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../orgs"
	name := c.Query("name")
	owner := c.Query("owner")
	_, err := orgAdmin.Create(name, owner)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
