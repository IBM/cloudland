/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	cidrFun "github.com/apparentlymart/go-cidr/cidr"
	"github.com/go-macaron/session"
	"github.com/go-openapi/strfmt"
	uuidPk "github.com/google/uuid"
	macaron "gopkg.in/macaron.v1"
)

var (
	subnetInstance = &SubnetRest{}
)

type SubnetRest struct{}

func (v *SubnetRest) ListNetworks(c *macaron.Context) {
	offset := c.QueryInt64("marker")
	limit := c.QueryInt64("limit")
	reverse := c.QueryBool("page_reverse")
	order := "created_at"
	if reverse {
		order = "-created_at"
	}
	_, subnets, err := subnetAdmin.List(offset, limit, order)
	if err != nil {
		c.JSON(500, NewResponseError("List subnets fail", err.Error(), 500))
		return
	}

	networks := &restModels.ListNetworksOKBody{}
	networkItems := []*restModels.Network{}
	for _, subnet := range subnets {
		creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
		network := &restModels.Network{
			AdminStateUp:           true,
			CreatedAt:              creatAt,
			AvailabilityZones:      []string{"nova"},
			ID:                     subnet.UUID,
			Name:                   subnet.Name,
			Status:                 "Active",
			UpdatedAt:              updateAt,
			ProviderNetworkType:    "vxlan",
			ProviderSegmentationID: &subnet.Vlan,
			Subnets:                []string{},
		}
		if subnet.Netmask != "" {
			network.Subnets = append(network.Subnets, subnet.UUID)
		}
		networkItems = append(networkItems, network)
	}
	networks.Networks = networkItems
	c.JSON(200, networks)
}

func (v *SubnetRest) DeleteNetwork(c *macaron.Context, store session.Store) (err error) {
	uuid := c.Params("id")
	if uuid == "" {
		log.Println("empty network uuid")
		c.JSON(400, NewResponseError("empty network uuid", "uuid is empty", 400))
		return
	}
	subnetID := findIDbyUUID(&model.Subnet{}, uuid)
	if subnetID < 0 {
		c.Resp.WriteHeader(404)
		return
	}
	err = subnetAdmin.Delete(int64(subnetID))
	if err != nil {
		c.Resp.WriteHeader(412)
		return
	}
	c.Resp.WriteHeader(204)
	return
}

func (v *SubnetRest) CreateNetwork(c *macaron.Context) {
	db := DB()
	body, _ := c.Req.Body().Bytes()
	log.Println(string(body))
	if err := JsonSchemeCheck(`network.json`, body); err != nil {
		log.Println(string(body))
		c.JSON(err.Code, ResponseError{
			Error: *err,
		})
		return
	}
	uuid := uuidPk.New().String()
	requestData := &restModels.CreateNetworkParamsBody{}
	if err := json.Unmarshal(body, requestData); err != nil {
		c.JSON(500, NewResponseError("Unmarshal fail", err.Error(), 500))
		return
	}
	if result, err := checkIfExistVni(requestData.Network.ProviderSegmentationID); err != nil {
		c.JSON(500, NewResponseError("check vni fail", err.Error(), 500))
	} else if result {
		c.JSON(
			400,
			NewResponseError(
				"duplicate vni",
				fmt.Sprintf("the vni %d has been used", requestData.Network.ProviderSegmentationID),
				400,
			),
		)
		return
	}
	subnet := &model.Subnet{
		Model: model.Model{
			UUID: uuid,
		},
		Name: requestData.Network.Name,
		Vlan: requestData.Network.ProviderSegmentationID,
		Type: "internal",
	}
	err := db.Create(subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		c.JSON(500, NewResponseError("create network fail", err.Error(), 500))
		return
	}
	creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
	updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
	/*
		// simulate network response body
		{
		    "network": {
		        "status": "ACTIVE",
		        "subnets": [],
		        "availability_zone_hints": [],
		        "availability_zones": [
		            "nova"
		        ],
		        "created_at": "2016-03-08T20:19:41",
		        "name": "net1",
		        "admin_state_up": true,
		        "dns_domain": "",
		        "ipv4_address_scope": null,
		        "ipv6_address_scope": null,
		        "l2_adjacency": true,
		        "mtu": 1500,
		        "port_security_enabled": true,
		        "project_id": "9bacb3c5d39d41a79512987f338cf177",
		        "tags": ["tag1,tag2"],
		        "tenant_id": "9bacb3c5d39d41a79512987f338cf177",
		        "updated_at": "2016-03-08T20:19:41",
		        "qos_policy_id": "6a8454ade84346f59e8d40665f878b2e",
		        "revision_number": 1,
		        "segments": [
		            {
		                "provider:segmentation_id": 2,
		                "provider:physical_network": "public",
		                "provider:network_type": "vlan"
		            },
		            {
		                "provider:segmentation_id": null,
		                "provider:physical_network": "default",
		                "provider:network_type": "flat"
		            }
		        ],
		        "shared": false,
		        "id": "4e8e5957-649f-477b-9e5b-f1f75b21c03c",
		        "description": "",
		        "is_default": false
		    }
		}
	*/
	responseBody := &restModels.CreateNetworkOKBody{
		Network: &restModels.Network{
			AdminStateUp:           true,
			AvailabilityZones:      []string{"nova"},
			CreatedAt:              creatAt,
			ID:                     subnet.UUID,
			IsDefault:              false,
			Mtu:                    1500,
			Name:                   subnet.Name,
			PortSecurityEnabled:    false,
			ProviderNetworkType:    "vxlan",
			ProviderSegmentationID: &subnet.Vlan,
			QosPolicyID:            subnet.UUID,
			RouterExternal:         false,
			Shared:                 false,
			Status:                 "ACTIVCE",
			Subnets:                []string{},
			UpdatedAt:              updateAt,
			VlanTransparent:        false,
		},
	}
	c.JSON(200, responseBody)
}

func (v *SubnetRest) ListSubnet(c *macaron.Context) {
	offset := c.QueryInt64("marker")
	limit := c.QueryInt64("limit")
	reverse := c.QueryBool("page_reverse")
	order := "created_at"
	if reverse {
		order = "-created_at"
	}
	_, subnets, err := subnetAdmin.List(offset, limit, order)
	if err != nil {
		c.JSON(500, NewResponseError("List subnets fail", err.Error(), 500))
		return
	}
	subnetItems := restModels.Subnets{}
	for _, subnet := range subnets {
		creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
		subnetItem := &restModels.Subnet{
			Cidr:           subnet.Network,
			CreatedAt:      creatAt,
			EnableDhcp:     true,
			GatewayIP:      strfmt.IPv4(subnet.Gateway),
			ID:             subnet.UUID,
			IPVersion:      4,
			Name:           subnet.Name,
			NetworkID:      subnet.UUID,
			ProjectID:      "default",
			RevisionNumber: 0,
			TenantID:       "default",
			UpdatedAt:      updateAt,
		}
		subnetItems = append(subnetItems, subnetItem)
	}
	subnetsResponse := &restModels.ListSubnetsOKBody{
		Subnets: subnetItems,
	}
	c.JSON(200, subnetsResponse)
}

//CreateSubnet : create subnet in db with network id
// netweork ID  and subnet ID is same
//if network ID is empty , create subnet with new network ID
//if network ID has been created and subnet address is empty , update subnet ID with new subnet infor
// if network ID has been created and subnet address is not empyt, the subnet and network has been create. return duplicate error msg
func (v *SubnetRest) CreateSubnet(c *macaron.Context) {
	db := DB()
	var cidr *net.IPNet
	var err error
	body, _ := c.Req.Body().Bytes()
	log.Println(string(body))
	if err := JsonSchemeCheck(`subnet.json`, body); err != nil {
		log.Println(string(body))
		c.JSON(err.Code, ResponseError{
			Error: *err,
		})
		return
	}
	requestData := &restModels.CreateSubnetParamsBody{}
	if err := json.Unmarshal(body, requestData); err != nil {
		c.JSON(500, NewResponseError("Unmarshal fail", err.Error(), 500))
		return
	}
	if _, cidr, err = net.ParseCIDR(requestData.Subnet.Cidr); err != nil {
		c.JSON(500, NewResponseError("parse Cidr fail", err.Error(), 500))
		return
	}
	name := requestData.Subnet.Name
	networkUUID := requestData.Subnet.NetworkID
	network, end := cidrFun.AddressRange(cidr)
	first := cidrFun.Inc(network)
	gateway := cidrFun.Dec(end)
	last := cidrFun.Dec(gateway)
	netmask := net.IP(cidr.Mask).String()
	netmaskSize, _ := cidr.Mask.Size()

	var subnet *model.Subnet
	var count int
	existingSubnet := &model.Subnet{}
	if networkUUID == "" {
		// if netID is empty, create subnet with random network ID
		networkUUID = uuidPk.New().String()
		subnet, err = subnetAdmin.Create(name, "", network.String(), netmask, gateway.String(), first.String(), last.String(), "internal", networkUUID)
		if err != nil {
			log.Println(fmt.Sprintf("Failed to create subnet with error: %v", err))
			c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
			return
		}
		log.Println("success to create subnet: %s", networkUUID)
	} else {
		if err := db.Model(&model.Subnet{}).Where("uuid = ?", networkUUID).Count(&count).Error; err != nil {
			log.Println(fmt.Sprintf("Failed to query existing vlan, %v", err))
			c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
			return
		}
		if count == 0 {
			// if network don't created, create subnet directly with specified network ID
			subnet, err = subnetAdmin.Create(name, "", network.String(), netmask, gateway.String(), first.String(), last.String(), "internal", networkUUID)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to create subnet with error: %v", err))
				c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
				return
			}
			log.Println(fmt.Sprintf("success to create subnet: %s", networkUUID))
		} else {
			db.Where("uuid = ?", networkUUID).First(existingSubnet)
			if existingSubnet.Network != "" {
				log.Println(fmt.Sprintf("duplicated subnet: %s", existingSubnet.UUID))
				c.JSON(
					500,
					NewResponseError(
						"create subnet fail",
						fmt.Sprintf("duplicate subnet: %s", existingSubnet.UUID),
						500,
					),
				)
				return
			}
		}
	}

	//if network has been created with empty subnet , update subnet
	if subnet == nil && existingSubnet.Network == "" {
		//update subnet
		gatewayStr := fmt.Sprintf("%s/%d", gateway, netmaskSize)
		existingSubnet.Start = first.String()
		existingSubnet.End = last.String()
		existingSubnet.Gateway = gatewayStr
		existingSubnet.Netmask = netmask
		existingSubnet.Network = network.String()
		db.Save(existingSubnet)
		log.Println(fmt.Sprintf("success update subnet: %s", networkUUID))
		subnet = existingSubnet
	}
	creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
	updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
	subnetRespons := &restModels.CreateSubnetOKBody{
		Subnet: &restModels.Subnet{
			Cidr:           subnet.Network,
			CreatedAt:      creatAt,
			EnableDhcp:     true,
			GatewayIP:      strfmt.IPv4(subnet.Gateway),
			ID:             subnet.UUID,
			IPVersion:      4,
			Name:           subnet.Name,
			NetworkID:      subnet.UUID,
			ProjectID:      "default",
			RevisionNumber: 0,
			TenantID:       "default",
			UpdatedAt:      updateAt,
		},
	}
	c.JSON(200, subnetRespons)
	return
}

func (v *SubnetRest) DeleteSubnet(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	subnetID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = subnetAdmin.Delete(int64(subnetID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "subnets",
	})
	return
}

// findIDbyUUID find ID (int )by UUID (string)
// if can't find ID in database, return number less than zero
func findIDbyUUID(obj interface{}, uuid string) (id int) {

	return 0
}
