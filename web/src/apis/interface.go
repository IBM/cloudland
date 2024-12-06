/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"

	. "web/src/common"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var interfaceAPI = &InterfaceAPI{}
var interfaceAdmin = &routes.InterfaceAdmin{}

type InterfaceAPI struct{}

type InterfaceResponse struct {
	*BaseReference
	Subnet         *BaseReference    `json:"subnet"`
	MacAddress     string            `json:"mac_address"`
	IPAddress      string            `json:"ip_address"`
	IsPrimary      bool              `json:"is_primary"`
	FloatingIps    []*FloatingIpInfo `json:"floating_ips,omitempty"`
	SecurityGroups []*BaseReference  `json:"security_groups,omitempty"`
}

type InterfacePayload struct {
	Subnet         *BaseReference   `json:"subnet" binding:"required"`
	IpAddress      string           `json:"ip_address", binding:"omitempty,ipv4"`
	MacAddress     string           `json:"mac_address" binding:"omitempty,mac"`
	Name           string           `json:"name" binding:"omitempty,min=2,max=32"`
	SecurityGroups []*BaseReference `json:"security_group" binding:"omitempty"`
}

type InterfacePatchPayload struct {
}

// @Summary get a interface
// @Description get a interface
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} InterfaceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /interfaces/{id} [get]
func (v *InterfaceAPI) Get(c *gin.Context) {
	interfaceResp := &InterfaceResponse{}
	c.JSON(http.StatusOK, interfaceResp)
}

// @Summary patch a interface
// @Description patch a interface
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   InterfacePatchPayload  true   "Interface patch payload"
// @Success 200 {object} InterfaceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /interfaces/{id} [patch]
func (v *InterfaceAPI) Patch(c *gin.Context) {
	interfaceResp := &InterfaceResponse{}
	c.JSON(http.StatusOK, interfaceResp)
}

// @Summary delete a interface
// @Description delete a interface
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /interfaces/{id} [delete]
func (v *InterfaceAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a interface
// @Description create a interface
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   InterfacePayload  true   "Interface create payload"
// @Success 200 {object} InterfaceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instance/{id}/interfaces [post]
func (v *InterfaceAPI) Create(c *gin.Context) {
	interfaceResp := &InterfaceResponse{}
	c.JSON(http.StatusOK, interfaceResp)
}

// @Summary list interfaces
// @Description list interfaces
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {array} InterfaceResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instance/{id}/interfaces [get]
func (v *InterfaceAPI) List(c *gin.Context) {
	interfaceListResp := []*InterfaceResponse{}
	c.JSON(http.StatusOK, interfaceListResp)
}
