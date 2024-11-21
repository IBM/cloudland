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

var zoneAPI = &ZoneAPI{}
var zoneAdmin = &routes.ZoneAdmin{}

type ZoneAPI struct{}

type ZoneResponse struct {
	*BaseReference
	Cpu    int32 `json:"cpu"`
	Memory int32 `json:"memory"`
	Disk   int32
}

type ZoneListResponse struct {
	Offset int             `json:"offset"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Zones  []*ZoneResponse `json:"zones"`
}

//
// @Summary get a zonevisor
// @Description get a zonevisor
// @tags Zone
// @Accept  json
// @Produce json
// @Success 200 {object} ZoneResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /zones/{name} [get]
func (v *ZoneAPI) Get(c *gin.Context) {
	zoneResp := &ZoneResponse{}
	c.JSON(http.StatusOK, zoneResp)
}

//
// @Summary list zonevisors
// @Description list zonevisors
// @tags Zone
// @Accept  json
// @Produce json
// @Success 200 {object} ZoneListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /zones [get]
func (v *ZoneAPI) List(c *gin.Context) {
	zoneListResp := &ZoneListResponse{}
	c.JSON(http.StatusOK, zoneListResp)
}
