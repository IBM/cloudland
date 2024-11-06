/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"

	"web/src/routes"
	"github.com/gin-gonic/gin"
)

var volumeAPI = &VolumeAPI{}
var volumeAdmin = &routes.VolumeAdmin{}

type VolumeAPI struct{}

type VolumeResponse struct {
	*BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type VolumeListResponse struct {
	Offset  int            `json:"offset"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Volumes []*VPCResponse `json:"volumes"`
}

type VolumePayload struct {
}

type VolumePatchPayload struct {
}

//
// @Summary get a volume
// @Description get a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} VolumeResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /volumes/:id [get]
func (v *VolumeAPI) Get(c *gin.Context) {
	volumeResp := &VolumeResponse{}
	c.JSON(http.StatusOK, volumeResp)
}

//
// @Summary patch a volume
// @Description patch a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   VolumePatchPayload  true   "Volume patch payload"
// @Success 200 {object} VolumeResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /volumes/:id [patch]
func (v *VolumeAPI) Patch(c *gin.Context) {
	volumeResp := &VolumeResponse{}
	c.JSON(http.StatusOK, volumeResp)
}

//
// @Summary delete a volume
// @Description delete a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /volumes/:id [delete]
func (v *VolumeAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a volume
// @Description create a volume
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   VolumePayload  true   "Volume create payload"
// @Success 200 {object} VolumeResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /volumes [post]
func (v *VolumeAPI) Create(c *gin.Context) {
	volumeResp := &VolumeResponse{}
	c.JSON(http.StatusOK, volumeResp)
}

//
// @Summary list volumes
// @Description list volumes
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} VolumeListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /volumes [get]
func (v *VolumeAPI) List(c *gin.Context) {
	volumeListResp := &VolumeListResponse{}
	c.JSON(http.StatusOK, volumeListResp)
}
