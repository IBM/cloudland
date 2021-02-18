/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	registryAdmin = &RegistryAdmin{}
	registryView  = &RegistryView{}
)

type RegistryAdmin struct{}
type RegistryView struct{}

func (a *RegistryAdmin) Create(label, ocpVersion, registryContent, initramfs, kernel, image, installer, cli, kubelet string) (registry *model.Registry, err error) {
	db := DB()
	registry = &model.Registry{
		Label:           label,
		OcpVersion:      ocpVersion,
		RegistryContent: registryContent,
		Initramfs:       initramfs, 
		Kernel           kernel, 
		Image            image, 
		Installer        installer, 
		Cli              cli, 
		Kubelet          kubelet,
		
	}
	err = db.Create(registry).Error
	return
}

func (a *RegistryAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Registry{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("Failed to delete registry", err)
		return
	}
	return
}

func (a *RegistryAdmin) List(offset, limit int64, order, query string) (total int64, registrys []*model.Registry, err error) {
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}
	if query != "" {
		query = fmt.Sprintf("label like '%%%s%%'", query)
	}

	registrys = []*model.Registry{}
	if err = db.Model(&model.Registry{}).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(query).Find(&registrys).Error; err != nil {
		return
	}

	return
}

func (v *RegistryView) List(c *macaron.Context, store session.Store) {
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
	total, registrys, err := registryAdmin.List(offset, limit, order, query)
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
	c.Data["Registrys"] = registrys
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"registrys": registrys,
			"total":     total,
			"pages":     pages,
			"query":     query,
		})
		return
	}
	c.HTML(200, "registrys")
}

func (v *RegistryView) Delete(c *macaron.Context, store session.Store) (err error) {
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
	err = registryAdmin.Delete(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "registrys",
	})
	return
}

func (v *RegistryView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "registrys_new")
}

func (v *RegistryView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../registrys"
	label := c.Query("label")
	ocpVersion := c.Query("ocpversion")
	registryContent := c.Query("registrycontent")
	initramfs := c.Query("initramfs")
	kernel := c.Query("kernel")
	image := c.Query("image")
	installer := c.Query("installer")
	cli := c.Query("cli")
	kubelet := c.Query("kubelet")
	
	registry, err := registryAdmin.Create(label, ocpVersion, registryContent, initramfs, kernel, image, installer, cli, kubelet)
	if err != nil {
		log.Println("Create registry failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(500, "500")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, registry)
		return
	}
	c.Redirect(redirectTo)
}
