/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var interfaceAPI = &InterfaceAPI{}
var interfaceAdmin = &routes.InterfaceAdmin{}

type InterfaceAPI struct{}

type InterfaceListResponse struct {
	Offset     int                  `json:"offset"`
	Total      int                  `json:"total"`
	Limit      int                  `json:"limit"`
	Interfaces []*InterfaceResponse `json:"interfaces"`
}

type InterfaceResponse struct {
	*BaseReference
	Subnet         *ResourceReference   `json:"subnet"`
	MacAddress     string               `json:"mac_address"`
	IPAddress      string               `json:"ip_address"`
	IsPrimary      bool                 `json:"is_primary"`
	Inbound        int32                `json:"inbound"`
	Outbound       int32                `json:"outbound"`
	FloatingIps    []*FloatingIpInfo    `json:"floating_ips,omitempty"`
	SecurityGroups []*ResourceReference `json:"security_groups,omitempty"`
}

type InterfacePayload struct {
	Subnet         *BaseReference   `json:"subnet" binding:"required"`
	IpAddress      string           `json:"ip_address", binding:"omitempty,ipv4"`
	MacAddress     string           `json:"mac_address" binding:"omitempty,mac"`
	Name           string           `json:"name" binding:"omitempty,min=2,max=32"`
	Inbound        int32            `json:"inbound" binding:"omitempty,min=1,max=20000"`
	Outbound       int32            `json:"outbound" binding:"omitempty,min=1,max=20000"`
	AllowSpoofing  bool             `json:"allow_spoofing" binding:"omitempty"`
	SecurityGroups []*BaseReference `json:"security_group" binding:"omitempty"`
}

type InterfacePatchPayload struct {
	Name           string           `json:"name" binding:"omitempty,min=2,max=32"`
	Inbound        int32            `json:"inbound" binding:"omitempty,min=1,max=20000"`
	Outbound       int32            `json:"outbound" binding:"omitempty,min=1,max=20000"`
	AllowSpoofing  *bool             `json:"allow_spoofing" binding:"omitempty"`
	SecurityGroups []*BaseReference `json:"security_group" binding:"omitempty"`
}

// @Summary get a interface
// @Description get a interface
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} InterfaceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/id/interfaces/{interface_id} [get]
func (v *InterfaceAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Patch instance interface %s", uuID)
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get instance %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	ifaceID := c.Param("interface_id")
	iface, err := interfaceAdmin.GetInterfaceByUUID(ctx, ifaceID)
	if err != nil {
		logger.Errorf("Failed to get interface %s, %+v", ifaceID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid interface query", err)
		return
	}
	interfaceResp, err := v.getInterfaceResponse(ctx, instance, iface)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Get interface successfully, %s, %+v", ifaceID, interfaceResp)
	c.JSON(http.StatusOK, interfaceResp)
}

func (v *InterfaceAPI) getInterfaceResponse(ctx context.Context, instance *model.Instance, iface *model.Interface) (interfaceResp *InterfaceResponse, err error) {
	interfaceResp = &InterfaceResponse{
		BaseReference: &BaseReference{
			ID:   iface.UUID,
			Name: iface.Name,
		},
		MacAddress: iface.MacAddr,
		IPAddress:  iface.Address.Address,
		IsPrimary:  iface.PrimaryIf,
		Inbound:    iface.Inbound,
		Outbound:   iface.Outbound,
		Subnet: &ResourceReference{
			ID:   iface.Address.Subnet.UUID,
			Name: iface.Address.Subnet.Name,
		},
	}
	if iface.PrimaryIf && len(instance.FloatingIps) > 0 {
		floatingIps := make([]*FloatingIpInfo, len(instance.FloatingIps))
		for i, floatingip := range instance.FloatingIps {
			floatingIps[i] = &FloatingIpInfo{
				ResourceReference: &ResourceReference{
					ID:   floatingip.UUID,
					Name: floatingip.Name,
				},
				IpAddress: floatingip.FipAddress,
			}
		}
		interfaceResp.FloatingIps = floatingIps
	}
	for _, sg := range iface.SecurityGroups {
		interfaceResp.SecurityGroups = append(interfaceResp.SecurityGroups, &ResourceReference{
			ID:   sg.UUID,
			Name: sg.Name,
		})
	}
	return
}

// @Summary patch a interface
// @Description patch a interface
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   InterfacePatchPayload  true   "Interface patch payload"
// @Success 200 {object} InterfaceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/id/interfaces/{interface_id} [patch]
func (v *InterfaceAPI) Patch(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Patch instance interface %s", uuID)
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get instance %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	ifaceID := c.Param("interface_id")
	iface, err := interfaceAdmin.GetInterfaceByUUID(ctx, ifaceID)
	if err != nil {
		logger.Errorf("Failed to get interface %s, %+v", ifaceID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid interface query", err)
		return
	}
	payload := &InterfacePatchPayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind JSON, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	ifaceName := iface.Name
	if payload.Name != "" {
		ifaceName = payload.Name
		logger.Debugf("Update interface name to %s", ifaceName)
	}
	inbound := iface.Inbound
	if payload.Inbound > 0 {
		inbound = payload.Inbound
	}
	outbound := iface.Outbound
	if payload.Outbound > 0 {
		outbound = payload.Outbound
	}
	allowSpoofing := iface.AllowSpoofing
	if payload.AllowSpoofing != nil {
		allowSpoofing = *payload.AllowSpoofing
	}
	secgroups := []*model.SecurityGroup{}
	if len(payload.SecurityGroups) == 0 {
		for _, sg := range payload.SecurityGroups {
			var secgroup *model.SecurityGroup
			secgroup, err = secgroupAdmin.GetSecurityGroup(ctx, sg)
			if err != nil {
				logger.Errorf("Get security group failed, %+v", err)
				ErrorResponse(c, http.StatusBadRequest, "Invalid security group", err)
				return
			}
			if secgroup.RouterID != instance.RouterID {
				err = fmt.Errorf("Security group not in instance vpc")
				ErrorResponse(c, http.StatusBadRequest, "Invalid security group", err)
				return
			}
			secgroups = append(secgroups, secgroup)
		}
	}
	err = interfaceAdmin.Update(ctx, instance, iface, ifaceName, inbound, outbound, allowSpoofing, secgroups)
	if err != nil {
		logger.Errorf("Patch instance failed, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Patch instance failed", err)
		return
	}
	interfaceResp, err := v.getInterfaceResponse(ctx, instance, iface)
	if err != nil {
		logger.Errorf("Get interface responsefailed, %+v", err)
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, interfaceResp)
}

// @Summary delete a interface
// @Description delete a interface
// @tags Network
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instance/{id}/interfaces/{id} [delete]
func (v *InterfaceAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a interface
// @Description create a interface
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   InterfacePayload  true   "Interface create payload"
// @Success 200 {object} InterfaceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instance/{id}/interfaces [post]
func (v *InterfaceAPI) Create(c *gin.Context) {
	interfaceResp := &InterfaceResponse{}
	c.JSON(http.StatusOK, interfaceResp)
}

// @Summary list interfaces
// @Description list interfaces
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {array} InterfaceResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instance/{id}/interfaces [get]
func (v *InterfaceAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	logger.Debugf("List interfaces for instance %s, offset:%s, limit:%s", uuID, offsetStr, limitStr)
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
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get instance: %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to get instance", err)
		return
	}
	total, interfaces, err := interfaceAdmin.List(ctx, int64(offset), int64(limit), "-created_at", instance)
	if err != nil {
		logger.Errorf("Failed to list interfaces, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to list secrules", err)
		return
	}
	interfaceListResp := &InterfaceListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(interfaces),
	}
	interfaceListResp.Interfaces = make([]*InterfaceResponse, interfaceListResp.Limit)
	for i, iface := range interfaces {
		interfaceListResp.Interfaces[i], err = v.getInterfaceResponse(ctx, instance, iface)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	logger.Debugf("List secrules successfully for SG %s, %+v", uuID, interfaceListResp)
	c.JSON(http.StatusOK, interfaceListResp)
}
