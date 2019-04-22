/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"net/http"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	imageAdmin = &ImageAdmin{}
	imageView  = &ImageView{}
)

type ImageAdmin struct{}
type ImageView struct{}

func (a *ImageAdmin) Create(name, oscode, format, architecture string) (image *model.Image, err error) {
	db := DB()
	image = &model.Image{Name: name, OSCode: oscode, Format: format, Architecture: architecture}
	err = db.Create(image).Error
	return
}

func (a *ImageAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Image{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *ImageAdmin) List(offset, limit int64, order string) (total int64, images []*model.Image, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	images = []*model.Image{}
	if err = db.Model(&model.Image{}).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&images).Error; err != nil {
		return
	}

	return
}

func (v *ImageView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, images, err := imageAdmin.List(offset, limit, order)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Images"] = images
	c.Data["Total"] = total
	c.HTML(200, "images")
}

func (v *ImageView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	imageID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = imageAdmin.Delete(int64(imageID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "images",
	})
	return
}

func (v *ImageView) New(c *macaron.Context, store session.Store) {
	c.HTML(200, "images_new")
}

func (v *ImageView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../images"
	name := c.Query("name")
	oscode := c.Query("oscode")
	format := c.Query("format")
	architecture := c.Query("architecture")
	_, err := imageAdmin.Create(name, oscode, format, architecture)
	if err != nil {
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
