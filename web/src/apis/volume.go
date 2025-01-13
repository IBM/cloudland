/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var volumeAPI = &VolumeAPI{}
var volumeAdmin = &routes.VolumeAdmin{}

type VolumeAPI struct{}

type VolumePayload struct {
	Count     int    `json:"count" binding:"omitempty,gte=1,lte=16"`
	Name      string `json:"name" binding:"required"`
	Size      int32  `json:"size" binding:"required"`
	PoolID    string `json:"pool_id" binding:"omitempty"`
	IopsLimit int32  `json:"iops_limit" binding:"omitempty,gte=0"`
	IopsBurst int32  `json:"iops_burst" binding:"omitempty,gte=0"`
	BpsLimit  int32  `json:"bps_limit" binding:"omitempty,gte=0"`
	BpsBurst  int32  `json:"bps_burst" binding:"omitempty,gte=0"`
}

type VolumePatchPayload struct {
	Name      string  `json:"name" binding:"omitempty"`
	Size      int32   `json:"size" binding:"omitempty"`
	Instance  *BaseID `json:"instance" binding:"omitempty"`
	IopsLimit int32   `json:"iops_limit" binding:"omitempty"`
	IopsBurst int32   `json:"iops_burst" binding:"omitempty"`
	BpsLimit  int32   `json:"bps_limit" binding:"omitempty"`
	BpsBurst  int32   `json:"bps_burst" binding:"omitempty"`
}

type VolumeResponse struct {
	*ResourceReference
	Path      string         `json:"path"`
	Size      int32          `json:"size"`
	Format    string         `json:"format"`
	Status    string         `json:"status"`
	Target    string         `json:"target"`
	Href      string         `json:"href"`
	Booting   bool           `json:"booting"`
	Instance  *BaseReference `json:"instance"`
	IopsLimit int32          `json:"iops_limit"`
	IopsBurst int32          `json:"iops_burst"`
	BpsLimit  int32          `json:"bps_limit"`
	BpsBurst  int32          `json:"bps_burst"`
}

type VolumeListResponse struct {
	Offset  int               `json:"offset"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Volumes []*VolumeResponse `json:"volumes"`
}

// @Summary get a volume
// @Description get a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Param   id     path    string     true  "Volume UUID"
// @Success 200 {object} VolumeResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /volumes/{id} [get]
func (v *VolumeAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Get volume by uuid: %s", uuID)
	volume, err := volumeAdmin.GetVolumeByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get volume by uuid: %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid volume query", err)
		return
	}
	volumeResp, err := v.getVolumeResponse(ctx, volume)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Got volume : %+v", volumeResp)
	c.JSON(http.StatusOK, volumeResp)
}

// @Summary patch a volume
// @Description patch a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   VolumePatchPayload  true   "Volume patch payload"
// @Success 200 {object} VolumeResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /volumes/{id} [patch]
func (v *VolumeAPI) Patch(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	payload := &VolumePatchPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind json: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	logger.Debugf("Patching volume %s with %+v", uuID, payload)
	instanceID := int64(0)
	if payload.Instance != nil {
		var instance *model.Instance
		instance, err = instanceAdmin.GetInstanceByUUID(ctx, payload.Instance.ID)
		if err != nil {
			logger.Errorf("Failed to get instance, %+v", err)
			ErrorResponse(c, http.StatusBadRequest, "Failed to get instance", err)
			return
		}
		instanceID = instance.ID
	}

	volume, err := volumeAdmin.UpdateByUUID(ctx, uuID, payload.Name, instanceID)
	if err != nil {
		logger.Errorf("Failed to update volume %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to update volume", err)
		return
	}
	volumeResp, err := v.getVolumeResponse(ctx, volume)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update volume response", err)
		return
	}
	logger.Debugf("Patch volume successfully, %s, %+v", uuID, volumeResp)
	c.JSON(http.StatusOK, volumeResp)
}

// @Summary delete a volume
// @Description delete a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /volumes/{id} [delete]
func (v *VolumeAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Deleting volume %s", uuID)
	err := volumeAdmin.DeleteVolumeByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to delete volume %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to delete volume", err)
		return
	}
	logger.Debugf("Deleted volume %s successfully", uuID)
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a volume
// @Description create a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   VolumePayload  true   "Volume create payload"
// @Success 200 {object} VolumeResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /volumes [post]
func (v *VolumeAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &VolumePayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind json: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	logger.Debugf("Creating volume with %+v", payload)
	volume, err := volumeAdmin.Create(ctx, payload.Name, payload.Size,
		payload.IopsLimit, payload.IopsBurst, payload.BpsLimit, payload.BpsBurst, payload.PoolID)
	if err != nil {
		logger.Errorf("Failed to create volume: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to create volume", err)
		return
	}
	volumeResp, err := v.getVolumeResponse(ctx, volume)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create volume response", err)
		return
	}
	logger.Debugf("Created volume successfully, %+v", volumeResp)
	c.JSON(http.StatusOK, volumeResp)
}

// @Summary list volumes
// @Description list volumes
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} VolumeListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /volumes [get]
func (v *VolumeAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	nameStr := c.DefaultQuery("name", "")

	// type: all, data, boot
	// default data
	typeStr := c.DefaultQuery("type", "data")
	logger.Debugf("List volumes, offset:%s, limit:%s, name:%s, type:%s", offsetStr, limitStr, nameStr, typeStr)
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		logger.Errorf("Invalid query offset: %s, %+v", offsetStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logger.Errorf("Invalid query limit: %s, %+v", limitStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		errStr := "Invalid query offset or limit, cannot be negative"
		logger.Errorf(errStr)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", errors.New(errStr))
		return
	}
	total, volumes, err := volumeAdmin.ListVolume(ctx, int64(offset), int64(limit), "-created_at", nameStr, typeStr)
	if err != nil {
		logger.Errorf("Failed to list volumes, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to list volumes", err)
		return
	}
	volumeListResp := &VolumeListResponse{
		Offset: offset,
		Total:  int(total),
		Limit:  len(volumes),
	}
	volumeList := make([]*VolumeResponse, volumeListResp.Limit)
	for i, volume := range volumes {
		volumeResp, err := v.getVolumeResponse(ctx, volume)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to list volume response", err)
			return
		}
		volumeList[i] = volumeResp
	}

	volumeListResp.Volumes = volumeList
	logger.Debugf("List volumes successfully, %+v", volumeListResp)
	c.JSON(http.StatusOK, volumeListResp)
}

func (v *VolumeAPI) getVolumeResponse(ctx context.Context, volume *model.Volume) (*VolumeResponse, error) {
	owner := orgAdmin.GetOrgName(volume.Owner)
	volumeResp := &VolumeResponse{
		ResourceReference: &ResourceReference{
			ID:        volume.UUID,
			Name:      volume.Name,
			Owner:     owner,
			CreatedAt: volume.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: volume.UpdatedAt.Format(TimeStringForMat),
		},
		Path:      volume.Path,
		Size:      volume.Size,
		Status:    volume.Status,
		Target:    volume.Target,
		Href:      volume.Href,
		IopsLimit: volume.IopsLimit,
		IopsBurst: volume.IopsBurst,
		BpsLimit:  volume.BpsLimit,
		BpsBurst:  volume.BpsBurst,
		Booting:   volume.Booting,
	}
	if volume.Instance == nil {
		volumeResp.Instance = nil
	} else {
		volumeResp.Instance = &BaseReference{
			ID:   volume.Instance.UUID,
			Name: volume.Instance.Hostname,
		}
	}
	return volumeResp, nil
}
