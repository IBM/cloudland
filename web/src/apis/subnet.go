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
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var subnetAPI = &SubnetAPI{}
var subnetAdmin = &routes.SubnetAdmin{}

type SubnetAPI struct{}

type SubnetResponse struct {
	*BaseReference
	Network    string                `json:"network"`
	Netmask    string                `json:"netmask"`
	Gateway    string                `json:"gateway"`
	NameServer string                `json:"dns,omitempty"`
	VPC        *BaseReference `json:"vpc,omitempty"`
	Type       SubnetType `json:"type"`
}

type SubnetListResponse struct {
	Offset  int               `json:"offset"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Subnets []*SubnetResponse `json:"subnets"`
}

type SubnetPayload struct {
	Name        string                `json:"name" binding:"required,min=2,max=32"`
	NetworkCIDR string                `json:"network_cidr" binding:"required,cidrv4"`
	Gateway     string                `json:"gateway" binding:"omitempty,ipv4"`
	StartIP     string                `json:"start_ip" binding:"omitempty,ipv4"`
	EndIP       string                `json:"end_ip" binding:"omitempty",ipv4`
	NameServer  string                `json:"dns" binding:"omitempty"`
	BaseDomain  string                `json:"base_domain" binding:"omitempty"`
	Dhcp        bool                  `json:"dhcp" binding:"omitempty"`
	VPC         *BaseReference `json:"vpc" binding:"omitempty"`
	Vlan        int `json:"vlan" binding:"omitempty,gte=1,lte=16777215"`
	Type        SubnetType `json:"type" binding:"omitempty,oneof=public internal"`
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
		ErrorResponse(c, http.StatusBadRequest, "Invalid subnet query", err)
		return
	}
	subnetResp, err := v.getSubnetResponse(ctx, subnet)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
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
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = subnetAdmin.Delete(ctx, subnet)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
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
	ctx := c.Request.Context()
	payload := &SubnetPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	if payload.VPC == nil && payload.Type != Public {
		ErrorResponse(c, http.StatusBadRequest, "VPC must be specified if network type not public", err)
		return
	}
	var router *model.Router
	if payload.VPC != nil {
		router, err = routerAdmin.GetRouter(ctx, payload.VPC)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Failed to get router", err)
			return
		}
	}
	subnet, err := subnetAdmin.Create(ctx, payload.Vlan, payload.Name, payload.NetworkCIDR, payload.Gateway, payload.StartIP, payload.EndIP, string(payload.Type), payload.NameServer, payload.BaseDomain, payload.Dhcp, router)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to create subnet", err)
		return
	}
	subnetResp, err := v.getSubnetResponse(ctx, subnet)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, subnetResp)
}

func (v *SubnetAPI) getSubnetResponse(ctx context.Context, subnet *model.Subnet) (subnetResp *SubnetResponse, err error) {
	subnetResp = &SubnetResponse{
		BaseReference: &BaseReference{
			ID:   subnet.UUID,
			Name: subnet.Name,
		},
		Network:    subnet.Network,
		Netmask:    subnet.Netmask,
		Gateway:    subnet.Gateway,
		NameServer: subnet.NameServer,
		Type:       SubnetType(subnet.Type),
	}
	if subnet.Router != nil {
		subnetResp.VPC = &BaseReference{
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
	total, subnets, err := subnetAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list subnets", err)
		return
	}
	subnetListResp := &SubnetListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(subnets),
	}
	subnetListResp.Subnets = make([]*SubnetResponse, subnetListResp.Limit)
	for i, subnet := range subnets {
		subnetListResp.Subnets[i], err = v.getSubnetResponse(ctx, subnet)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, subnetListResp)
}
