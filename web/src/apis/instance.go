/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"log"
	"net/http"
	"strconv"

	"github.com/IBM/cloudland/web/src/routes"
	"github.com/gin-gonic/gin"
)

var instanceAPI = &InstanceAPI{}
var instanceAdmin = &routes.InstanceAdmin{}

type InstanceAPI struct{}

type BaseReference struct {
	ID   string `json:"id,required"`
	Name string `json:"name,omitempty"`
}

type InstancePayload struct {
}

type InstanceResponse struct {
	ID         string           `json:"id"`
	Hostname   string           `json:"hostname"`
	Status     string           `json:"status"`
	Interfaces []*InterfaceInfo `json:"interfaces"`
	Flavor     *BaseReference   `json:"flavor"`
	Image      *BaseReference   `json:"image"`
	Keys       []*BaseReference `json:"keys"`
	Zone       string           `json:"zone"`
	VPC        *BaseReference   `json:"vpc,omitempty"`
}

type InstanceListResponse struct {
	Offset    int
	Total     int
	Limit     int
	Instances []*InstanceResponse
}

//
// @Summary create a instance
// @Description create a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   instancePayload  true   "Instance request payload"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances [get]
func (v *InstanceAPI) Create(c *gin.Context) {
	instanceResp := &InstanceResponse{}
	c.JSON(http.StatusOK, instanceResp)
}

//
// @Summary list instances
// @Description list instances
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} InstanceListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances [get]
func (v *InstanceAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	memberShip := routes.GetMemberShip(ctx)
	log.Printf("Membership: %v\n", memberShip)
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
	total, instances, err := instanceAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list instances", err)
		return
	}
	instanceListResp := &InstanceListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(instances),
	}
	instanceList := make([]*InstanceResponse, instanceListResp.Limit)
	for i, instance := range instances {
		instanceList[i] = &InstanceResponse{
			ID:       instance.UUID,
			Hostname: instance.Hostname,
			Status:   instance.Status,
			Flavor: &BaseReference{
				ID:   instance.Flavor.UUID,
				Name: instance.Flavor.Name,
			},
			Image: &BaseReference{
				ID:   instance.Image.UUID,
				Name: instance.Image.Name,
			},
			Zone: instance.Zone.Name,
		}
		keys := make([]*BaseReference, len(instance.Keys))
		for j, key := range instance.Keys {
			keys[j] = &BaseReference{
				ID:   key.UUID,
				Name: key.Name,
			}
		}
		instanceList[i].Keys = keys
		interfaces := make([]*InterfaceInfo, len(instance.Interfaces))
		for j, iface := range instance.Interfaces {
			interfaces[j] = &InterfaceInfo{
				BaseReference: &BaseReference{
					ID:   iface.UUID,
					Name: iface.Name,
				},
				MacAddress: iface.MacAddr,
				IPAddress:  iface.Address.Address,
				IsPrimary: iface.PrimaryIf,
			}
			if iface.PrimaryIf && len(instance.FloatingIps) > 0{
				floatingIps := make([]*FloatingIpInfo, len(instance.FloatingIps))
				for j, floatingip := range instance.FloatingIps {
					floatingIps[j] = &FloatingIpInfo{
						BaseReference: &BaseReference{
							ID: floatingip.UUID,
						},
						IpAddress: floatingip.FipAddress,
					}
				}
				interfaces[j].FloatingIps = floatingIps
			}
		}
		instanceList[i].Interfaces = interfaces
		if instance.RouterID > 0 {
			router, err := routerAdmin.Get(ctx, instance.RouterID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, &APIError{ErrorMessage: "Failed to get VPC info"})
			}
			instanceList[i].VPC = &BaseReference{
				ID:   router.UUID,
				Name: router.Name,
			}
		}
	}
	instanceListResp.Instances = instanceList
	c.JSON(http.StatusOK, instanceListResp)
	return
}
