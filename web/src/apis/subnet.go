/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"net/http"
	"strconv"

	"web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var subnetAPI = &SubnetAPI{}
var subnetAdmin = &routes.SubnetAdmin{}

type SubnetAPI struct{}

type SubnetResponse struct {
	*common.BaseReference
	Network    string                `json:"network"`
	Netmask    string                `json:"netmask"`
	Gateway    string                `json:"gateway"`
	NameServer string                `json:"dns,omitempty"`
	VPC        *common.BaseReference `json:"vpc,omitempty"`
	Type       string                `json:"type"`
}

type SubnetListResponse struct {
	Offset  int               `json:"offset"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Subnets []*SubnetResponse `json:"subnets"`
}

type SubnetPayload struct {
}

type SubnetPatchPayload struct {
}

// @Summary get a subnet
// @Description get a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SubnetResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /subnets/{id} [get]
func (v *SubnetAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	subnet, err := subnetAdmin.GetSubnetByUUID(ctx, uuID)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid subnet query", err)
		return
	}
	subnetResp, err := getSubnetResponse(ctx, subnet)
	if err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, subnetResp)
}

// @Summary patch a subnet
// @Description patch a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SubnetPatchPayload  true   "Subnet patch payload"
// @Success 200 {object} SubnetResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /subnets/{id} [patch]
func (v *SubnetAPI) Patch(c *gin.Context) {
	subnetResp := &SubnetResponse{}
	c.JSON(http.StatusOK, subnetResp)
}

// @Summary delete a subnet
// @Description delete a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /subnets/{id} [delete]
func (v *SubnetAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	subnet, err := subnetAdmin.GetSubnetByUUID(ctx, uuID)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = subnetAdmin.Delete(ctx, subnet)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a subnet
// @Description create a subnet
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   SubnetPayload  true   "Subnet create payload"
// @Success 200 {object} SubnetResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /subnets [post]
func (v *SubnetAPI) Create(c *gin.Context) {
	subnetResp := &SubnetResponse{}
	c.JSON(http.StatusOK, subnetResp)
}

func getSubnetResponse(ctx context.Context, subnet *model.Subnet) (subnetResp *SubnetResponse, err error) {
	subnetResp = &SubnetResponse{
		BaseReference: &common.BaseReference{
			ID:   subnet.UUID,
			Name: subnet.Name,
		},
		Network:    subnet.Network,
		Netmask:    subnet.Netmask,
		Gateway:    subnet.Gateway,
		NameServer: subnet.NameServer,
		Type:       subnet.Type,
	}
	if subnet.Router != nil {
		subnetResp.VPC = &common.BaseReference{
			ID:   subnet.Router.UUID,
			Name: subnet.Router.Name,
		}
	}
	return
}

// @Summary list subnets
// @Description list subnets
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} SubnetListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /subnets [get]
func (v *SubnetAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, subnets, err := subnetAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Failed to list subnets", err)
		return
	}
	subnetListResp := &SubnetListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(subnets),
	}
	subnetListResp.Subnets = make([]*SubnetResponse, subnetListResp.Limit)
	for i, subnet := range subnets {
		subnetListResp.Subnets[i], err = getSubnetResponse(ctx, subnet)
		if err != nil {
			common.ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, subnetListResp)
}
