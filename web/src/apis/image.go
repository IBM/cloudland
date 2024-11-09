/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"

	"web/src/common"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var imageAPI = &ImageAPI{}
var imageAdmin = &routes.ImageAdmin{}

type ImageAPI struct{}

type ImageResponse struct {
	*common.BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type ImageListResponse struct {
	Offset int              `json:"offset"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Images []*ImageResponse `json:"images"`
}

type ImagePayload struct {
}

type ImagePatchPayload struct {
}

//
// @Summary get a image
// @Description get a image
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} ImageResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images/:id [get]
func (v *ImageAPI) Get(c *gin.Context) {
	imageResp := &ImageResponse{}
	c.JSON(http.StatusOK, imageResp)
}

//
// @Summary patch a image
// @Description patch a image
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   ImagePatchPayload  true   "Image patch payload"
// @Success 200 {object} ImageResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images/:id [patch]
func (v *ImageAPI) Patch(c *gin.Context) {
	imageResp := &ImageResponse{}
	c.JSON(http.StatusOK, imageResp)
}

//
// @Summary delete a image
// @Description delete a image
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images/:id [delete]
func (v *ImageAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a image
// @Description create a image
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   ImagePayload  true   "Image create payload"
// @Success 200 {object} ImageResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images [post]
func (v *ImageAPI) Create(c *gin.Context) {
	imageResp := &ImageResponse{}
	c.JSON(http.StatusOK, imageResp)
}

//
// @Summary list images
// @Description list images
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} ImageListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images [get]
func (v *ImageAPI) List(c *gin.Context) {
	imageListResp := &ImageListResponse{}
	c.JSON(http.StatusOK, imageListResp)
}
