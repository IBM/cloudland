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

var floatingIpAPI = &FloatingIpAPI{}
var floatingIpAdmin = &routes.FloatingIpAdmin{}

type FloatingIpAPI struct{}

type FloatingIpInfo struct {
	*BaseReference
	IpAddress string `json:"ip_address"`
}

type FloatingIpResponse struct {
	*BaseReference
	PublicIP  string         `json:"public_ip"`
	PrivateIP string         `json:"private_ip"`
	Instance  *BaseReference `json:"instance"`
	Interface *BaseReference `json:"interface"`
	VPC       *BaseReference `json:"vpc"`
}

type FloatingIpListResponse struct {
	Offset      int                   `json:"offset"`
	Total       int                   `json:"total"`
	Limit       int                   `json:"limit"`
	FloatingIps []*FloatingIpResponse `json:"floatingIps"`
}

type FloatingIpPayload struct {
	PublicSubnet *BaseReference `json:"public_subnet" binding:"omitempty"`
	PublicIP string  `json:"public_ip" binding:"omitempty,ipv4"`
	Instance *BaseID `json:"instance" binding:"omitempty"`
	Bandwidth int64  `json:"bandwidth" binding:"omitempty"`
}

type FloatingIpPatchPayload struct {
	Instance *BaseID `json:"instance" binding:"omitempty"`
}

// @Summary get a floating ip
// @Description get a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} FloatingIpResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips/{id} [get]
func (v *FloatingIpAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	floatingIp, err := floatingIpAdmin.GetFloatingIpByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	floatingIpResp, err := v.getFloatingIpResponse(ctx, floatingIp)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, floatingIpResp)
}

// @Summary patch a floating ip
// @Description patch a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   FloatingIpPatchPayload  true   "Floating ip patch payload"
// @Success 200 {object} FloatingIpResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips/{id} [patch]
func (v *FloatingIpAPI) Patch(c *gin.Context) {
	floatingIpResp := &FloatingIpResponse{}
	c.JSON(http.StatusOK, floatingIpResp)
}

// @Summary delete a floating ip
// @Description delete a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips/{id} [delete]
func (v *FloatingIpAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	floatingIp, err := floatingIpAdmin.GetFloatingIpByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = floatingIpAdmin.Delete(ctx, floatingIp)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a floating ip
// @Description create a floating ip
// @tags Network
// @Accept  json
// @Produce json
// @Param   message	body   FloatingIpPayload  true   "Floating ip create payload"
// @Success 200 {object} FloatingIpResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips [post]
func (v *FloatingIpAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &FloatingIpPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	var publicSubnet *model.Subnet
	if payload.PublicSubnet != nil {
		publicSubnet, err = subnetAdmin.GetSubnet(ctx, payload.PublicSubnet)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Failed to get public subnet", err)
			return
		}
	}
	var instance *model.Instance
	if payload.Instance != nil {
		instance, err = instanceAdmin.GetInstanceByUUID(ctx, payload.Instance.ID)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Failed to get instance", err)
			return
		}
	}
	floatingIp, err := floatingIpAdmin.Create(ctx, instance, publicSubnet, payload.PublicIP)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to create floating ip", err)
		return
	}
	floatingIpResp, err := v.getFloatingIpResponse(ctx, floatingIp)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, floatingIpResp)
}

func (v *FloatingIpAPI) getFloatingIpResponse(ctx context.Context, floatingIp *model.FloatingIp) (floatingIpResp *FloatingIpResponse, err error) {
	floatingIpResp = &FloatingIpResponse{
		BaseReference: &BaseReference{
			ID: floatingIp.UUID,
		},
		PublicIP:  floatingIp.IPAddress,
		PrivateIP: floatingIp.IntAddress,
	}
	if floatingIp.Router != nil {
		floatingIpResp.VPC = &BaseReference{
			ID:   floatingIp.Router.UUID,
			Name: floatingIp.Router.Name,
		}
	}
	if floatingIp.Instance != nil {
		floatingIpResp.Instance = &BaseReference{
			ID: floatingIp.Instance.UUID,
		}
	}
	if floatingIp.Interface != nil {
		floatingIpResp.Interface = &BaseReference{
			ID:   floatingIp.Interface.UUID,
			Name: floatingIp.Interface.Name,
		}
	}
	return
}

// @Summary list floating ips
// @Description list floating ips
// @tags Network
// @Accept  json
// @Produce json
// @Success 200 {object} FloatingIpListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /floating_ips [get]
func (v *FloatingIpAPI) List(c *gin.Context) {
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
	total, floatingIps, err := floatingIpAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list floatingIps", err)
		return
	}
	floatingIpListResp := &FloatingIpListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(floatingIps),
	}
	floatingIpListResp.FloatingIps = make([]*FloatingIpResponse, floatingIpListResp.Limit)
	for i, floatingIp := range floatingIps {
		floatingIpListResp.FloatingIps[i], err = v.getFloatingIpResponse(ctx, floatingIp)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, floatingIpListResp)
}
