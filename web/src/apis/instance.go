/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

type InstanceSetUserPasswordPayload struct {
	Password string `json:"password" binding:"required,min=8,max=64"`
	UserName string `json:"user_name" binding:"required,min=2,max=32"`
}

type InstancePayload struct {
	Count               int                 `json:"count" binding:"omitempty,gte=1,lte=16"`
	Hypervisor          *int                `json:"hypervisor" binding:"omitempty,gte=0,lte=65535"`
	Hostname            string              `json:"hostname" binding:"required,hostname|fqdn"`
	Keys                []*BaseReference    `json:"keys" binding:"omitempty,gte=0,lte=16"`
	RootPasswd          string              `json:"root_passwd" binding:"omitempty,min=8,max=32"`
	Flavor              string              `json:"flavor" binding:"required,min=1,max=32"`
	Image               *BaseReference      `json:"image" binding:"required"`
	PrimaryInterface    *InterfacePayload   `json:"primary_interface", binding:"required"`
	SecondaryInterfaces []*InterfacePayload `json:"secondary_interfaces" binding:"omitempty"`
	Zone                string              `json:"zone" binding:"required,min=1,max=32"`
	VPC                 *BaseReference      `json:"vpc" binding:"omitempty"`
	Userdata            string              `json:"userdata,omitempty"`
}

type InstanceResponse struct {
	*ResourceReference
	Hostname    string               `json:"hostname"`
	Status      string               `json:"status"`
	Interfaces  []*InterfaceResponse `json:"interfaces"`
	Volumes     []*ResourceReference `json:"volumes"`
	Flavor      string               `json:"flavor"`
	Image       *ResourceReference   `json:"image"`
	Keys        []*ResourceReference `json:"keys"`
	PasswdLogin bool                 `json:"passwd_login"`
	Zone        string               `json:"zone"`
	VPC         *ResourceReference   `json:"vpc,omitempty"`
	Hypervisor  string               `json:"hypervisor,omitempty"`
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
// @Param   id  path  string  true  "Instance UUID"
// @Success 200 {object} InstanceResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id} [get]
func (v *InstanceAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Get instance %s", uuID)
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
	logger.Debugf("Patch instance %s", uuID)
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get instance %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	payload := &InstancePatchPayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind JSON, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	hostname := instance.Hostname
	if payload.Hostname != "" {
		hostname = payload.Hostname
		logger.Debugf("Update hostname to %s", hostname)
	}
	var flavor *model.Flavor
	if payload.Flavor != "" {
		flavor, err = flavorAdmin.GetFlavorByName(ctx, payload.Flavor)
		if err != nil {
			logger.Errorf("Failed to get flavor %+v, %+v", payload.Flavor, err)
			ErrorResponse(c, http.StatusBadRequest, "Invalid flavor query", err)
			return
		}
	}
	err = instanceAdmin.Update(ctx, instance, flavor, hostname, payload.PowerAction, int(instance.Hyper))
	if err != nil {
		logger.Errorf("Patch instance failed, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Patch instance failed", err)
		return
	}
	instanceResp, err := v.getInstanceResponse(ctx, instance)
	if err != nil {
		logger.Errorf("Failed to create instance response, %+v", err)
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Patch instance %s success, response: %+v", uuID, instanceResp)
	c.JSON(http.StatusOK, instanceResp)
}

// @Summary set user password for a instance
// @Description set user password for a instance
// @tags Compute
// @Accept  json
// @Produce json
// @Param   id  path  string  true  "Instance UUID"
// @Param   message	body   InstanceSetUserPasswordPayload  true   "Instance set user password payload"
// @Success 200
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /instances/{id}/set_user_password [post]
func (v *InstanceAPI) SetUserPassword(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Set user password for instance %s", uuID)
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get instance %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid instance query", err)
		return
	}
	payload := &InstanceSetUserPasswordPayload{}
	err = c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind JSON, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	err = instanceAdmin.SetUserPassword(ctx, instance.ID, payload.UserName, payload.Password)
	if err != nil {
		logger.Errorf("Set user password failed, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Set user password failed", err)
		return
	}
	c.JSON(http.StatusOK, nil)
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
	logger.Debugf("Delete instance %s", uuID)
	instance, err := instanceAdmin.GetInstanceByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get instance %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = instanceAdmin.Delete(ctx, instance)
	if err != nil {
		logger.Errorf("Failed to delete instance %s, %+v", uuID, err)
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
	logger.Debug("Create instance")
	ctx := c.Request.Context()
	payload := &InstancePayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Failed to bind instance payload JSON, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	logger.Debugf("Creating instance with payload: %+v", payload)
	hostname := payload.Hostname
	rootPasswd := payload.RootPasswd
	userdata := payload.Userdata
	image, err := imageAdmin.GetImage(ctx, payload.Image)
	if err != nil {
		logger.Errorf("Failed to get image %+v, %+v", payload.Image, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid image", err)
		return
	}
	flavor, err := flavorAdmin.GetFlavorByName(ctx, payload.Flavor)
	if err != nil {
		logger.Errorf("Failed to get flavor %+v, %+v", payload.Flavor, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid flavor", err)
		return
	}
	zone, err := zoneAdmin.GetZoneByName(ctx, payload.Zone)
	if err != nil {
		logger.Errorf("Failed to get zone %+v, %+v", payload.Zone, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid zone", err)
		return
	}
	var router *model.Router
	if payload.VPC != nil {
		router, err = routerAdmin.GetRouter(ctx, payload.VPC)
		if err != nil {
			logger.Errorf("Failed to get VPC %+v, %+v", payload.VPC, err)
			ErrorResponse(c, http.StatusBadRequest, "Invalid VPC", nil)
			return
		}
	}
	router, primaryIface, err := v.getInterfaceInfo(ctx, router, payload.PrimaryInterface)
	if err != nil {
		logger.Errorf("Failed to get primary interface %+v, %+v", payload.PrimaryInterface, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid primary interface", err)
		return
	}
	var secondaryIfaces []*routes.InterfaceInfo
	for _, ifacePayload := range payload.SecondaryInterfaces {
		var ifaceInfo *routes.InterfaceInfo
		_, ifaceInfo, err = v.getInterfaceInfo(ctx, router, ifacePayload)
		if err != nil {
			logger.Errorf("Failed to get secondary interface %+v, %+v", ifacePayload, err)
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
	hypervisor := -1
	if payload.Hypervisor != nil {
		hypervisor = *payload.Hypervisor
	}
	logger.Debugf("Creating %d instances with hostname %s, userdata %s, image %s, flavor %s, zone %s, router %d, primaryIface %v, secondaryIfaces %v, keys %v, hypervisor %d",
		count, hostname, userdata, image.Name, flavor.Name, zone.Name, routerID, primaryIface, secondaryIfaces, keys, hypervisor)
	instances, err := instanceAdmin.Create(ctx, count, hostname, userdata, image, flavor, zone, routerID, primaryIface, secondaryIfaces, keys, rootPasswd, hypervisor)
	if err != nil {
		logger.Errorf("Failed to create instances, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to create instances", err)
		return
	}
	logger.Debugf("Created %d instances, %+v", len(instances), instances)
	instancesResp := make([]*InstanceResponse, len(instances))
	for i, instance := range instances {
		instancesResp[i], err = v.getInstanceResponse(ctx, instance)
		if err != nil {
			logger.Errorf("Failed to create instance response, %+v", err)
			ErrorResponse(c, http.StatusInternalServerError, "Failed to create instances", err)
			return
		}
	}
	logger.Debugf("Create instance success, %+v", instancesResp)
	c.JSON(http.StatusOK, instancesResp)
}

func (v *InstanceAPI) getInterfaceInfo(ctx context.Context, vpc *model.Router, ifacePayload *InterfacePayload) (router *model.Router, ifaceInfo *routes.InterfaceInfo, err error) {
	logger.Debugf("Get interface info with VPC %+v, ifacePayload %+v", vpc, ifacePayload)
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
		AllowSpoofing: ifacePayload.AllowSpoofing,
	}
	if ifacePayload.IpAddress != "" {
		ifaceInfo.IpAddress = ifacePayload.IpAddress
	}
	if ifacePayload.MacAddress != "" {
		ifaceInfo.MacAddress = ifacePayload.MacAddress
	}
	if ifacePayload.Inbound > 0 {
		ifaceInfo.Inbound = ifacePayload.Inbound
	}
	if ifacePayload.Outbound > 0 {
		ifaceInfo.Outbound = ifacePayload.Outbound
	}
	if len(ifacePayload.SecurityGroups) == 0 {
		var routerID, sgID int64
		if router != nil {
			routerID = router.ID
			sgID = router.DefaultSG
		}
		var secgroup *model.SecurityGroup
		secgroup, err = secgroupAdmin.Get(ctx, sgID)
		if err != nil {
			return
		}
		if secgroup.RouterID != routerID {
			err = fmt.Errorf("Security group not in subnet vpc")
			return
		}
		ifaceInfo.SecurityGroups = append(ifaceInfo.SecurityGroups, secgroup)
	} else {
		for _, sg := range ifacePayload.SecurityGroups {
			var secgroup *model.SecurityGroup
			secgroup, err = secgroupAdmin.GetSecurityGroup(ctx, sg)
			if err != nil {
				return
			}
			if secgroup.RouterID != subnet.RouterID {
				err = fmt.Errorf("Security group not in subnet vpc")
				return
			}
			ifaceInfo.SecurityGroups = append(ifaceInfo.SecurityGroups, secgroup)
		}
	}
	logger.Debugf("Get interface info success, router %+v, ifaceInfo %+v", router, ifaceInfo)
	return
}

func (v *InstanceAPI) getInstanceResponse(ctx context.Context, instance *model.Instance) (instanceResp *InstanceResponse, err error) {
	logger.Debugf("Create instance response for instance %+v", instance)
	owner := orgAdmin.GetOrgName(instance.Owner)
	instanceResp = &InstanceResponse{
		ResourceReference: &ResourceReference{
			ID:        instance.UUID,
			Owner:     owner,
			CreatedAt: instance.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: instance.UpdatedAt.Format(TimeStringForMat),
		},
		Hostname: instance.Hostname,
		Status:   instance.Status,
	}
	if instance.Image != nil {
		instanceResp.Image = &ResourceReference{
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
	keys := make([]*ResourceReference, len(instance.Keys))
	for i, key := range instance.Keys {
		keys[i] = &ResourceReference{
			ID:   key.UUID,
			Name: key.Name,
		}
	}
	instanceResp.Keys = keys
	volumes := make([]*ResourceReference, len(instance.Volumes))
	for i, volume := range instance.Volumes {
		volumes[i] = &ResourceReference{
			ID:   volume.UUID,
			Name: volume.Name,
		}
	}
	instanceResp.Volumes = volumes
	hyper, hyperErr := hyperAdmin.GetHyperByHostid(ctx, instance.Hyper)
	if hyperErr == nil {
		instanceResp.Hypervisor = hyper.Hostname
	}
	interfaces := make([]*InterfaceResponse, len(instance.Interfaces))
	for i, iface := range instance.Interfaces {
		interfaces[i], err = interfaceAPI.getInterfaceResponse(ctx, instance, iface)
	}
	instanceResp.Interfaces = interfaces
	if instance.RouterID > 0 && instance.Router != nil {
		router := instance.Router
		instanceResp.VPC = &ResourceReference{
			ID:   router.UUID,
			Name: router.Name,
		}
	}
	logger.Debugf("Create instance response success, %+v", instanceResp)
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
	queryStr := c.DefaultQuery("query", "")
	vpcID := strings.TrimSpace(c.DefaultQuery("vpc_id", "")) // Retrieve vpc_id from query params
	logger.Debugf("List instances with offset %s, limit %s, query %s, vpc_id %s", offsetStr, limitStr, queryStr, vpcID)

	if vpcID != "" {
		logger.Debugf("Filtering instances by VPC ID: %s", vpcID)
		var router *model.Router
		router, err := routerAdmin.GetRouterByUUID(ctx, vpcID)
		if err != nil {
			logger.Errorf("Invalid query vpc_id: %s, %+v", vpcID, err)
			ErrorResponse(c, http.StatusBadRequest, "Invalid query router by vpc_id UUID: "+vpcID, err)
			return
		}

		logger.Debugf("The router with vpc_id: %+v\n", router)
		logger.Debugf("The router_id in vpc is: %d", router.ID)
		queryStr = fmt.Sprintf("router_id = %d", router.ID)
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		logger.Errorf("Invalid query offset: %s, %+v", offsetStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logger.Errorf("Invalid query limit: %s, %+v", limitStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		logger.Errorf("Invalid query offset or limit, %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, instances, err := instanceAdmin.List(ctx, int64(offset), int64(limit), "-created_at", queryStr)
	if err != nil {
		logger.Errorf("Failed to list instances, %+v", err)
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
			logger.Errorf("Failed to create instance response, %+v", err)
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	instanceListResp.Instances = instanceList
	logger.Debugf("List instances success, %+v", instanceListResp)
	c.JSON(http.StatusOK, instanceListResp)
	return
}
