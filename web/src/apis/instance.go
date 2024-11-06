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
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type BasePayloadRef struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
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
	Hostname    *string       `json:"hostname,omitempty"`
	PowerAction *PowerAction `json:"power_action,omitempty"`
}

type InstancePayload struct {
	Count      *int                 `json:"count,omitempty"`
	Hyper      *int                 `json:"hyper,omitempty"`
	Hostname   *string              `json:"hostname,required"`
	Keys       []*BasePayloadRef    `json:"keys,required"`
	Flavor     *string      `json:"flavor,required"`
	Image      *BasePayloadRef      `json:"image,required"`
	PrimaryInterface *InterfacePayload `json:"primary_interface,required"`
	SecondaryInterfaces []*InterfacePayload `json:"secondary_interfaces,omitempty"`
	Zone       *string              `json:"zone,required"`
	vpc       *BasePayloadRef              `json:"zone,omitempty"`
	Userdata   *string              `json:"userdata,omitempty"`
}

type InstanceResponse struct {
	ID         string               `json:"id"`
	Hostname   string               `json:"hostname"`
	Status     string               `json:"status"`
	Interfaces []*InterfaceResponse `json:"interfaces"`
	Flavor     string       `json:"flavor"`
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
// @Success 200 {array} InstanceResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /instances [post]
func (v *InstanceAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &InstancePayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Input JSON format error", err)
		return
	}
	count := 1
	if payload.Count != nil {
		count = *payload.Count
	}
	if count <= 0 || count > InstanceQuota {
		errMsg := fmt.Sprintf("Provided count must be 1-%d", InstanceQuota)
		ErrorResponse(c, http.StatusBadRequest, errMsg, err)
		return
	}
	if payload.Hostname == nil {
		ErrorResponse(c, http.StatusBadRequest, "Hostname must be provided", nil)
		return
	}
	hostname := *payload.Hostname
	userdata := ""
	if payload.Userdata != nil {
		userdata = *payload.Userdata
	}
	if payload.Image == nil || (payload.Image.ID == nil && payload.Image.Name == nil) {
		ErrorResponse(c, http.StatusBadRequest, "Image must be provided", nil)
		return
	}
	image, err := imageAdmin.GetImage(ctx, payload.Image.ID, payload.Image.Name)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid image", nil)
		return
	}
	if payload.Flavor == nil {
		ErrorResponse(c, http.StatusBadRequest, "Flavor must be provided", nil)
		return
	}
	flavor, err := flavorAdmin.GetFlavorByName(ctx, *payload.Flavor)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid flavor", nil)
		return
	}
	if payload.Zone == nil {
		ErrorResponse(c, http.StatusBadRequest, "Zone must be provided", nil)
		return
	}
	zone, err := zoneAdmin.GetZoneByName(ctx, *payload.Zone)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid zone", nil)
		return
	}
	primaryIface, err := v.getInterfaceInfo(ctx, payload.PrimaryInterface)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid primary interface", nil)
		return
	}
	routerID := primaryIface.Subnet.RouterID
	/*
	if payload.VPC != nil {
		router := routerAdmin.GetRouter(ctx, payload.VPC)
	}
	*/
	var secondaryIfaces []*routes.InterfaceInfo
	for _, ifacePayload := range payload.SecondaryInterfaces {
		var ifaceInfo *routes.InterfaceInfo
		ifaceInfo, err = v.getInterfaceInfo(ctx, ifacePayload)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid secondary interfaces", nil)
			return
		}
		secondaryIfaces = append(secondaryIfaces, ifaceInfo)
	}
	instances, err := instanceAdmin.Create(ctx, count, hostname, userdata, image, flavor, zone, routerID, primaryIface, secondaryIfaces, nil, 0)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create instances", nil)
		return
	}
	c.JSON(http.StatusOK, instances)
}

func (v *InstanceAPI) getInterfaceInfo(ctx context.Context, ifacePayload *InterfacePayload) (ifaceInfo *routes.InterfaceInfo, err error) {
	if ifacePayload == nil || ifacePayload.Subnet == nil {
		err = fmt.Errorf("Interface with subnet must be provided")
		return
	}
	/*
	subnet, err := subnetAdmin.GetSubnet(ctx, subnetRef)
	if err != nil {
		return
	}
	*/
	ifaceInfo = &routes.InterfaceInfo{
	//	Subnet: subnet,
	}
	if ifacePayload.IpAddress != nil {
		ifaceInfo.IpAddress = *ifacePayload.IpAddress
	}
	if ifacePayload.MacAddress != nil {
		ifaceInfo.MacAddress = *ifacePayload.MacAddress
	}
	/*
	for _, sg := range primaryPayload.SecurityGroups {
		secGroup, sgErr := secgroupAdmin.GetSecurityGroup(ctx, sg.ID, sg.Name, primarySubnet.RouterID)
		if sgErr != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid security group", nil)
			return
		}
		ifaceInfo.SecurityGroups = append(ifaceInfo.SecurityGroups, secGroup)
	}
	*/
	return
}

func (v *InstanceAPI) getInstanceResponse(ctx context.Context, instance *model.Instance) (instanceResp *InstanceResponse, err error) {
	instanceResp = &InstanceResponse{
		ID:       instance.UUID,
		Hostname: instance.Hostname,
		Status:   instance.Status,
		Flavor: instance.Flavor.Name,
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
