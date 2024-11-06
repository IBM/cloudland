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

var hyperAPI = &HyperAPI{}
var hyperAdmin = &routes.HyperAdmin{}

type HyperAPI struct{}

type HyperResponse struct {
	*BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type HyperListResponse struct {
	Offset  int            `json:"offset"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Hypers []*VPCResponse `json:"hypers"`
}

type HyperPayload struct {
}

type HyperPatchPayload struct {
}

//
// @Summary get a hypervisor
// @Description get a hypervisor
// @tags Administration
// @Accept  json
// @Produce json
// @Success 200 {object} HyperResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /hypers/:id [get]
func (v *HyperAPI) Get(c *gin.Context) {
	hyperResp := &HyperResponse{}
	c.JSON(http.StatusOK, hyperResp)
}

//
// @Summary list hypervisors
// @Description list hypervisors
// @tags Administration
// @Accept  json
// @Produce json
// @Success 200 {object} HyperListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /hypers [get]
func (v *HyperAPI) List(c *gin.Context) {
	hyperListResp := &HyperListResponse{}
	c.JSON(http.StatusOK, hyperListResp)
}
