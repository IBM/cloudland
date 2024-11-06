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

var secgroupAPI = &SecgroupAPI{}
var secgroupAdmin = &routes.SecgroupAdmin{}

type SecgroupAPI struct{}

type SecgroupResponse struct {
	*BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type SecgroupListResponse struct {
	Offset  int            `json:"offset"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Secgroups []*VPCResponse `json:"secgroups"`
}

type SecgroupPayload struct {
}

type SecgroupPatchPayload struct {
}

//
// @Summary get a secgroup
// @Description get a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecgroupResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id [get]
func (v *SecgroupAPI) Get(c *gin.Context) {
	secgroupResp := &SecgroupResponse{}
	c.JSON(http.StatusOK, secgroupResp)
}

//
// @Summary patch a secgroup
// @Description patch a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecgroupPatchPayload  true   "Secgroup patch payload"
// @Success 200 {object} SecgroupResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id [patch]
func (v *SecgroupAPI) Patch(c *gin.Context) {
	secgroupResp := &SecgroupResponse{}
	c.JSON(http.StatusOK, secgroupResp)
}

//
// @Summary delete a secgroup
// @Description delete a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id [delete]
func (v *SecgroupAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a secgroup
// @Description create a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecgroupPayload  true   "Secgroup create payload"
// @Success 200 {object} SecgroupResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups [post]
func (v *SecgroupAPI) Create(c *gin.Context) {
	secgroupResp := &SecgroupResponse{}
	c.JSON(http.StatusOK, secgroupResp)
}

//
// @Summary list secgroups
// @Description list secgroups
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecgroupListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups [get]
func (v *SecgroupAPI) List(c *gin.Context) {
	secgroupListResp := &SecgroupListResponse{}
	c.JSON(http.StatusOK, secgroupListResp)
}
