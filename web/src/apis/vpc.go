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

var vpcAPI = &VPCAPI{}
var routerAdmin = &routes.RouterAdmin{}

type VPCAPI struct{}

type VPCResponse struct {
	*ResourceReference
	Subnets []*SubnetResponse `json:"subnets,omitempty"`
}

type VPCListResponse struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	VPCs   []*VPCResponse `json:"vpcs"`
}

type VPCPayload struct {
	Name string `json:"name" binding:"required,min=2,max=32"`
}

type VPCPatchPayload struct {
}

// @Summary get a vpc
// @Description get a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} VPCResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs/{id} [get]
func (v *VPCAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Get vpc by uuid: %s", uuID)
	router, err := routerAdmin.GetRouterByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get vpc by uuid: %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid vpc query", err)
		return
	}
	vpcResp, err := v.getVPCResponse(ctx, router)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Get vpc by uuid: %s, %+v", uuID, vpcResp)
	c.JSON(http.StatusOK, vpcResp)
}

// @Summary patch a vpc
// @Description patch a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   VPCPatchPayload  true   "VPC patch payload"
// @Success 200 {object} VPCResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs/{id} [patch]
func (v *VPCAPI) Patch(c *gin.Context) {
	vpcResp := &VPCResponse{}
	c.JSON(http.StatusOK, vpcResp)
}

// @Summary delete a vpc
// @Description delete a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs/{id} [delete]
func (v *VPCAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Delete vpc by uuid: %s", uuID)
	router, err := routerAdmin.GetRouterByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get vpc by uuid: %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = routerAdmin.Delete(ctx, router)
	if err != nil {
		logger.Errorf("Failed to delete vpc by uuid: %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	logger.Debugf("Deleted vpc by uuid: %s", uuID)
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a vpc
// @Description create a vpc
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   VPCPayload  true   "VPC create payload"
// @Success 200 {object} VPCResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs [post]
func (v *VPCAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &VPCPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind json: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	logger.Debugf("Creating vpc with %+v", payload)
	router, err := routerAdmin.Create(ctx, payload.Name)
	if err != nil {
		logger.Errorf("Failed to create vpc: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to create vpc", err)
		return
	}
	vpcResp, err := v.getVPCResponse(ctx, router)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Create vpc successfully, %+v", vpcResp)
	c.JSON(http.StatusOK, vpcResp)
}

func (v *VPCAPI) getVPCResponse(ctx context.Context, router *model.Router) (vpcResp *VPCResponse, err error) {
	owner := orgAdmin.GetOrgName(router.Owner)
	vpcResp = &VPCResponse{
		ResourceReference: &ResourceReference{
			ID:        router.UUID,
			Name:      router.Name,
			Owner:     owner,
			CreatedAt: router.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: router.UpdatedAt.Format(TimeStringForMat),
		},
	}
	vpcResp.Subnets = make([]*SubnetResponse, len(router.Subnets))
	for i, subnet := range router.Subnets {
		vpcResp.Subnets[i], err = subnetAPI.getSubnetResponse(ctx, subnet)
		if err != nil {
			return
		}
	}
	return
}

// @Summary list vpcs
// @Description list vpcs
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} VPCListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /vpcs [get]
func (v *VPCAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	queryStr := c.DefaultQuery("query", "")
	logger.Debugf("List vpcs, offset:%s, limit:%s, query:%s", offsetStr, limitStr, queryStr)
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		logger.Errorf("Invalid query offset: %s, %+v", offsetStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logger.Errorf("Invalid query limit: %s, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		errStr := "Invalid query offset or limit, cannot be negative"
		logger.Errorf(errStr)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", errors.New(errStr))
		return
	}
	total, routers, err := routerAdmin.List(ctx, int64(offset), int64(limit), "-created_at", queryStr)
	if err != nil {
		logger.Errorf("Failed to list vpcs, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to list vpcs", err)
		return
	}
	vpcListResp := &VPCListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(routers),
	}
	vpcListResp.VPCs = make([]*VPCResponse, vpcListResp.Limit)
	for i, router := range routers {
		vpcListResp.VPCs[i], err = v.getVPCResponse(ctx, router)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	logger.Debugf("List vpcs successfully, %+v", vpcListResp)
	c.JSON(http.StatusOK, vpcListResp)
}
