/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	flavorAdmin = &FlavorAdmin{}
	flavorView  = &FlavorView{}
)

type FlavorAdmin struct{}
type FlavorView struct{}

func (a *FlavorAdmin) Create(name string, cpu, memory, disk, swap int32) (flavor *model.Flavor, err error) {
	db := DB()
	flavor = &model.Flavor{
		Name:   name,
		Cpu:    cpu,
		Disk:   disk,
		Memory: memory,
		Swap:   swap,
	}
	err = db.Create(flavor).Error
	return
}

func (a *FlavorAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Flavor{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *FlavorAdmin) List(offset, limit int64, order string) (total int64, flavors []*model.Flavor, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	flavors = []*model.Flavor{}
	if err = db.Model(&model.Flavor{}).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&flavors).Error; err != nil {
		return
	}

	return
}

func (v *FlavorView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, flavors, err := flavorAdmin.List(offset, limit, order)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Flavors"] = flavors
	c.Data["Total"] = total
	c.HTML(200, "flavors")
}

func (v *FlavorView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	flavorID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = flavorAdmin.Delete(int64(flavorID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
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
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "flavors_new")
}

func (v *FlavorView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	if memberShip.UserName != "admin" {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../flavors"
	name := c.Query("name")
	cores := c.Query("cpu")
	cpu, err := strconv.Atoi(cores)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	mem := c.Query("memory")
	memory, err := strconv.Atoi(mem)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}

	dSize := c.Query("disk")
	disk, err := strconv.Atoi(dSize)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	sSize := c.Query("swap")
	swap := 0
	if sSize != "" {
		swap, err = strconv.Atoi(sSize)
		if err != nil {
			code := http.StatusBadRequest
			c.Error(code, http.StatusText(code))
			return
		}
	}
	_, err = flavorAdmin.Create(name, int32(cpu), int32(memory), int32(disk), int32(swap))
	if err != nil {
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
