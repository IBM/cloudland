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
	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var instanceAPI = &InstanceAPI{}
var instanceAdmin = &routes.InstanceAdmin{}

type InstanceAPI struct{}

type MigrateAction struct {
	FromHypervisor string `json:"from_hypervisor" binding:"omitempty"`
	ToHypervisor   string `json:"to_hypervisor" binding:"required"`
}

type InstancePatchPayload struct {
	Hostname      string        `json:"hostname" binding:"omitempty,hostname|fqdn"`
	PowerAction   PowerAction   `json:"power_action" binding:"omitempty,oneof=stop hard_stop start restart hard_restart pause resume"`
	MigrateAction MigrateAction `json:"migrate_action" binding:"omitempty"`
	Flavor        string        `json:"flavor" binding:"omitempty,min=1,max=32"`
}

type InstancePayload struct {
	Count               int                 `json:"count" binding:"omitempty,gte=1,lte=16"`
	Hypervisor          int                 `json:"hypervisor,default=-1" binding:"omitempty"`
	Hostname            string              `json:"hostname" binding:"required,hostname|fqdn"`
	Keys                []*BaseReference    `json:"keys" binding:"required,gte=1,lte=16"`
	Flavor              string              `json:"flavor" binding:"required,min=1,max=32"`
	Image               *BaseReference      `json:"image" binding:"required"`
	PrimaryInterface    *InterfacePayload   `json:"primary_interface", binding:"required"`
	SecondaryInterfaces []*InterfacePayload `json:"secondary_interfaces" binding:"omitempty"`
	Zone                string              `json:"zone" binding:"required,min=1,max=32"`
	VPC                 *BaseReference      `json:"vpc" binding:"omitempty"`
	Userdata            string              `json:"userdata,omitempty"`
}

type InstanceResponse struct {
	ID         string               `json:"id"`
	Hostname   string               `json:"hostname"`
	Status     string               `json:"status"`
	Interfaces []*InterfaceResponse `json:"interfaces"`
	Flavor     string               `json:"flavor"`
	Image      *BaseReference       `json:"image"`
	Keys       []*BaseReference     `json:"keys"`
	Zone       string               `json:"zone"`
	VPC        *BaseReference       `json:"vpc,omitempty"`
	Hypervisor string               `json:"hypervisor,omitempty"`
}

type InstanceListResponse struct {
	Offset    int                 `json:"offset"`
	Total     int                 `json:"total"`
	Limit     int                 `json:"limit"`
	Instances []*InstanceResponse `json:"instances"`
}

// @Summary get a instance
// @Description get a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   id  path  int  true  "Instance ID"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [get]
func (v *InstanceAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	instanceResp, err := v.getInstanceResponse(ctx, instance)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}

	c.JSON(http.StatusOK, instanceResp)
}

// @Summary patch a instance
// @Description patch a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   InstancePatchPayload  true   "Instance patch payload"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [patch]
func (v *InstanceAPI) Patch(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	payload := &InstancePatchPayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	hostname := instance.Hostname
	if payload.Hostname != "" {
		hostname = payload.Hostname
	}
	var flavor *model.Flavor
	if payload.Flavor != "" {
		flavor, err = flavorAdmin.GetFlavorByName(ctx, payload.Flavor)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid flavor query", err)
			return
		}
	}
	err = instanceAdmin.Update(ctx, instance, flavor, hostname, payload.PowerAction, int(instance.Hyper))
	if err != nil {
		log.Println("Patch instance failed, %v", err)
		return
	}
	instanceResp, err := v.getInstanceResponse(ctx, instance)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, instanceResp)
}

// @Summary delete a instance
// @Description delete a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   id  path  int  true  "Instance ID"
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [delete]
func (v *InstanceAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = instanceAdmin.Delete(ctx, instance)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a instance
// @Description create a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   InstancePayload  true   "Instance create payload"
// @Success 200 {array} InstanceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances [post]
func (v *InstanceAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &InstancePayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	hostname := payload.Hostname
	userdata := payload.Userdata
	image, err := imageAdmin.GetImage(ctx, payload.Image)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid image", err)
		return
	}
	flavor, err := flavorAdmin.GetFlavorByName(ctx, payload.Flavor)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid flavor", err)
		return
	}
	zone, err := zoneAdmin.GetZoneByName(ctx, payload.Zone)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid zone", err)
		return
	}
	var router *model.Router
	if payload.VPC != nil {
		router, err = routerAdmin.GetRouter(ctx, payload.VPC)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid VPC", nil)
			return
		}
	}
	router, primaryIface, err := v.getInterfaceInfo(ctx, router, payload.PrimaryInterface)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid primary interface", err)
		return
	}
	var secondaryIfaces []*routes.InterfaceInfo
	for _, ifacePayload := range payload.SecondaryInterfaces {
		var ifaceInfo *routes.InterfaceInfo
		_, ifaceInfo, err = v.getInterfaceInfo(ctx, router, ifacePayload)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid secondary interfaces", err)
			return
		}
		secondaryIfaces = append(secondaryIfaces, ifaceInfo)
	}
	count := 1
	if payload.Count > count {
		count = payload.Count
	}
	var keys []*model.Key
	for _, ky := range payload.Keys {
		var key *model.Key
		key, err = keyAdmin.GetKey(ctx, ky)
		keys = append(keys, key)
	}
	var routerID int64
	if router != nil {
		routerID = router.ID
	}
	instances, err := instanceAdmin.Create(ctx, count, hostname, userdata, image, flavor, zone, routerID, primaryIface, secondaryIfaces, keys, payload.Hypervisor)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to create instances", err)
		return
	}
	instancesResp := make([]*InstanceResponse, len(instances))
	for i, instance := range instances {
		instancesResp[i], err = v.getInstanceResponse(ctx, instance)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to create instances", err)
			return
		}
	}
	c.JSON(http.StatusOK, instancesResp)
}

func (v *InstanceAPI) getInterfaceInfo(ctx context.Context, vpc *model.Router, ifacePayload *InterfacePayload) (router *model.Router, ifaceInfo *routes.InterfaceInfo, err error) {
	if ifacePayload == nil || ifacePayload.Subnet == nil {
		err = fmt.Errorf("Interface with subnet must be provided")
		return
	}
	subnet, err := subnetAdmin.GetSubnet(ctx, ifacePayload.Subnet)
	if err != nil {
		return
	}
	router = vpc
	if router != nil && router.ID != subnet.RouterID {
		err = fmt.Errorf("VPC of subnet must be the same with VPC of instance")
		return
	}
	if router == nil && subnet.RouterID > 0 {
		router, err = routerAdmin.Get(ctx, subnet.RouterID)
		if err != nil {
			return
		}
	}
	ifaceInfo = &routes.InterfaceInfo{
		Subnet: subnet,
	}
	if ifacePayload.IpAddress != "" {
		ifaceInfo.IpAddress = ifacePayload.IpAddress
	}
	if ifacePayload.MacAddress != "" {
		ifaceInfo.MacAddress = ifacePayload.MacAddress
	}
	var secGroup *model.SecurityGroup
	if len(ifacePayload.SecurityGroups) == 0 && router != nil {
		secGroup, err = secgroupAdmin.Get(ctx, router.DefaultSG, router.ID)
		if err != nil {
			return
		}
		ifaceInfo.SecurityGroups = append(ifaceInfo.SecurityGroups, secGroup)
	} else {
		for _, sg := range ifacePayload.SecurityGroups {
			secGroup, err = secgroupAdmin.GetSecurityGroup(ctx, sg, subnet.RouterID)
			if err != nil {
				return
			}
			ifaceInfo.SecurityGroups = append(ifaceInfo.SecurityGroups, secGroup)
		}
	}
	return
}

func (v *InstanceAPI) getInstanceResponse(ctx context.Context, instance *model.Instance) (instanceResp *InstanceResponse, err error) {
	instanceResp = &InstanceResponse{
		ID:       instance.UUID,
		Hostname: instance.Hostname,
		Status:   instance.Status,
	}
	if instance.Image != nil {
		instanceResp.Image = &BaseReference{
			ID:   instance.Image.UUID,
			Name: instance.Image.Name,
		}
	}
	if instance.Flavor != nil {
		instanceResp.Flavor = instance.Flavor.Name
	}
	if instance.Zone != nil {
		instanceResp.Zone = instance.Zone.Name
	}
	keys := make([]*BaseReference, len(instance.Keys))
	for i, key := range instance.Keys {
		keys[i] = &BaseReference{
			ID:   key.UUID,
			Name: key.Name,
		}
	}
	hyper, hyperErr := hyperAdmin.GetHyperByHostid(ctx, instance.Hyper)
	if hyperErr == nil {
		instanceResp.Hypervisor = hyper.Hostname
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
		for _, sg := range iface.Secgroups {
			interfaces[i].SecurityGroups = append(interfaces[i].SecurityGroups, &BaseReference{
				ID:   sg.UUID,
				Name: sg.Name,
			})
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

// @Summary list instances
// @Description list instances
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} InstanceListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances [get]
func (v *InstanceAPI) List(c *gin.Context) {
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
			return
		}
	}
	instanceListResp.Instances = instanceList
	c.JSON(http.StatusOK, instanceListResp)
	return
}
