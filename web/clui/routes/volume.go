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
	volumeAdmin = &VolumeAdmin{}
	volumeView  = &VolumeView{}
)

type VolumeAdmin struct{}
type VolumeView struct{}

func (a *VolumeAdmin) Create(ctx context.Context, name string, size int) (volume *model.Volume, err error) {
	db := DB()
	volume = &model.Volume{Name: name, Format: "raw", Size: int32(size), Status: "pending"}
	err = db.Create(volume).Error
	if err != nil {
		log.Println("DB failed to create volume", err)
		return
	}
	control := fmt.Sprintf("inter=")
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_volume.sh %d %d", volume.ID, volume.Size)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create volume execution failed", err)
		return
	}
	return
}

func (a *VolumeAdmin) Update(ctx context.Context, id int64, name string, instID int64) (volume *model.Volume, err error) {
	db := DB()
	volume = &model.Volume{Model: model.Model{ID: id}}
	if err = db.Preload("Instance").Take(volume).Error; err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	if volume.InstanceID > 0 && instID > 0 && volume.InstanceID != instID {
		err = fmt.Errorf("Pease detach volume before attach it to new instance")
		return
	}
	if name != "" {
		volume.Name = name
	}
	if volume.InstanceID > 0 && instID == 0 && volume.Status == "attached" {
		control := fmt.Sprintf("inter=%d", volume.Instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/detach_volume.sh %d %d", volume.Instance.ID, volume.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Detach volume execution failed", err)
			return
		}
		volume.Instance = nil
		volume.InstanceID = 0
	} else if instID > 0 && volume.InstanceID == 0 && volume.Status == "available" {
		instance := &model.Instance{Model: model.Model{ID: instID}}
		if err = db.Model(instance).Take(instance).Error; err != nil {
			log.Println("DB: query instance failed", err)
			return
		}
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/attach_volume.sh %d %d %s", instance.ID, volume.ID, volume.Path)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Create volume execution failed", err)
			return
		}
		volume.InstanceID = instID
		volume.Instance = nil
	}
	if err = db.Model(volume).Save(volume).Error; err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	return
}

func (a *VolumeAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	volume := &model.Volume{Model: model.Model{ID: id}}
	if err = db.Model(volume).Take(volume).Error; err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	if err = db.Model(volume).Delete(volume).Error; err != nil {
		log.Println("DB: update volume failed", err)
		return
	}
	control := fmt.Sprintf("inter=")
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_volume.sh %d %d", volume.ID, volume.Path)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete volume execution failed", err)
		return
	}
	if err = db.Delete(volume).Error; err != nil {
		log.Println("DB: update volume failed", err)
		return
	}
	return
}

func (a *VolumeAdmin) List(offset, limit int64, order string) (total int64, volumes []*model.Volume, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	where := memberShip.GetWhere()
	volumes = []*model.Volume{}
	if err = db.Model(&model.Volume{}).Where(where).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Instance").Where(where).Find(&volumes).Error; err != nil {
		return
	}

	return
}

func (v *VolumeView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, volumes, err := volumeAdmin.List(offset, limit, order)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Volumes"] = volumes
	c.Data["Total"] = total
	c.HTML(200, "volumes")
}

func (v *VolumeView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	volumeID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = volumeAdmin.Delete(c.Req.Context(), int64(volumeID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "volumes",
	})
	return
}

func (v *VolumeView) New(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "volumes_new")
}

func (v *VolumeView) Edit(c *macaron.Context, store session.Store) {
	db := DB()
	id := c.Params(":id")
	volID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	volume := &model.Volume{Model: model.Model{ID: int64(volID)}}
	if err := db.Preload("Instance").Take(volume).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	instances := []*model.Instance{}
	if err := db.Find(&instances).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["Volume"] = volume
	c.Data["Instances"] = instances
	c.HTML(200, "volumes_patch")
}

func (v *VolumeView) Patch(c *macaron.Context, store session.Store) {
	redirectTo := "../volumes"
	id := c.Params(":id")
	name := c.Query("name")
	instance := c.Query("instance")
	volID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	instID, err := strconv.Atoi(instance)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	_, err = volumeAdmin.Update(c.Req.Context(), int64(volID), name, int64(instID))
	if err != nil {
		c.HTML(500, err.Error())
	}
	c.Redirect(redirectTo)
	return
}

func (v *VolumeView) Create(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../volumes"
	name := c.Query("name")
	size := c.Query("size")
	vsize, err := strconv.Atoi(size)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	_, err = volumeAdmin.Create(c.Req.Context(), name, vsize)
	if err != nil {
		c.HTML(500, err.Error())
	}
	c.Redirect(redirectTo)
}
