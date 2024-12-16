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
	"github.com/spf13/viper"
	macaron "gopkg.in/macaron.v1"
)

var (
	volumeAdmin = &VolumeAdmin{}
	volumeView  = &VolumeView{}
)

type VolumeAdmin struct{}
type VolumeView struct{}

func GetVolumeDriver() (driver string) {
	if viper.IsSet("volume.driver") {
		driver = viper.GetString("volume.driver")
	} else {
		driver = "local"
	}
	return
}

func (a *VolumeAdmin) Get(ctx context.Context, id int64) (volume *model.Volume, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	volume = &model.Volume{Model: model.Model{ID: id}}
	if err = db.Preload("Instance").Where(where).Take(volume).Error; err != nil {
		log.Println("Failed to query volume, %v", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, volume.Owner)
	if !permit {
		log.Println("Not authorized to read the volume")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *VolumeAdmin) GetVolumeByUUID(ctx context.Context, uuID string) (volume *model.Volume, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	volume = &model.Volume{}
	where := memberShip.GetWhere()
	err = db.Preload("Instance").Where(where).Where("uuid = ?", uuID).Take(volume).Error
	if err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, volume.Owner)
	if !permit {
		log.Println("Not authorized to read the volume")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *VolumeAdmin) CreateVolume(ctx context.Context, name string, size int32, instanceID int64,
	iopsLimit int32, iopsBurst int32, bpsLimit int32, bpsBurst int32, poolID string) (volume *model.Volume, err error) {
	db := DB()
	if iopsLimit == 0 {
		iopsLimit = viper.GetInt32("volume.default_iops_limit")
	}
	if iopsBurst == 0 {
		iopsBurst = viper.GetInt32("volume.default_iops_burst")
	}
	if bpsLimit == 0 {
		bpsLimit = viper.GetInt32("volume.default_bps_limit")
	}
	if bpsBurst == 0 {
		bpsBurst = viper.GetInt32("volume.default_bps_burst")
	}
	if poolID == "" {
		poolID = viper.GetString("volume.default_wds_pool_id")
	}
	memberShip := GetMemberShip(ctx)
	volume = &model.Volume{
		Model:     model.Model{Creater: memberShip.UserID},
		Owner:     memberShip.OrgID,
		Name:      name,
		Format:    "raw",
		Size:      int32(size),
		IopsLimit: iopsLimit,
		IopsBurst: iopsBurst,
		BpsLimit:  bpsLimit,
		BpsBurst:  bpsBurst,
		Status:    "pending",
	}
	err = db.Create(volume).Error
	if err != nil {
		log.Println("DB failed to create volume", err)
		return
	}
	return
}

func (a *VolumeAdmin) Create(ctx context.Context, name string, size int32,
	iopsLimit int32, iopsBurst int32, bpsLimit int32, bpsBurst int32, poolID string) (volume *model.Volume, err error) {
	memberShip := GetMemberShip(ctx)
	// check the permission
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized to create volume")
		err = fmt.Errorf("Not authorized")
		return
	}

	volume, err = a.CreateVolume(ctx, name, size, 0, iopsLimit, iopsBurst, bpsLimit, bpsBurst, poolID)
	if err != nil {
		log.Println("DB create volume failed", err)
		return
	}

	control := fmt.Sprintf("inter=")
	// RN-156: append the volume UUID to the command
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_volume_%s.sh '%d' '%d' '%s' '%d' '%d' '%d' '%d' '%s'",
		GetVolumeDriver(), volume.ID, volume.Size, volume.UUID, iopsLimit, iopsBurst, bpsLimit, bpsBurst, poolID)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create volume execution failed", err)
		return
	}
	return
}

func (a *VolumeAdmin) UpdateByUUID(ctx context.Context, uuid string, name string, instID int64) (volume *model.Volume, err error) {
	db := DB()
	volume = &model.Volume{}
	if err = db.Where("uuid = ?", uuid).Take(volume).Error; err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	return a.Update(ctx, volume.ID, name, instID)
}

func (a *VolumeAdmin) Update(ctx context.Context, id int64, name string, instID int64) (volume *model.Volume, err error) {
	db := DB()
	volume = &model.Volume{Model: model.Model{ID: id}}
	if err = db.Preload("Instance").Take(volume).Error; err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	// check the permission
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, volume.Owner)
	if !permit {
		log.Println("Not authorized to update the volume")
		err = fmt.Errorf("Not authorized")
		return
	}

	if volume.InstanceID > 0 && instID > 0 && volume.InstanceID != instID {
		err = fmt.Errorf("Pease detach volume before attach it to new instance")
		return
	}
	if name != "" {
		volume.Name = name
	}
	vol_driver := GetVolumeDriver()
	uuid := volume.UUID
	if vol_driver != "local" {
		uuid = volume.GetOriginVolumeID()
	}
	// RN-156: append the volume UUID to the command
	if volume.InstanceID > 0 && instID == 0 && volume.Status == "attached" {
		control := fmt.Sprintf("inter=%d", volume.Instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/detach_volume_%s.sh '%d' '%d' '%s'", vol_driver, volume.Instance.ID, volume.ID, uuid)
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
		// RN-156: append the volume UUID to the command
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/attach_volume_%s.sh '%d' '%d' '%s' '%s'", vol_driver, instance.ID, volume.ID, volume.GetVolumePath(), uuid)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Create volume execution failed", err)
			return
		}
		volume.InstanceID = instID
		volume.Instance = nil
	}
	if err = db.Model(volume).Save(volume).Error; err != nil {
		log.Println("DB: update volume failed", err)
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
	// check the permission
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, volume.Owner)
	if !permit {
		log.Println("Not authorized to delete the volume")
		err = fmt.Errorf("Not authorized")
		return
	}

	if err = db.Model(volume).Delete(volume).Error; err != nil {
		log.Println("DB: delete volume failed", err)
		return
	}
	control := fmt.Sprintf("inter=")
	vol_driver := GetVolumeDriver()
	uuid := volume.UUID
	if vol_driver != "local" {
		uuid = volume.GetOriginVolumeID()
	}
	log.Println("Delete volume", vol_driver, volume.ID, uuid, volume.GetVolumePath())
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_volume_%s.sh '%d' '%s' '%s'", vol_driver, volume.ID, uuid, volume.GetVolumePath())
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Delete volume execution failed", err)
		return
	}
	if err = db.Delete(volume).Error; err != nil {
		log.Println("DB: delete volume failed", err)
		return
	}
	return
}

func (a *VolumeAdmin) DeleteVolumeByUUID(ctx context.Context, uuID string) (err error) {
	db := DB()
	volume := &model.Volume{}
	if err = db.Where("uuid = ?", uuID).Take(volume).Error; err != nil {
		log.Println("DB: query volume failed", err)
		return
	}
	return a.Delete(ctx, volume.ID)
}

func (a *VolumeAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, volumes []*model.Volume, err error) {
	memberShip := GetMemberShip(ctx)
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
	where := memberShip.GetWhere()
	volumes = []*model.Volume{}
	if err = db.Model(&model.Volume{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Instance").Where(where).Where(query).Find(&volumes).Error; err != nil {
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, vol := range volumes {
			vol.OwnerInfo = &model.Organization{Model: model.Model{ID: vol.Owner}}
			if err = db.Take(vol.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				return
			}
		}
	}

	return
}

func (v *VolumeView) List(c *macaron.Context, store session.Store) {
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
	total, volumes, err := volumeAdmin.List(c.Req.Context(), offset, limit, order, query)
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
	c.Data["Volumes"] = volumes
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"volumes": volumes,
			"total":   total,
			"pages":   pages,
			"query":   query,
		})
		return
	}
	c.HTML(200, "volumes")
}

func (v *VolumeView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	volumeID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "volumes", int64(volumeID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = volumeAdmin.Delete(c.Req.Context(), int64(volumeID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "volumes",
	})
	return
}

func (v *VolumeView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "volumes_new")
}

func (v *VolumeView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params(":id")
	volID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "volumes", int64(volID))
	if err != nil {
		log.Println("Failed to check permission", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	volume := &model.Volume{Model: model.Model{ID: int64(volID)}}
	if err := db.Preload("Instance").Take(volume).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	_, instances, err := instanceAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["Volume"] = volume
	c.Data["Instances"] = instances
	c.HTML(200, "volumes_patch")
}

func (v *VolumeView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../volumes"
	id := c.Params(":id")
	name := c.QueryTrim("name")
	instance := c.QueryTrim("instance")
	volID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "volumes", int64(volID))
	if err != nil {
		log.Println("Failed to check permission", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}

	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	log.Printf("Patch volume(%s) to instance(%s)", id, instance)
	instID, err := strconv.Atoi(instance)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if instID > 0 {
		// have to check the instance permission
		permit, err = memberShip.CheckOwner(model.Writer, "instances", int64(instID))
		if err != nil {
			log.Println("Failed to check permission", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}

		if !permit {
			log.Println("Not authorized for this operation")
			c.Data["ErrorMsg"] = "Not authorized for this operation"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	volume, err := volumeAdmin.Update(c.Req.Context(), int64(volID), name, int64(instID))
	if err != nil {
		log.Println("Failed to update volume", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, volume)
		return
	}
	c.Redirect(redirectTo)
	return
}

func (v *VolumeView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../volumes"
	name := c.QueryTrim("name")
	size := c.QueryTrim("size")
	vsize, err := strconv.Atoi(size)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	volume, err := volumeAdmin.Create(c.Req.Context(), name, int32(vsize), 0, 0, 0, 0, "")
	if err != nil {
		log.Println("Create volume failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, err.Error())
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, volume)
		return
	}
	c.Redirect(redirectTo)
}
