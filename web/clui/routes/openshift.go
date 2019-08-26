/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	openshiftAdmin = &OpenshiftAdmin{}
	openshiftView  = &OpenshiftView{}
)

type OpenshiftAdmin struct{}
type OpenshiftView struct{}

func (a *OpenshiftAdmin) Create(name, domain string, haflag bool, nworkers int32, key int64) (openshift *model.Openshift, err error) {
	db := DB()
	openshift = &model.Openshift{
		ClusterName: name,
		BaseDomain:  domain,
		Haflag:      haflag,
		WorkerNum:   nworkers,
	}
	err = db.Create(openshift).Error
	return
}

func (a *OpenshiftAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Openshift{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *OpenshiftAdmin) List(ctx context.Context, offset, limit int64, order string) (total int64, openshifts []*model.Openshift, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	where := memberShip.GetWhere()
	openshifts = []*model.Openshift{}
	if err = db.Model(&model.Openshift{}).Where(where).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&openshifts).Error; err != nil {
		return
	}

	return
}

func (v *OpenshiftView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 10
	}
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, openshifts, err := openshiftAdmin.List(c.Req.Context(), offset, limit, order)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Openshifts"] = openshifts
	c.Data["Total"] = total
	c.Data["Pages"] = GetPages(total, limit)
	c.HTML(200, "openshifts")
}

func (v *OpenshiftView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	openshiftID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", int64(openshiftID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = openshiftAdmin.Delete(int64(openshiftID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "openshifts",
	})
	return
}

func (v *OpenshiftView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "openshifts_new")
}

func (v *OpenshiftView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../openshift"
	name := c.QueryTrim("clustername")
	domain := c.QueryTrim("basedomain")
	haflagStr := c.QueryTrim("haflag")
	nworkers := c.QueryInt("nworkers")
	key := c.QueryInt64("key")
	haflag := false
	if haflagStr == "" || haflagStr == "no" {
		haflag = false
	} else if haflagStr == "yes" {
		haflag = true
	}
	_, err := openshiftAdmin.Create(name, domain, haflag, int32(nworkers), key)
	if err != nil {
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
