/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var imageAPI = &ImageAPI{}
var imageAdmin = &routes.ImageAdmin{}

type ImageAPI struct{}

type ImageResponse struct {
	*BaseReference
	Size         int64  `json:"size"`
	Format       string `json:"format"`
	Architecture string `json:"architecture"`
	User         string `json:"user"`
	Status       string `json:"status"`
}

type ImageListResponse struct {
	Offset int              `json:"offset"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Images []*ImageResponse `json:"images"`
}

type ImagePayload struct {
	Name        string `json:"name" binding:"required,min=2,max=32"`
	DownloadURL string `json:"download_url" binding:"required,http_url"`
	OSVersion   string `json:"os_version" binding:"required,min=2,max=32"`
	User        string `json:"user" binding:"required,min=2,max=32"`
}

type ImagePatchPayload struct {
}

// @Summary get a image
// @Description get a image
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} ImageResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images/{id} [get]
func (v *ImageAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	image, err := imageAdmin.GetImageByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid vpc query", err)
		return
	}
	imageResp, err := v.getImageResponse(ctx, image)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, imageResp)
}

// @Summary patch a image
// @Description patch a image
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   ImagePatchPayload  true   "Image patch payload"
// @Success 200 {object} ImageResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images/{id} [patch]
func (v *ImageAPI) Patch(c *gin.Context) {
	imageResp := &ImageResponse{}
	c.JSON(http.StatusOK, imageResp)
}

// @Summary delete a image
// @Description delete a image
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images/{id} [delete]
func (v *ImageAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

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
	ctx := c.Request.Context()
	payload := &ImagePayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	image, err := imageAdmin.Create(ctx, payload.Name, payload.OSVersion, "kvm-x86_64", payload.User, payload.DownloadURL, "x86_64", 0)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	imageResp, err := v.getImageResponse(ctx, image)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, imageResp)
}

func (v *ImageAPI) getImageResponse(ctx context.Context, image *model.Image) (imageResp *ImageResponse, err error) {
	imageResp = &ImageResponse{
		BaseReference: &BaseReference{
			ID:   image.UUID,
			Name: image.Name,
		},
		Size:         image.Size,
		Format:       image.Format,
		Architecture: image.Architecture,
		User:         image.UserName,
		Status:       image.Status,
	}
	return
}

// @Summary list images
// @Description list images
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} ImageListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /images [get]
func (v *ImageAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, images, err := imageAdmin.List(int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list images", err)
		return
	}
	imageListResp := &ImageListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(images),
	}
	imageListResp.Images = make([]*ImageResponse, imageListResp.Limit)
	for i, image := range images {
		imageListResp.Images[i], err = v.getImageResponse(ctx, image)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, imageListResp)
}
