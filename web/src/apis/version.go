/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	VersionFile = "/opt/cloudland/version"
)

var (
	Version    = "unknown"
	versionAPI = &VersionAPI{}
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
	if Version == "unknown" {
		// read version from file
		// check if file exists
		// if not, return default version
		// if yes, read version from file
		// return version
		version, err := os.ReadFile(VersionFile)
		if err != nil {
			logger.Warningf("failed to read version file: %v", err)
		} else {
			Version = string(version)
		}
	}
	versionResp := &VersionResponse{
		Version: Version,
	}
	c.JSON(http.StatusOK, versionResp)
}
