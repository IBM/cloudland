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
	"os"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	imageAdmin = &ImageAdmin{}
	imageView  = &ImageView{}
)

type ImageAdmin struct{}
type ImageView struct{}

func FileExist(filename string) bool {
	_, err := os.Lstat(filename)
	return !os.IsNotExist(err)
}

func (a *ImageAdmin) Create(ctx context.Context, name, osVersion, virtType, userName, url, architecture string, instID int64) (image *model.Image, err error) {
	memberShip := GetMemberShip(ctx)
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	image = &model.Image{Model: model.Model{Creater: memberShip.UserID}, Owner: memberShip.OrgID, OsVersion: osVersion, VirtType: virtType, UserName: userName, Name: name, OSCode: name, Status: "creating", Architecture: architecture}
	err = db.Create(image).Error
	if err != nil {
		log.Println("DB create image failed, %v", err)
	}
	prefix := strings.Split(image.UUID, "-")[0]
	control := "inter="
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_image.sh '%d' '%s' '%s'", image.ID, prefix, url)
	if instID > 0 {
		instance := &model.Instance{Model: model.Model{ID: instID}}
		err = db.Take(instance).Error
		if err != nil {
			log.Println("DB failed to query instance", err)
			return
		}
		control = fmt.Sprintf("inter=%d", instance.Hyper)
		command = fmt.Sprintf("/opt/cloudland/scripts/backend/capture_image.sh '%d' '%s' '%d'", image.ID, prefix, instance.ID)
	}
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create image command execution failed", err)
		return
	}
	return
}

func (a *ImageAdmin) GetImageByUUID(ctx context.Context, uuID string) (image *model.Image, err error) {
	db := DB()
	image = &model.Image{}
	err = db.Where("uuid = ?", uuID).Take(image).Error
	if err != nil {
		log.Println("Failed to query image, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized to get image")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *ImageAdmin) GetImageByName(ctx context.Context, name string) (image *model.Image, err error) {
	db := DB()
	image = &model.Image{}
	err = db.Where("name = ?", name).Take(image).Error
	if err != nil {
		log.Println("Failed to query image, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized to get image")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *ImageAdmin) Get(ctx context.Context, id int64) (image *model.Image, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid image ID: %d", id)
		log.Println(err)
		return
	}
	db := DB()
	image = &model.Image{Model: model.Model{ID: id}}
	err = db.Take(image).Error
	if err != nil {
		log.Println("DB failed to query image, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized to get image")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *ImageAdmin) GetImage(ctx context.Context, reference *BaseReference) (image *model.Image, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Image base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		image, err = a.GetImageByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		image, err = a.GetImageByName(ctx, reference.Name)
		return
	}
	return
}

func (a *ImageAdmin) Delete(ctx context.Context, image *model.Image) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, image.Owner)
	if !permit {
		log.Println("Not authorized to delete image")
		err = fmt.Errorf("Not authorized")
		return
	}
	if image.Status == "available" {
		prefix := strings.Split(image.UUID, "-")[0]
		control := "inter="
		command := fmt.Sprint("/opt/cloudland/scripts/backend/clear_image.sh %d %s", image.ID, prefix, image.Format)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Clear image command execution failed", err)
			return
		}
	}
	if err = db.Delete(image).Error; err != nil {
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
	ctx := c.Req.Context()
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
	image, err := imageAdmin.Get(ctx, int64(imageID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = imageAdmin.Delete(ctx, image)
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
	instance := c.QueryInt64("instance")
	osVersion := c.QueryTrim("osVersion")
	virtType := "kvm-x86_64"
	userName := c.QueryTrim("userName")
	architecture := "x86_64"
	_, err := imageAdmin.Create(c.Req.Context(), name, osVersion, virtType, userName, url, architecture, instance)
	if err != nil {
		log.Println("Create instance failed", err)
		c.HTML(http.StatusBadRequest, err.Error())
		return
	}
	c.Redirect(redirectTo)
}
