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

var keyAPI = &KeyAPI{}
var keyAdmin = &routes.KeyAdmin{}

type KeyAPI struct{}

type KeyResponse struct {
	*common.BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type KeyListResponse struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Keys   []*KeyResponse `json:"keys"`
}

type KeyPayload struct {
}

type KeyPatchPayload struct {
}

//
// @Summary get a key
// @Description get a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} KeyResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys/:id [get]
func (v *KeyAPI) Get(c *gin.Context) {
	keyResp := &KeyResponse{}
	c.JSON(http.StatusOK, keyResp)
}

//
// @Summary patch a key
// @Description patch a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   KeyPatchPayload  true   "Key patch payload"
// @Success 200 {object} KeyResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys/:id [patch]
func (v *KeyAPI) Patch(c *gin.Context) {
	keyResp := &KeyResponse{}
	c.JSON(http.StatusOK, keyResp)
}

//
// @Summary delete a key
// @Description delete a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys/:id [delete]
func (v *KeyAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a key
// @Description create a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   KeyPayload  true   "Key create payload"
// @Success 200 {object} KeyResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys [post]
func (v *KeyAPI) Create(c *gin.Context) {
	keyResp := &KeyResponse{}
	c.JSON(http.StatusOK, keyResp)
}

//
// @Summary list keys
// @Description list keys
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} KeyListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys [get]
func (v *KeyAPI) List(c *gin.Context) {
	keyListResp := &KeyListResponse{}
	c.JSON(http.StatusOK, keyListResp)
}
