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

var floatingIpAPI = &FloatingIpAPI{}
var floatingIpAdmin = &routes.FloatingIpAdmin{}

type FloatingIpAPI struct{}

type FloatingIpInfo struct {
	*common.BaseReference
	IpAddress string `json:"ip_address"`
}

type FloatingIpResponse struct {
	*common.BaseReference
}

type FloatingIpListResponse struct {
	Offset      int                   `json:"offset"`
	Total       int                   `json:"total"`
	Limit       int                   `json:"limit"`
	FloatingIps []*FloatingIpResponse `json:"floatingIps"`
}

type FloatingIpPayload struct {
}

type FloatingIpPatchPayload struct {
}

//
// @Summary get a floating ip
// @Description get a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} FloatingIpResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips/:id [get]
func (v *FloatingIpAPI) Get(c *gin.Context) {
	floatingIpResp := &FloatingIpResponse{}
	c.JSON(http.StatusOK, floatingIpResp)
}

//
// @Summary patch a floating ip
// @Description patch a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   FloatingIpPatchPayload  true   "Floating ip patch payload"
// @Success 200 {object} FloatingIpResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips/:id [patch]
func (v *FloatingIpAPI) Patch(c *gin.Context) {
	floatingIpResp := &FloatingIpResponse{}
	c.JSON(http.StatusOK, floatingIpResp)
}

//
// @Summary delete a floating ip
// @Description delete a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips/:id [delete]
func (v *FloatingIpAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a floating ip
// @Description create a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   FloatingIpPayload  true   "Floating ip create payload"
// @Success 200 {object} FloatingIpResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips [post]
func (v *FloatingIpAPI) Create(c *gin.Context) {
	floatingIpResp := &FloatingIpResponse{}
	c.JSON(http.StatusOK, floatingIpResp)
}

//
// @Summary list floating ips
// @Description list floating ips
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} FloatingIpListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips [get]
func (v *FloatingIpAPI) List(c *gin.Context) {
	floatingIpListResp := &FloatingIpListResponse{}
	c.JSON(http.StatusOK, floatingIpListResp)
}
