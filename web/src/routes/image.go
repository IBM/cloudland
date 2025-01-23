/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"fmt"
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

func (a *ImageAdmin) Create(ctx context.Context, osCode, name, osVersion, virtType, userName, url, architecture string, qaEnabled bool, instID int64) (image *model.Image, err error) {
	logger.Debugf("Creating image %s %s %s %s %s %s %s %t %d", osCode, name, osVersion, virtType, userName, url, architecture, qaEnabled, instID)
	memberShip := GetMemberShip(ctx)
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	var instance *model.Instance
	if instID > 0 {
		instance = &model.Instance{Model: model.Model{ID: instID}}
		err = db.Preload("Image").Preload("Volumes").Take(instance).Error
		if err != nil {
			logger.Error("DB failed to query instance", err)
			return
		}
		if instance.Status != "shut_off" {
			err = fmt.Errorf("instance [%s] is running, shut it down first before capturing", instance.Hostname)
			logger.Error(err)
			return
		}
		image = instance.Image.Clone()
		image.Model = model.Model{Creater: memberShip.UserID}
		image.Owner = memberShip.OrgID
		image.Name = name
		image.Status = "creating"
		image.CaptureFromInstanceID = instance.ID
		image.CaptureFromInstance = instance
	} else {
		image = &model.Image{
			Model:        model.Model{Creater: memberShip.UserID},
			Owner:        memberShip.OrgID,
			OsVersion:    osVersion,
			VirtType:     virtType,
			UserName:     userName,
			Name:         name,
			OSCode:       osCode,
			Status:       "creating",
			Architecture: architecture,
			QAEnabled:    qaEnabled,
		}
	}
	logger.Debugf("Creating image %+v", image)
	err = db.Create(image).Error
	if err != nil {
		logger.Error("DB create image failed, %v", err)
	}
	prefix := strings.Split(image.UUID, "-")[0]
	control := "inter="
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_image.sh '%d' '%s' '%s'", image.ID, prefix, url)
	if instID > 0 {
		bootVolumeUUID := ""
		if instance.Volumes != nil {
			for _, volume := range instance.Volumes {
				if volume.Booting {
					bootVolumeUUID = volume.GetOriginVolumeID()
					break
				}
			}
		}
		control = fmt.Sprintf("inter=%d", instance.Hyper)
		command = fmt.Sprintf("/opt/cloudland/scripts/backend/capture_image.sh '%d' '%s' '%d' '%s'", image.ID, prefix, instance.ID, bootVolumeUUID)
	}
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Create image command execution failed", err)
		return
	}
	return
}

func (a *ImageAdmin) GetImageByUUID(ctx context.Context, uuID string) (image *model.Image, err error) {
	db := DB()
	image = &model.Image{}
	err = db.Where("uuid = ?", uuID).Take(image).Error
	if err != nil {
		logger.Error("Failed to query image, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get image")
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
		logger.Error("Failed to query image, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get image")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *ImageAdmin) Get(ctx context.Context, id int64) (image *model.Image, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid image ID: %d", id)
		logger.Error(err)
		return
	}
	db := DB()
	image = &model.Image{Model: model.Model{ID: id}}
	err = db.Take(image).Error
	if err != nil {
		logger.Error("DB failed to query image, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get image")
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
		logger.Error("Not authorized to delete image")
		err = fmt.Errorf("Not authorized")
		return
	}
	refCount := 0
	err = db.Model(&model.Instance{}).Where("image_id = ?", image.ID).Count(&refCount).Error
	if err != nil {
		logger.Error("Failed to count the number of instances using the image", err)
		return
	}
	if refCount > 0 {
		logger.Error("Image can not be deleted if there are instances using it")
		err = fmt.Errorf("Image can not be deleted if there are instances using it")
		return
	}
	if image.Status == "available" {
		prefix := strings.Split(image.UUID, "-")[0]
		control := "inter="
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_image.sh '%d' '%s' '%s'", image.ID, prefix, image.Format)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Clear image command execution failed", err)
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
		logger.Error("Not authorized for this operation")
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
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusInternalServerError)
		return
	}
	pages := GetPages(total, limit)
	c.Data["Images"] = images
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.HTML(200, "images")
}

func (v *ImageView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	imageID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	image, err := imageAdmin.Get(ctx, int64(imageID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = imageAdmin.Delete(ctx, image)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
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
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, instances, err := instanceAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusInternalServerError)
		return
	}
	c.Data["Instances"] = instances
	c.HTML(200, "images_new")
}

func (v *ImageView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../images"
	osCode := c.QueryTrim("osCode")
	name := c.QueryTrim("name")
	url := c.QueryTrim("url")
	instance := c.QueryInt64("instance")
	osVersion := c.QueryTrim("osVersion")
	virtType := "kvm-x86_64"
	userName := c.QueryTrim("userName")
	qaEnabled := true
	architecture := "x86_64"
	_, err := imageAdmin.Create(c.Req.Context(), osCode, name, osVersion, virtType, userName, url, architecture, qaEnabled, instance)
	if err != nil {
		logger.Error("Create image failed", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}
