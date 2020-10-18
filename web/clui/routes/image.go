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
	imageAdmin = &ImageAdmin{}
	imageView  = &ImageView{}
)

type ImageAdmin struct{}
type ImageView struct{}

func (a *ImageAdmin) Create(ctx context.Context, name, url, format, architecture string, instID int64) (image *model.Image, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if architecture == "" {
		architecture = "x86-64"
	}
	image = &model.Image{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Name: name, OSCode: name, Format: format, Status: "creating", Architecture: architecture}
	err = db.Create(image).Error
	if err != nil {
		log.Println("DB create image failed, %v", err)
	}
	if instID > 0 {
		instance := &model.Instance{Model: model.Model{ID: instID}}
		err = db.Take(instance).Error
		if err != nil {
			log.Println("DB failed to query instance", err)
			return
		}
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/capture_image.sh '%d' '%d'", image.ID, instance.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Create image command execution failed", err)
			return
		}
	} else {
		control := "inter=0"
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_image.sh '%d' '%s'", image.ID, url)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Create image command execution failed", err)
			return
		}
	}
	return
}

func (a *ImageAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	image := &model.Image{Model: model.Model{ID: id}}
	if err = db.Take(image).Error; err != nil {
		log.Println("Image query failed, %v", err)
		return
	}
	if image.Format == "available" {
		control := "inter="
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_image.sh '%d', '%s'", image.ID, image.Format)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Clear image command execution failed", err)
			return
		}
	}
	if err = db.Delete(&model.Image{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *ImageAdmin) List(offset, limit int64, order, query string) (total int64, images []*model.Image, err error) {
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
	images = []*model.Image{}
	if err = db.Model(&model.Image{}).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(query).Find(&images).Error; err != nil {
		return
	}

	return
}

func (v *ImageView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, images, err := imageAdmin.List(offset, limit, order, query)
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
	c.Data["Images"] = images
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"images": images,
			"total":  total,
			"pages":  pages,
			"query":  query,
		})
		return
	}
	c.HTML(200, "images")
}

func (v *ImageView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	imageID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "images", int64(imageID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = imageAdmin.Delete(c.Req.Context(), int64(imageID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "images",
	})
	return
}

func (v *ImageView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, instances, err := instanceAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Instances"] = instances
	c.HTML(200, "images_new")
}

func (v *ImageView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../images"
	name := c.QueryTrim("name")
	url := c.QueryTrim("url")
	format := c.QueryTrim("format")
	architecture := c.QueryTrim("architecture")
	instance := c.QueryInt64("instance")
	image, err := imageAdmin.Create(c.Req.Context(), name, url, format, architecture, instance)
	if err != nil {
		log.Println("Create instance failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, err.Error())
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, image)
		return
	}
	c.Redirect(redirectTo)
}
