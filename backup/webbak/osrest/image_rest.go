/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/go-openapi/strfmt"
	macaron "gopkg.in/macaron.v1"
)

var (
	imageInstance          = &ImageRest{}
	defaultContainerFormat = `bare`
)

type ImageRest struct{}

func (v *ImageRest) List(c *macaron.Context) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	_, images, err := imageAdmin.List(offset, limit, order, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewResponseError("List images fail", err.Error(), http.StatusInternalServerError))
		return
	}
	imagesOK := restModels.ListImagesOKBody{
		Images: restModels.Images{},
	}
	for _, image := range images {
		creatAt, _ := strfmt.ParseDateTime(image.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(image.UpdatedAt.Format(time.RFC3339))
		imageItem := &restModels.Image{
			Checksum:        image.Checksum,
			ContainerFormat: defaultContainerFormat,
			CreatedAt:       creatAt,
			DiskFormat:      image.Format,
			File:            "/v2/image/" + image.UUID + "file",
			ID:              image.UUID,
			MinDisk:         image.MiniDisk,
			MinRAM:          image.MiniMem,
			Name:            image.Name,
			OsHashAlgo:      image.OsHashAlgo,
			OsHashValue:     image.OsHashValue,
			Owner:           image.Holder,
			Protected:       image.Protected,
			UpdatedAt:       updateAt,
			Visibility:      image.Visibility,
			Status:          image.Status,
			Size:            image.Size,
			Self:            "/v2/image/" + image.UUID,
			Schema:          `/v2/schemas/image`,
		}
		imagesOK.Images = append(imagesOK.Images, imageItem)
	}
	c.JSON(http.StatusOK, imagesOK)
}

func (v *ImageRest) Create(c *macaron.Context) {
	db := DB()
	claims := c.Data[ClaimKey].(*HypercubeClaims)
	//check role
	if claims.Role < model.Writer {
		// if token was issued before promote user privilige, the user need to re-apply token
		c.Data["ErrorMsg"] = "claims.Role < model.Writer"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	uid := c.Data[claims.UID].(int64)
	oid := c.Data[claims.OID].(int64)
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`image.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{ErrorMsg: *err})
		return
	}
	requestData := &restModels.CreateImageParamsBody{}
	if err := json.Unmarshal(body, requestData); err != nil {
		c.JSON(http.StatusInternalServerError, NewResponseError("Unmarshal fail", err.Error(), http.StatusInternalServerError))
		return
	}
	if requestData.Name == "" {
		requestData.Name = generateName()
	}
	image := &model.Image{
		Name:      requestData.Name,
		MiniMem:   requestData.MinRAM,
		MiniDisk:  requestData.MinDisk,
		Protected: requestData.Protected,
		Model: model.Model{
			Creater: oid,
			Owner:   uid,
		},
	}
	if err := db.Create(image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, NewResponseError("create image fail", err.Error(), http.StatusInternalServerError))
		return
	}
	imageOK := &restModels.Image{
		// ID:        image.ID,
		// CreatedAt: image.CreatedAt,
	}
	c.JSON(http.StatusCreated, imageOK)

}
