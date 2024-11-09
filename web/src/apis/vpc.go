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

var vpcAPI = &VPCAPI{}
var routerAdmin = &routes.RouterAdmin{}

type VPCAPI struct{}

type VPCResponse struct {
	*common.BaseReference
	Subnets []*common.BaseReference `json:"subnets,omitempty"`
}

type VPCListResponse struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	VPCs   []*VPCResponse `json:"vpcs"`
}

type VPCPayload struct {
}

type VPCPatchPayload struct {
}

//
// @Summary get a vpc
// @Description get a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} VPCResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs/:id [get]
func (v *VPCAPI) Get(c *gin.Context) {
	vpcResp := &VPCResponse{}
	c.JSON(http.StatusOK, vpcResp)
}

//
// @Summary patch a vpc
// @Description patch a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   VPCPatchPayload  true   "VPC patch payload"
// @Success 200 {object} VPCResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs/:id [patch]
func (v *VPCAPI) Patch(c *gin.Context) {
	vpcResp := &VPCResponse{}
	c.JSON(http.StatusOK, vpcResp)
}

//
// @Summary delete a vpc
// @Description delete a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs/:id [delete]
func (v *VPCAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a vpc
// @Description create a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   VPCPayload  true   "VPC create payload"
// @Success 200 {object} VPCResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs [post]
func (v *VPCAPI) Create(c *gin.Context) {
	vpcResp := &VPCResponse{}
	c.JSON(http.StatusOK, vpcResp)
}

//
// @Summary list vpcs
// @Description list vpcs
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} VPCListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs [get]
func (v *VPCAPI) List(c *gin.Context) {
	vpcListResp := &VPCListResponse{}
	c.JSON(http.StatusOK, vpcListResp)
}
