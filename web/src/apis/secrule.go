/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var secruleAPI = &SecruleAPI{}
var secruleAdmin = &routes.SecruleAdmin{}

type SecruleAPI struct{}

type SecruleResponse struct {
	*ResourceReference
	RemoteCIDR  string         `json:"remote_cidr,omitempty"`
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
	RemoteCIDR string `json:"remote_cidr" binding:"cidrv4"`
	Direction  string `json:"direction" binding:"required,oneof=ingress egress"`
	Protocol   string `json:"protocol" binding:"required,oneof=tcp udp icmp"`
	PortMin    int32  `json:"port_min" binding:"omitempty,gte=1,lte=65535"`
	PortMax    int32  `json:"port_max" binding:"omitempty,gte=1,lte=65535"`
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
		logger.Errorf("Failed to get security group: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid security group query", err)
		return
	}
	ruleID := c.Param("rule_id")
	logger.Debugf("Get secrule %s of SG %s", ruleID, sgID)
	secrule, err := secruleAdmin.GetSecruleByUUID(ctx, ruleID, secgroup)
	if err != nil {
		logger.Errorf("Failed to get secrule %s, %+v", ruleID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid security rule query", err)
		return
	}
	secruleResp, err := v.getSecruleResponse(ctx, secrule)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Get secrule successfully, %s, %+v", ruleID, secruleResp)
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
		logger.Errorf("Failed to get security group: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid security group query", err)
		return
	}
	ruleID := c.Param("rule_id")
	secrule, err := secruleAdmin.GetSecruleByUUID(ctx, ruleID, secgroup)
	if err != nil {
		logger.Errorf("Failed to get secrule %s, %+v", ruleID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	logger.Debugf("Delete secrule %s of SG %s", ruleID, sgID)
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
	logger.Debugf("Create secrule for SG %s", sgID)
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, sgID)
	if err != nil {
		logger.Errorf("Failed to get security group: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to get security group", err)
		return
	}
	payload := &SecurityRulePayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind json, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	logger.Debugf("Creating secrule with %+v", payload)
	secrule, err := secruleAdmin.Create(ctx, payload.RemoteCIDR, payload.Direction, payload.Protocol, payload.PortMin, payload.PortMax, secgroup)
	if err != nil {
		logger.Errorf("Failed to create secrule, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	secruleResp, err := v.getSecruleResponse(ctx, secrule)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Create secrule successfully for SG %s, %+v", sgID, secruleResp)
	c.JSON(http.StatusOK, secruleResp)
}

func (v *SecruleAPI) getSecruleResponse(ctx context.Context, secrule *model.SecurityRule) (secruleResp *SecruleResponse, err error) {
	secruleResp = &SecruleResponse{
		ResourceReference: &ResourceReference{
			ID:        secrule.UUID,
			CreatedAt: secrule.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: secrule.UpdatedAt.Format(TimeStringForMat),
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
	logger.Debugf("List secrules for SG %s, offset:%s, limit:%s", uuID, offsetStr, limitStr)
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		logger.Errorf("Failed to parse offset: %s, %+v", offsetStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logger.Errorf("Failed to parse limit: %s, %+v", limitStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		errStr := "Invalid query offset or limit, cannot be negative"
		logger.Errorf(errStr)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", errors.New(errStr))
		return
	}
	secgroup, err := secgroupAdmin.GetSecgroupByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get security group: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to get security group", err)
		return
	}
	total, secrules, err := secruleAdmin.List(ctx, int64(offset), int64(limit), "-created_at", secgroup)
	if err != nil {
		logger.Errorf("Failed to list secrules, %+v", err)
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
	logger.Debugf("List secrules successfully for SG %s, %+v", uuID, secruleListResp)
	c.JSON(http.StatusOK, secruleListResp)
}
