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

var orgAPI = &OrgAPI{}
var orgAdmin = &routes.OrgAdmin{}

type OrgAPI struct{}

type OrgResponse struct {
	*common.BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type OrgListResponse struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Orgs   []*OrgResponse `json:"orgs"`
}

type OrgPayload struct {
}

type OrgPatchPayload struct {
}

//
// @Summary get a org
// @Description get a org
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} OrgResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /orgs/{id} [get]
func (v *OrgAPI) Get(c *gin.Context) {
	orgResp := &OrgResponse{}
	c.JSON(http.StatusOK, orgResp)
}

//
// @Summary patch a org
// @Description patch a org
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   OrgPatchPayload  true   "Org patch payload"
// @Success 200 {object} OrgResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /orgs/{id} [patch]
func (v *OrgAPI) Patch(c *gin.Context) {
	orgResp := &OrgResponse{}
	c.JSON(http.StatusOK, orgResp)
}

//
// @Summary delete a org
// @Description delete a org
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /orgs/{id} [delete]
func (v *OrgAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a org
// @Description create a org
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   OrgPayload  true   "Org create payload"
// @Success 200 {object} OrgResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /orgs [post]
func (v *OrgAPI) Create(c *gin.Context) {
	orgResp := &OrgResponse{}
	c.JSON(http.StatusOK, orgResp)
}

//
// @Summary list orgs
// @Description list orgs
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} OrgListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /orgs [get]
func (v *OrgAPI) List(c *gin.Context) {
	orgListResp := &OrgListResponse{}
	c.JSON(http.StatusOK, orgListResp)
}
