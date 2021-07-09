/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
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
	flavorAdmin = &FlavorAdmin{}
	flavorView  = &FlavorView{}
	apiFlavorView  = &APIFlavorView{}
)

type FlavorAdmin struct{}
type FlavorView struct{}
type APIFlavorView struct{
    Name        string
    Cpu         string
    Memory      int
    Dist        int
    Swap        int
    Ephemeral   int
}

func (a *FlavorAdmin) Create(name string, cpu, memory, disk, swap, ephemeral int32) (flavor *model.Flavor, err error) {
	db := DB()
	flavor = &model.Flavor{
		Name:      name,
		Cpu:       cpu,
		Disk:      disk,
		Memory:    memory,
		Swap:      swap,
		Ephemeral: ephemeral,
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
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	id := c.ParamsInt64("id")
	if id <= 0 {
		c.Data["ErrorMsg"] = "id <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = flavorAdmin.Delete(id)
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
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
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
	swap := c.QueryInt("swap")
	ephemeral := c.QueryInt("ephemeral")
	flavor, err := flavorAdmin.Create(name, int32(cpu), int32(memory), int32(disk), int32(swap), int32(ephemeral))
	if err != nil {
		log.Println("Create flavor failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(500, "500")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, flavor)
		return
	}
	c.Redirect(redirectTo)
}

func (v *APIFlavorView) List(c *macaron.Context, store session.Store) {
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
		c.JSON(500, map[string]interface{}{
			"ErrorMsg": "Failed to list flavors."+err.Error(),
		})
		return
	}
	pages := GetPages(total, limit)

	c.JSON(200, map[string]interface{}{
		"flavors": flavors,
		"total":   total,
		"pages":   pages,
		"query":   query,
	})
}

func (v *APIFlavorView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.JSON(403, map[string]interface{}{
			"ErrorMsg": "Not authorized for this operation",
		})
		return
	}

	id := c.Params("id")
	if id == "" {
		log.Println("ID is empty, %v", err)
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Flavor id is empty.",
		})
		return
	}
	flavorID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid flavor ID, %v", err)
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Failed to get flavor id.",
		})
		return
	}	
	
	err = flavorAdmin.Delete(flavorID)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"ErrorMsg": "Failed to delete flavor."+err.Error(),
		})
		return
	}
	c.JSON(200, map[string]interface{}{
		"Msg": "Success.",
	})
	return
}

func (v *APIFlavorView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.JSON(403, map[string]interface{}{
			"ErrorMsg": "Not authorized for this operation",
		})
		return
	}
	name := c.Query("name")
	cores := c.Query("cpu")
	cpu, err := strconv.Atoi(cores)
	if err != nil {
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Failed to get numbe of cpu.",
		})
		return
	}
	memory := c.QueryInt("memory")
	if memory <= 0 {
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Failed to get memory size.",
		})
		return
	}

	disk := c.QueryInt("disk")
	if disk <= 0 {
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Failed to get  disk size.",
		})
		return
	}
	swap := c.QueryInt("swap")
	ephemeral := c.QueryInt("ephemeral")
	flavor, err := flavorAdmin.Create(name, int32(cpu), int32(memory), int32(disk), int32(swap), int32(ephemeral))
	if err != nil {
		log.Println("Create flavor failed", err)
		c.JSON(500, map[string]interface{}{
			"ErrorMsg": "Failed to create flavor."+err.Error(),
		})
		return
	}
	 
	c.JSON(200, flavor)
	
}
