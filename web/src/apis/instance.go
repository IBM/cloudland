/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var instanceAPI = &InstanceAPI{}
var instanceAdmin = &routes.InstanceAdmin{}

type InstanceAPI struct{}

type BaseReference struct {
	ID   string `json:"id,required"`
	Name string `json:"name,omitempty"`
}

type PowerAction string

const (
	Stop        PowerAction = "stop"
	HardStop    PowerAction = "hard_stop"
	Start       PowerAction = "start"
	Restart     PowerAction = "restart"
	HardRestart PowerAction = "hard_restart"
	Pause       PowerAction = "pause"
	Unpause     PowerAction = "unpause"
)

type InstancePatchPayload struct {
	Hostname    string       `json:"hostname,omitempty"`
	PowerAction *PowerAction `json:"power_action,omitempty"`
}

type InstancePayload struct {
	Hostname   string              `json:"hostname,required"`
	Keys       []string            `json:"keys,required"`
	Flavor     string              `json:"flavor,required"`
	Image      string              `json:"image,required"`
	Interfaces []*InterfacePayload `json:"interfaces,required"`
	Zone       string              `json:"zone,required"`
}

type InstanceResponse struct {
	ID         string               `json:"id"`
	Hostname   string               `json:"hostname"`
	Status     string               `json:"status"`
	Interfaces []*InterfaceResponse `json:"interfaces"`
	Flavor     *BaseReference       `json:"flavor"`
	Image      *BaseReference       `json:"image"`
	Keys       []*BaseReference     `json:"keys"`
	Zone       string               `json:"zone"`
	VPC        *BaseReference       `json:"vpc,omitempty"`
}

type InstanceListResponse struct {
	Offset    int                 `json:"offset"`
	Total     int                 `json:"total"`
	Limit     int                 `json:"limit"`
	Instances []*InstanceResponse `json:"instances"`
}

//
// @Summary get a instance
// @Description get a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   id  path  int  true  "Instance ID"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances/{id} [get]
func (v *InstanceAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
	}
	instanceResp, err := v.getInstanceResponse(ctx, instance)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
	}

	c.JSON(http.StatusOK, instanceResp)
}

//
// @Summary patch a instance
// @Description patch a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   InstancePatchPayload  true   "Instance patch payload"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances/{id} [patch]
func (v *InstanceAPI) Patch(c *gin.Context) {
	instanceResp := &InstanceResponse{}
	c.JSON(http.StatusOK, instanceResp)
}

//
// @Summary delete a instance
// @Description delete a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   id  path  int  true  "Instance ID"
// @Success 200
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances/{id} [delete]
func (v *InstanceAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a instance
// @Description create a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   InstancePayload  true   "Instance create payload"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances [post]
func (v *InstanceAPI) Create(c *gin.Context) {
	instanceResp := &InstanceResponse{}
	c.JSON(http.StatusOK, instanceResp)
}

func (v *InstanceAPI) getInstanceResponse(ctx context.Context, instance *model.Instance) (instanceResp *InstanceResponse, err error) {
	instanceResp = &InstanceResponse{
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
	for i, key := range instance.Keys {
		keys[i] = &BaseReference{
			ID:   key.UUID,
			Name: key.Name,
		}
	}
	instanceResp.Keys = keys
	interfaces := make([]*InterfaceResponse, len(instance.Interfaces))
	for i, iface := range instance.Interfaces {
		interfaces[i] = &InterfaceResponse{
			BaseReference: &BaseReference{
				ID:   iface.UUID,
				Name: iface.Name,
			},
			MacAddress: iface.MacAddr,
			IPAddress:  iface.Address.Address,
			IsPrimary:  iface.PrimaryIf,
			Subnet: &BaseReference{
				ID:   iface.Address.Subnet.UUID,
				Name: iface.Address.Subnet.Name,
			},
		}
		if iface.PrimaryIf && len(instance.FloatingIps) > 0 {
			floatingIps := make([]*FloatingIpInfo, len(instance.FloatingIps))
			for i, floatingip := range instance.FloatingIps {
				floatingIps[i] = &FloatingIpInfo{
					BaseReference: &BaseReference{
						ID: floatingip.UUID,
					},
					IpAddress: floatingip.FipAddress,
				}
			}
			interfaces[i].FloatingIps = floatingIps
		}
	}
	instanceResp.Interfaces = interfaces
	if instance.RouterID > 0 {
		router, err := routerAdmin.Get(ctx, instance.RouterID)
		if err != nil {
			err = fmt.Errorf("Failed to get VPC")
		}
		instanceResp.VPC = &BaseReference{
			ID:   router.UUID,
			Name: router.Name,
		}
	}
	return
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
		instanceList[i], err = v.getInstanceResponse(ctx, instance)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		}
	}
	instanceListResp.Instances = instanceList
	c.JSON(http.StatusOK, instanceListResp)
	return
}
