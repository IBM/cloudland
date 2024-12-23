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

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	flavorAdmin = &FlavorAdmin{}
	flavorView  = &FlavorView{}
)

type FlavorAdmin struct{}
type FlavorView struct{}

func (a *FlavorAdmin) Create(ctx context.Context, name string, cpu, memory, disk int32) (flavor *model.Flavor, err error) {
	ctx, db := GetContextDB(ctx)
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	flavor = &model.Flavor{
		Name:   name,
		Cpu:    cpu,
		Disk:   disk,
		Memory: memory,
	}
	err = db.Create(flavor).Error
	return
}

func (a *FlavorAdmin) GetFlavorByName(ctx context.Context, name string) (flavor *model.Flavor, err error) {
	db := DB()
	flavor = &model.Flavor{}
	err = db.Where("name = ?", name).Take(flavor).Error
	if err != nil {
		log.Println("Failed to query flavor, %v", err)
		return
	}
	return
}

func (a *FlavorAdmin) Get(ctx context.Context, id int64) (flavor *model.Flavor, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid flavor ID: %d", id)
		log.Println(err)
		return
	}
	db := DB()
	flavor = &model.Flavor{Model: model.Model{ID: id}}
	err = db.Take(flavor).Error
	if err != nil {
		log.Println("DB failed to query flavor, err", err)
		return
	}
	return
}

func (a *FlavorAdmin) Delete(ctx context.Context, flavor *model.Flavor) (err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	ctx, db := GetContextDB(ctx)
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(flavor).Error; err != nil {
		log.Println("Failed to delete flavor", err)
		return
	}
	return
}

func (a *FlavorAdmin) List(offset, limit int64, order, query string) (total int64, flavors []*model.Flavor, err error) {
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

	flavors = []*model.Flavor{}
	if err = db.Model(&model.Flavor{}).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(query).Find(&flavors).Error; err != nil {
		return
	}

	return
}

func (v *FlavorView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, flavors, err := flavorAdmin.List(offset, limit, order, query)
	if err != nil {
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
	c.Data["Flavors"] = flavors
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"flavors": flavors,
			"total":   total,
			"pages":   pages,
			"query":   query,
		})
		return
	}
	c.HTML(200, "flavors")
}

func (v *FlavorView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.ParamsInt64("id")
	if id <= 0 {
		c.Data["ErrorMsg"] = "id <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor, err := flavorAdmin.Get(ctx, id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = flavorAdmin.Delete(ctx, flavor)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "flavors",
	})
	return
}

func (v *FlavorView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "flavors_new")
}

func (v *FlavorView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../flavors"
	name := c.Query("name")
	cores := c.Query("cpu")
	cpu, err := strconv.Atoi(cores)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	memory := c.QueryInt("memory")
	if memory <= 0 {
		c.Data["ErrorMsg"] = "memory <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}

	disk := c.QueryInt("disk")
	if disk <= 0 {
		c.Data["ErrorMsg"] = "disk <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, err = flavorAdmin.Create(c.Req.Context(), name, int32(cpu), int32(memory), int32(disk))
	if err != nil {
		log.Println("Create flavor failed", err)
		c.HTML(500, "500")
		return
	}
	c.Redirect(redirectTo)
}
