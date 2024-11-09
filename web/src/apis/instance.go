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
	"web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var instanceAPI = &InstanceAPI{}
var instanceAdmin = &routes.InstanceAdmin{}

type InstanceAPI struct{}

type InstancePatchPayload struct {
	Hostname    string             `json:"hostname" binding:"omitempty,hostname|fqdn"`
	PowerAction common.PowerAction `json:"power_action" binding:"omitempty,oneof=stop hard_stop start restart hard_restart pause unpause)"`
	Flavor      string             `json:"flavor" binding:"required,min=1,max=32"`
}

type InstancePayload struct {
	Count               int                     `json:"count" binding:"omitempty,gte=1,lte=16"`
	Hyper               int                     `json:"hyper" binding:"omitempty,gte=0"`
	Hostname            string                  `json:"hostname" binding:"required,hostname|fqdn"`
	Keys                []*common.BaseReference `json:"keys" binding:"required,gte=1,lte=16"`
	Flavor              string                  `json:"flavor" binding:"required,min=1,max=32"`
	Image               *common.BaseReference   `json:"image" binding:"required"`
	PrimaryInterface    *InterfacePayload       `json:"primary_interface", binding:"required"`
	SecondaryInterfaces []*InterfacePayload     `json:"secondary_interfaces" binding:"omitempty"`
	Zone                string                  `json:"zone" binding:"required,min=1,max=32"`
	VPC                 *common.BaseReference   `json:"vpc" binding:"omitempty"`
	Userdata            string                  `json:"userdata,omitempty"`
}

type InstanceResponse struct {
	ID         string                  `json:"id"`
	Hostname   string                  `json:"hostname"`
	Status     string                  `json:"status"`
	Interfaces []*InterfaceResponse    `json:"interfaces"`
	Flavor     string                  `json:"flavor"`
	Image      *common.BaseReference   `json:"image"`
	Keys       []*common.BaseReference `json:"keys"`
	Zone       string                  `json:"zone"`
	VPC        *common.BaseReference   `json:"vpc,omitempty"`
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
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [get]
func (v *InstanceAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
	}
	instanceResp, err := v.getInstanceResponse(ctx, instance)
	if err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
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
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [patch]
func (v *InstanceAPI) Patch(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
	}
	payload := &InstancePatchPayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
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
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid flavor query", err)
		}
	}
	err = instanceAdmin.Update(ctx, instance, flavor, hostname, payload.PowerAction, int(instance.Hyper))
	if err != nil {
		log.Println("Patch instance failed, %v", err)
		return
	}
	instanceResp, err := v.getInstanceResponse(ctx, instance)
	if err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
	}
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
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [delete]
func (v *InstanceAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
	}
	err = instanceAdmin.Delete(ctx, instance)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
	}
	c.JSON(http.StatusNoContent, nil)
}

//
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
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	hostname := payload.Hostname
	userdata := payload.Userdata
	image, err := imageAdmin.GetImage(ctx, payload.Image)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid image", err)
		return
	}
	flavor, err := flavorAdmin.GetFlavorByName(ctx, payload.Flavor)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid flavor", err)
		return
	}
	zone, err := zoneAdmin.GetZoneByName(ctx, payload.Zone)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid zone", err)
		return
	}
	primaryIface, err := v.getInterfaceInfo(ctx, payload.PrimaryInterface)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid primary interface", err)
		return
	}
	routerID := primaryIface.Subnet.RouterID
	if payload.VPC != nil {
		var router *model.Router
		router, err = routerAdmin.GetRouter(ctx, payload.VPC)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid VPC", nil)
			return
		}
		routerID = router.ID
	}
	var secondaryIfaces []*routes.InterfaceInfo
	for _, ifacePayload := range payload.SecondaryInterfaces {
		var ifaceInfo *routes.InterfaceInfo
		ifaceInfo, err = v.getInterfaceInfo(ctx, ifacePayload)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid secondary interfaces", err)
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
	instances, err := instanceAdmin.Create(ctx, count, hostname, userdata, image, flavor, zone, routerID, primaryIface, secondaryIfaces, keys, -1)
	if err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "Failed to create instances", err)
		return
	}
	instancesResp := make([]*InstanceResponse, len(instances))
	for i, instance := range instances {
		instancesResp[i], err = v.getInstanceResponse(ctx, instance)
		if err != nil {
			common.ErrorResponse(c, http.StatusInternalServerError, "Failed to create instances", err)
			return
		}
	}
	c.JSON(http.StatusOK, instancesResp)
}

func (v *InstanceAPI) getInterfaceInfo(ctx context.Context, ifacePayload *InterfacePayload) (ifaceInfo *routes.InterfaceInfo, err error) {
	if ifacePayload == nil || ifacePayload.Subnet == nil {
		err = fmt.Errorf("Interface with subnet must be provided")
		return
	}
	subnet, err := subnetAdmin.GetSubnet(ctx, ifacePayload.Subnet)
	if err != nil {
		return
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
	for _, sg := range ifacePayload.SecurityGroups {
		var secGroup *model.SecurityGroup
		secGroup, err = secgroupAdmin.GetSecurityGroup(ctx, sg, subnet.RouterID)
		if err != nil {
			return
		}
		ifaceInfo.SecurityGroups = append(ifaceInfo.SecurityGroups, secGroup)
	}
	return
}

func (v *InstanceAPI) getInstanceResponse(ctx context.Context, instance *model.Instance) (instanceResp *InstanceResponse, err error) {
	instanceResp = &InstanceResponse{
		ID:       instance.UUID,
		Hostname: instance.Hostname,
		Status:   instance.Status,
		Flavor:   instance.Flavor.Name,
		Image: &common.BaseReference{
			ID:   instance.Image.UUID,
			Name: instance.Image.Name,
		},
		Zone: instance.Zone.Name,
	}
	keys := make([]*common.BaseReference, len(instance.Keys))
	for i, key := range instance.Keys {
		keys[i] = &common.BaseReference{
			ID:   key.UUID,
			Name: key.Name,
		}
	}
	instanceResp.Keys = keys
	interfaces := make([]*InterfaceResponse, len(instance.Interfaces))
	for i, iface := range instance.Interfaces {
		interfaces[i] = &InterfaceResponse{
			BaseReference: &common.BaseReference{
				ID:   iface.UUID,
				Name: iface.Name,
			},
			MacAddress: iface.MacAddr,
			IPAddress:  iface.Address.Address,
			IsPrimary:  iface.PrimaryIf,
			Subnet: &common.BaseReference{
				ID:   iface.Address.Subnet.UUID,
				Name: iface.Address.Subnet.Name,
			},
		}
		if iface.PrimaryIf && len(instance.FloatingIps) > 0 {
			floatingIps := make([]*FloatingIpInfo, len(instance.FloatingIps))
			for i, floatingip := range instance.FloatingIps {
				floatingIps[i] = &FloatingIpInfo{
					BaseReference: &common.BaseReference{
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
		instanceResp.VPC = &common.BaseReference{
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
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances [get]
func (v *InstanceAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	memberShip := routes.GetMemberShip(ctx)
	log.Printf("Membership: %v\n", memberShip)
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
	total, instances, err := instanceAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Failed to list instances", err)
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
			common.ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		}
	}
	instanceListResp.Instances = instanceList
	c.JSON(http.StatusOK, instanceListResp)
	return
}
