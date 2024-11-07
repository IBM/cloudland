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

var secruleAPI = &SecruleAPI{}
var secruleAdmin = &routes.SecruleAdmin{}

type SecruleAPI struct{}

type SecruleResponse struct {
	*common.BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type SecruleListResponse struct {
	Offset   int            `json:"offset"`
	Total    int            `json:"total"`
	Limit    int            `json:"limit"`
	Secrules []*VPCResponse `json:"secrules"`
}

type SecrulePayload struct {
}

type SecrulePatchPayload struct {
}

//
// @Summary get a secrule
// @Description get a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecruleResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id/rules/:rule_id [get]
func (v *SecruleAPI) Get(c *gin.Context) {
	secruleResp := &SecruleResponse{}
	c.JSON(http.StatusOK, secruleResp)
}

//
// @Summary patch a secrule
// @Description patch a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecrulePatchPayload  true   "Secrule patch payload"
// @Success 200 {object} SecruleResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id/rules/:rule_id [patch]
func (v *SecruleAPI) Patch(c *gin.Context) {
	secruleResp := &SecruleResponse{}
	c.JSON(http.StatusOK, secruleResp)
}

//
// @Summary delete a secrule
// @Description delete a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id/rules/:rule_id [delete]
func (v *SecruleAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a secrule
// @Description create a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecrulePayload  true   "Secrule create payload"
// @Success 200 {object} SecruleResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id/rules [post]
func (v *SecruleAPI) Create(c *gin.Context) {
	secruleResp := &SecruleResponse{}
	c.JSON(http.StatusOK, secruleResp)
}

//
// @Summary list secrules
// @Description list secrules
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecruleListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /security_groups/:id/rules [get]
func (v *SecruleAPI) List(c *gin.Context) {
	secruleListResp := &SecruleListResponse{}
	c.JSON(http.StatusOK, secruleListResp)
}
