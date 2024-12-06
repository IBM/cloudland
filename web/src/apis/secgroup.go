/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var secgroupAPI = &SecgroupAPI{}
var secgroupAdmin = &routes.SecgroupAdmin{}

type SecgroupAPI struct{}

type SecurityGroupResponse struct {
	*BaseReference
	IsDefault        bool               `json:"is_default"`
	VPC              *BaseReference     `json:"vpc,omitempty"`
	TargetInterfaces []*TargetInterface `json:"target_interfaces,omitempty"`
}

type SecurityGroupListResponse struct {
	Offset         int                      `json:"offset"`
	Total          int                      `json:"total"`
	Limit          int                      `json:"limit"`
	SecurityGroups []*SecurityGroupResponse `json:"security_groups"`
}

type SecurityGroupPayload struct {
	Name      string         `json:"name" binding:"required,min=2,max=32"`
	VPC       *BaseReference `json:"vpc" binding:"omitempty"`
	IsDefault bool           `json:"is_default" binding:"omitempty"`
}

type SecurityGroupPatchPayload struct {
	Name      string `json:"name" binding:"required,min=2,max=32"`
	IsDefault bool   `json:"is_default" binding:"omitempty"`
}

// @Summary get a secgroup
// @Description get a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecurityGroupResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id} [get]
func (v *SecgroupAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid vpc query", err)
		return
	}
	secgroupResp, err := v.getSecgroupResponse(ctx, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, secgroupResp)
}

// @Summary patch a secgroup
// @Description patch a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecurityGroupPatchPayload  true   "Secgroup patch payload"
// @Success 200 {object} SecurityGroupResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id} [patch]
func (v *SecgroupAPI) Patch(c *gin.Context) {
	secgroupResp := &SecurityGroupResponse{}
	c.JSON(http.StatusOK, secgroupResp)
}

// @Summary delete a secgroup
// @Description delete a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id} [delete]
func (v *SecgroupAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = secgroupAdmin.Delete(ctx, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a secgroup
// @Description create a secgroup
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecurityGroupPayload  true   "Secgroup create payload"
// @Success 200 {object} SecurityGroupResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups [post]
func (v *SecgroupAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &SecurityGroupPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	var router *model.Router
	if payload.VPC != nil {
		router, err = routerAdmin.GetRouter(ctx, payload.VPC)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Failed to get vpc", err)
			return
		}
	}
	secgroup, err := secgroupAdmin.Create(ctx, payload.Name, payload.IsDefault, router)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	secgroupResp, err := v.getSecgroupResponse(ctx, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, secgroupResp)
}

func (v *SecgroupAPI) getSecgroupResponse(ctx context.Context, secgroup *model.SecurityGroup) (secgroupResp *SecurityGroupResponse, err error) {
	secgroupResp = &SecurityGroupResponse{
		BaseReference: &BaseReference{
			ID:   secgroup.UUID,
			Name: secgroup.Name,
		},
		IsDefault: secgroup.IsDefault,
	}
	if secgroup.Router != nil {
		secgroupResp.VPC = &BaseReference{
			ID:   secgroup.Router.UUID,
			Name: secgroup.Router.Name,
		}
	}
	err = secgroupAdmin.GetSecgroupInterfaces(ctx, secgroup)
	if err != nil {
		return
	}
	for _, iface := range secgroup.Interfaces {
		targetIface := &TargetInterface{
			BaseID: &BaseID{
				ID: iface.UUID,
			},
		}
		if iface.Address != nil {
			targetIface.IpAddress = strings.Split(iface.Address.Address, "/")[0]
		}
		if iface.Instance > 0 {
			var instance *model.Instance
			instance, err = instanceAdmin.Get(ctx, iface.Instance)
			if err != nil {
				return
			}
			targetIface.FromInstance = &InstanceInfo{
				BaseID: &BaseID{
					ID: instance.UUID,
				},
				Hostname: instance.Hostname,
			}
		}
		secgroupResp.TargetInterfaces = append(secgroupResp.TargetInterfaces, targetIface)
	}
	return
}

// @Summary list secgroups
// @Description list secgroups
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecurityGroupListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups [get]
func (v *SecgroupAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, secgroups, err := secgroupAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list secgroups", err)
		return
	}
	secgroupListResp := &SecurityGroupListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(secgroups),
	}
	secgroupListResp.SecurityGroups = make([]*SecurityGroupResponse, secgroupListResp.Limit)
	for i, secgroup := range secgroups {
		secgroupListResp.SecurityGroups[i], err = v.getSecgroupResponse(ctx, secgroup)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, secgroupListResp)
}
