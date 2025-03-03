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

var (
	versionAPI   = &VersionAPI{}
	sysInfoAdmin = &routes.SysInfoAdmin{}
)

type VersionAPI struct{}

type VersionResponse struct {
	Version string `json:"version"`
}

// @Summary get version
// @Description get version
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} VersionResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /version [get]
func (v *VersionAPI) Get(c *gin.Context) {

	versionResp := &VersionResponse{
		Version: sysInfoAdmin.GetVersion(),
	}
	c.JSON(http.StatusOK, versionResp)
}
