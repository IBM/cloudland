/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/routes"
	"web/src/model"

	"github.com/gin-gonic/gin"
)

var secruleAPI = &SecruleAPI{}
var secruleAdmin = &routes.SecruleAdmin{}

type SecruleAPI struct{}

type SecruleResponse struct {
	*BaseID
	RemoteCIDR    string         `json:"remote_cidr,omitempty"`
	RemoteGroup *BaseReference `json:"remote_group,omitempty"`
	Direction   string         `json:"direction"`
	IpVersion   string         `json:"ip_version"`
	Protocol    string         `json:"protocol"`
	PortMin     int32          `json:"port_min"`
	PortMax     int32          `json:"port_max"`
}

type SecruleListResponse struct {
	Offset        int                `json:"offset"`
	Total         int                `json:"total"`
	Limit         int                `json:"limit"`
	SecurityRules []*SecruleResponse `json:"security_rules"`
}

type SecurityRulePayload struct {
	RemoteCIDR  string         `json:"remote_cidr" binding:"cidrv4"`
	Direction   string         `json:"direction" binding:"required,oneof=ingress egress"`
	Protocol    string         `json:"protocol" binding:"required,oneof=tcp udp icmp"`
	PortMin     int32          `json:"port_min" binding:"omitempty,gte=1,lte=65535"`
	PortMax     int32          `json:"port_max" binding:"omitempty,gte=1,lte=65535"`
}

type SecrulePatchPayload struct {
}

// @Summary get a secrule
// @Description get a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecruleResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id}/rules/{rule_id} [get]
func (v *SecruleAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	sgID := c.Param("id")
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, sgID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid security group query", err)
		return
	}
	ruleID := c.Param("rule_id")
	secrule, err := secruleAdmin.GetSecruleByUUID(ctx, ruleID, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid security rule query", err)
		return
	}
	secruleResp, err := v.getSecruleResponse(ctx, secrule)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, secruleResp)
}

// @Summary delete a secrule
// @Description delete a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id}/rules/{rule_id} [delete]
func (v *SecruleAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	sgID := c.Param("id")
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, sgID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid security group query", err)
		return
	}
	ruleID := c.Param("rule_id")
	secrule, err := secruleAdmin.GetSecruleByUUID(ctx, ruleID, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = secruleAdmin.Delete(ctx, secrule, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a secrule
// @Description create a secrule
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SecurityRulePayload  true   "Secrule create payload"
// @Success 200 {object} SecruleResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id}/rules [post]
func (v *SecruleAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	sgID := c.Param("id")
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, sgID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to get security group", err)
		return
	}
	payload := &SecurityRulePayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	secrule, err := secruleAdmin.Create(ctx, payload.RemoteCIDR, payload.Direction, payload.Protocol, payload.PortMin, payload.PortMax, secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	secruleResp, err := v.getSecruleResponse(ctx, secrule)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, secruleResp)
}

func (v *SecruleAPI) getSecruleResponse(ctx context.Context, secrule *model.SecurityRule) (secruleResp *SecruleResponse, err error) {
	secruleResp = &SecruleResponse{
		BaseID: &BaseID{
			ID: secrule.UUID,
		},
		PortMin:   secrule.PortMin,
		PortMax:   secrule.PortMax,
		Direction: secrule.Direction,
		IpVersion: secrule.IpVersion,
		Protocol:  secrule.Protocol,
	}
	if secrule.RemoteIp != "" {
		secruleResp.RemoteCIDR = secrule.RemoteIp
	} else if secrule.RemoteGroup != nil {
		secruleResp.RemoteGroup = &BaseReference{
			ID:   secrule.RemoteGroup.UUID,
			Name: secrule.RemoteGroup.Name,
		}
	}
	return
}

// @Summary list secrules
// @Description list secrules
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SecruleListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /security_groups/{id}/rules [get]
func (v *SecruleAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
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
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to get security group", err)
		return
	}
	total, secrules, err := secruleAdmin.List(ctx, int64(offset), int64(limit), "-created_at", secgroup)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list secrules", err)
		return
	}
	secruleListResp := &SecruleListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(secrules),
	}
	secruleListResp.SecurityRules = make([]*SecruleResponse, secruleListResp.Limit)
	for i, secrule := range secrules {
		secruleListResp.SecurityRules[i], err = v.getSecruleResponse(ctx, secrule)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, secruleListResp)
}
