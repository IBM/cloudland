/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"net/http"
	"strconv"

	"web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var flavorAPI = &FlavorAPI{}
var flavorAdmin = &routes.FlavorAdmin{}

type FlavorAPI struct{}

type FlavorResponse struct {
	Name   string `json:"name"`
	Cpu    int32  `json:"cpu"`
	Memory int32  `json:"memory"`
	Disk   int32  `json:"disk"`
}

type FlavorListResponse struct {
	Offset  int               `json:"offset"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Flavors []*FlavorResponse `json:"flavors"`
}

type FlavorPayload struct {
}

type FlavorPatchPayload struct {
}

//
// @Summary get a flavor
// @Description get a flavor
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} FlavorResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /flavors/{id} [get]
func (v *FlavorAPI) Get(c *gin.Context) {
	flavorResp := &FlavorResponse{}
	c.JSON(http.StatusOK, flavorResp)
}

//
// @Summary patch a flavor
// @Description patch a flavor
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   FlavorPatchPayload  true   "Flavor patch payload"
// @Success 200 {object} FlavorResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /flavors/{id} [patch]
func (v *FlavorAPI) Patch(c *gin.Context) {
	flavorResp := &FlavorResponse{}
	c.JSON(http.StatusOK, flavorResp)
}

//
// @Summary delete a flavor
// @Description delete a flavor
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /flavors/{id} [delete]
func (v *FlavorAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a flavor
// @Description create a flavor
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   FlavorPayload  true   "Flavor create payload"
// @Success 200 {object} FlavorResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /flavors [post]
func (v *FlavorAPI) Create(c *gin.Context) {
	flavorResp := &FlavorResponse{}
	c.JSON(http.StatusOK, flavorResp)
}

func (v *FlavorAPI) getFlavorResponse(ctx context.Context, flavor *model.Flavor) (flavorResp *FlavorResponse, err error) {
	flavorResp = &FlavorResponse{
		Name: flavor.Name,
		Cpu:    flavor.Cpu,
		Memory: flavor.Memory,
		Disk:   flavor.Disk,
	}
	return
}

//
// @Summary list flavors
// @Description list flavors
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} FlavorListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /flavors [get]
func (v *FlavorAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, flavors, err := flavorAdmin.List(int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Failed to list flavors", err)
		return
	}
	flavorListResp := &FlavorListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(flavors),
	}
	flavorListResp.Flavors = make([]*FlavorResponse, flavorListResp.Limit)
	for i, flavor := range flavors {
		flavorListResp.Flavors[i], err = v.getFlavorResponse(ctx, flavor)
		if err != nil {
			common.ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, flavorListResp)
}
