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

var subnetAPI = &SubnetAPI{}
var subnetAdmin = &routes.SubnetAdmin{}

type SubnetAPI struct{}

type SubnetResponse struct {
	*common.BaseReference
}

type SubnetListResponse struct {
	Offset  int               `json:"offset"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Subnets []*SubnetResponse `json:"subnets"`
}

type SubnetPayload struct {
}

type SubnetPatchPayload struct {
}

//
// @Summary get a subnet
// @Description get a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SubnetResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /subnets/:id [get]
func (v *SubnetAPI) Get(c *gin.Context) {
	subnetResp := &SubnetResponse{}
	c.JSON(http.StatusOK, subnetResp)
}

//
// @Summary patch a subnet
// @Description patch a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SubnetPatchPayload  true   "Subnet patch payload"
// @Success 200 {object} SubnetResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /subnets/:id [patch]
func (v *SubnetAPI) Patch(c *gin.Context) {
	subnetResp := &SubnetResponse{}
	c.JSON(http.StatusOK, subnetResp)
}

//
// @Summary delete a subnet
// @Description delete a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /subnets/:id [delete]
func (v *SubnetAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a subnet
// @Description create a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SubnetPayload  true   "Subnet create payload"
// @Success 200 {object} SubnetResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /subnets [post]
func (v *SubnetAPI) Create(c *gin.Context) {
	subnetResp := &SubnetResponse{}
	c.JSON(http.StatusOK, subnetResp)
}

//
// @Summary list subnets
// @Description list subnets
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SubnetListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /subnets [get]
func (v *SubnetAPI) List(c *gin.Context) {
	subnetListResp := &SubnetListResponse{}
	c.JSON(http.StatusOK, subnetListResp)
}
