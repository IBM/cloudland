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

var flavorAPI = &FlavorAPI{}
var flavorAdmin = &routes.FlavorAdmin{}

type FlavorAPI struct{}

type FlavorResponse struct {
	*common.BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
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
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /flavors/:id [get]
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
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /flavors/:id [patch]
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
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /flavors/:id [delete]
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
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /flavors [post]
func (v *FlavorAPI) Create(c *gin.Context) {
	flavorResp := &FlavorResponse{}
	c.JSON(http.StatusOK, flavorResp)
}

//
// @Summary list flavors
// @Description list flavors
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} FlavorListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /flavors [get]
func (v *FlavorAPI) List(c *gin.Context) {
	flavorListResp := &FlavorListResponse{}
	c.JSON(http.StatusOK, flavorListResp)
}
