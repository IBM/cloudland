/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"fmt"
	"net/http"

	. "web/src/common"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var consoleAPI = &ConsoleAPI{}

type ConsoleAPI struct{}

type ConsoleResponse struct {
	Instance   *ResourceReference `json:"instance"`
	Token      string             `json:"token"`
	ConsoleURL string             `json:"console_url"`
}

// @Summary create a console
// @Description create a console
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   id  path  int  true  "Instance ID"
// @Success 200 {object} ConsoleResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/:id/console [post]
func (v *ConsoleAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	token, err := routes.MakeToken(ctx, instance)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	consoleURL := fmt.Sprintf("wss://%s/websockify?token=%s", c.Request.Host, token)
	owner := orgAdmin.GetOrgName(instance.Owner)
	consoleResp := &ConsoleResponse{
		Instance: &ResourceReference{
			ID:    instance.UUID,
			Owner: owner,
		},
		Token:      token,
		ConsoleURL: consoleURL,
	}
	c.JSON(http.StatusOK, consoleResp)
}
