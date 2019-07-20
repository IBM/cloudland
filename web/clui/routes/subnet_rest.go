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
	"strings"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	cidrFun "github.com/apparentlymart/go-cidr/cidr"
	"github.com/go-openapi/strfmt"
	uuidPk "github.com/google/uuid"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	subnetInstance = &SubnetRest{}
)

type SubnetRest struct{}

const MAXVLAN int64 = 4096
const MINVLAN int64 = 1

type networkType string

const INTERNAL networkType = `internal`
const PUBLICE networkType = `public`

var totalSubnets int64 = 50

func (v networkType) String() string {
	return string(v)
}

func (v networkType) GetBool() bool {
	if v == INTERNAL {
		return false
	}
	return true
}

func (v *SubnetRest) ListNetworks(c *macaron.Context) {
	// offset := c.QueryInt64("marker")
	// limit := c.QueryInt64("limit")
	reverse := c.QueryBool("page_reverse")
	order := "uuid,created_at"
	if reverse {
		order = "uuid,-created_at"
	}
	var subnets []*model.Subnet
	var err error
	totalSubnets, subnets, err = subnetAdmin.List(0, totalSubnets, order)
	if err != nil {
		c.JSON(500, NewResponseError("List subnets fail", err.Error(), 500))
		return
	}

	networks := &restModels.ListNetworksOKBody{}
	networkItems := []*restModels.Network{}
	var networkUUID string
	var index int
	for _, subnet := range subnets {
		creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
		if subnet.UUID == networkUUID {
			networkItems[index-1].Subnets = append(
				networkItems[index-1].Subnets,
				strconv.FormatInt(subnet.ID, 10),
			)
		} else {
			index++
			networkUUID = subnet.UUID
			networkType := restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVxlan
			if subnet.Vlan < 4096 {
				networkType = restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVlan
			}
			network := &restModels.Network{
				AdminStateUp:           true,
				CreatedAt:              creatAt,
				AvailabilityZones:      []string{"cloudland"},
				ID:                     subnet.UUID,
				Name:                   subnet.Name,
				Status:                 "Active",
				UpdatedAt:              updateAt,
				ProviderNetworkType:    networkType,
				ProviderSegmentationID: strconv.FormatInt(subnet.Vlan, 10),
				Subnets:                []string{},
			}
			if subnet.Netmask != "" {
				network.Subnets = append(network.Subnets, strconv.FormatInt(subnet.ID, 10))
			}
			networkItems = append(networkItems, network)
		}
	}
	networks.Networks = networkItems
	c.JSON(200, networks)
}

func (v *SubnetRest) DeleteNetwork(c *macaron.Context) {
	uuid := c.Params("id")
	if uuid == "" {
		log.Println("empty network uuid")
		c.JSON(400, NewResponseError("empty network uuid", "uuid is empty", 400))
		return
	}
	if err := subnetAdmin.DeleteByUUID(uuid); err != nil {
		c.JSON(500, NewResponseError("Delete network fail", err.Error(), 500))
		return
	}
	c.Status(204)
	return
}

func (v *SubnetRest) CreateNetwork(c *macaron.Context) {
	db := DB()
	body, _ := c.Req.Body().Bytes()
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
	vlanID, err := formateStringToInt64(c, requestData.Network.ProviderSegmentationID)
	if err != nil {
		return
	}
	if requestData.Network.ProviderNetworkType == "" {
		requestData.Network.ProviderNetworkType = restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVxlan
	}
	if requestData.Network.ProviderNetworkType == restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVxlan &&
		vlanID == 0 {
		vlanNo, err := getValidVni()
		if err != nil {
			c.JSON(500, NewResponseError("create vni fail", err.Error(), 500))
			return
		}
		vlanID = int64(vlanNo)
	}
	if requestData.Network.ProviderNetworkType == restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVlan {
		if requestData.Network.ProviderSegmentationID == "" {
			c.JSON(500, NewResponseError("must provide vlan ID", "empty vlan id", 500))
		}
		if vlanID > MAXVLAN || vlanID < MINVLAN {
			c.JSON(500, NewResponseError("invalid vlan range", "vlan range must be in 1-4096", 500))
		}
	}
	networkType := INTERNAL
	if requestData.Network.RouterExternal {
		networkType = PUBLICE
	}
	//check vni and vlan network whether has been created priviously
	if result, err := checkIfExistVni(vlanID); err != nil {
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
		Vlan: vlanID,
		Type: networkType.String(),
	}
	err = db.Create(subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		c.JSON(500, NewResponseError("create network fail", err.Error(), 500))
		return
	}
	creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
	updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))

	responseBody := &restModels.CreateNetworkOKBody{
		Network: &restModels.Network{
			AdminStateUp:           true,
			AvailabilityZones:      []string{"cloudland"},
			CreatedAt:              creatAt,
			ID:                     subnet.UUID,
			IsDefault:              false,
			Mtu:                    1500,
			Name:                   subnet.Name,
			PortSecurityEnabled:    false,
			ProviderNetworkType:    requestData.Network.ProviderNetworkType,
			ProviderSegmentationID: strconv.FormatInt(subnet.Vlan, 10),
			RouterExternal:         networkType.GetBool(),
			Shared:                 false,
			Status:                 "ACTIVE",
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
		// process empty network
		if subnet.Network == "" {
			continue
		}
		creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
		gateSub := strings.Split(subnet.Gateway, `/`)
		network := fmt.Sprintf(`%s/%s`, subnet.Network, gateSub[1])
		subnetItem := &restModels.Subnet{
			Cidr:           network,
			CreatedAt:      creatAt,
			EnableDhcp:     true,
			GatewayIP:      strfmt.IPv4(gateSub[0]),
			ID:             strconv.FormatInt(subnet.ID, 10),
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
	gatewayStr := fmt.Sprintf("%s/%d", gateway, netmaskSize)

	var subnet *model.Subnet
	var count int
	existingSubnet := &model.Subnet{}
	if networkUUID == "" {
		// if netID is empty, create subnet with random vni ID and network type is internal
		networkUUID = uuidPk.New().String()
		subnet, err = subnetAdmin.Create(c.Req.Context(), name, "", network.String(), netmask, gateway.String(), first.String(), last.String(), INTERNAL.String(), networkUUID, 0)
		if err != nil {
			log.Println(fmt.Sprintf("Failed to create subnet with error: %v", err))
			c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
			return
		}
		log.Println("success to create subnet: %s", networkUUID)
	} else {
		// if network is unexisting, create subnet with vlxan network type and internal
		if err := db.Model(&model.Subnet{}).Where("uuid = ?", networkUUID).Count(&count).Error; err != nil {
			log.Println(fmt.Sprintf("Failed to query existing vlan, %v", err))
			c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
			return
		}
		if count == 0 {
			// if network don't created, create subnet directly with specified network ID
			subnet, err = subnetAdmin.Create(c.Req.Context(), name, "", network.String(), netmask, gateway.String(), first.String(), last.String(), "internal", networkUUID, 0)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to create subnet with error: %v", err))
				c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
				return
			}
			log.Println(fmt.Sprintf("success to create subnet: %s", networkUUID))
		} else {
			// if network has been created and network type is vxlan, check subnet is unique
			db.Where("uuid = ?", networkUUID).First(existingSubnet)
			log.Println(fmt.Sprintf("%+v", existingSubnet))
			if existingSubnet.Vlan > MAXVLAN && existingSubnet.Network != "" {
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

		//if network has been created with empty subnet, update subnet
		if subnet == nil && existingSubnet.Network == "" && existingSubnet.Vlan != 0 {
			//update subnet with vxlan and vlan type
			existingSubnet.Start = first.String()
			existingSubnet.End = last.String()
			existingSubnet.Gateway = gatewayStr
			existingSubnet.Netmask = netmask
			existingSubnet.Network = network.String()
			db.Save(existingSubnet)
			//create ipaddress for subnet
			subnet = existingSubnet
			ip := first
			for {
				ipstr := fmt.Sprintf("%s/%d", ip.String(), netmaskSize)
				address := &model.Address{Address: ipstr, Netmask: netmask, Type: "ipv4", SubnetID: subnet.ID}
				err = db.Create(address).Error
				if err != nil {
					log.Println("Database create address failed, %v", err)
				}
				if ip.String() == last.String() {
					break
				}
				ip = cidrFun.Inc(ip)
				if ipstr == gatewayStr {
					ip = cidrFun.Inc(ip)
				}
			}
			log.Println(fmt.Sprintf("success update subnet: %s", networkUUID))
		}
		// if network type is vlan and network has been created with a no-empty subnet, create a subnet with same network uuid
		if subnet == nil && existingSubnet.Network != "" &&
			(existingSubnet.Vlan < MAXVLAN && existingSubnet.Vlan >= MINVLAN) {
			//before created vlan subnet with existing network UUID , check the subnet is unique
			subnets := []*model.Subnet{}
			db.Where("uuid = ?", networkUUID).Find(&subnets)
			for _, sub := range subnets {
				//check subnet conflict issue
				_, subNet, _ := net.ParseCIDR(sub.Gateway)
				log.Printf("%+v", cidr)
				log.Printf("%+v", subNet)
				if cidr.Contains(net.ParseIP(sub.Start)) ||
					cidr.Contains(net.ParseIP(sub.End)) ||
					subNet.Contains(first) ||
					subNet.Contains(last) {
					//	if err := cidrFun.VerifyNoOverlap([]*net.IPNet{cidr}, subNet); err != nil {
					c.JSON(
						500,
						NewResponseError(
							"create subnet fail",
							fmt.Sprintf("duplicate subnet: %d", existingSubnet.ID),
							500,
						),
					)
					return
				}
			}
			subnet, err = subnetAdmin.Create(c.Req.Context(), name, strconv.FormatInt(existingSubnet.Vlan, 10), network.String(), netmask, gateway.String(), first.String(), last.String(), existingSubnet.Type, existingSubnet.UUID, 0)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to create subnet with error: %v", err))
				c.JSON(500, NewResponseError("create subnet fail", err.Error(), 500))
				return
			}
			log.Println(fmt.Sprintf("success to create subnet: %s", networkUUID))
		}
	}

	creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
	updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
	subnetRespons := &restModels.CreateSubnetOKBody{
		Subnet: &restModels.Subnet{
			Cidr:           subnet.Network,
			CreatedAt:      creatAt,
			EnableDhcp:     true,
			GatewayIP:      strfmt.IPv4(subnet.Gateway),
			ID:             strconv.FormatInt(subnet.ID, 10),
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

//DeleteSubnet with following logica
//if subnet type is vxlan , reset subnet infor to nil
//if subnet type is vlan, and network include more than 1 subnet ,
//directly delete this subnet ,otherwise reset network of subnet to nil
func (v *SubnetRest) DeleteSubnet(c *macaron.Context) {
	db := DB()
	db = db.Begin()
	var err error
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	// check all of ipaddress in this subnet is idle status
	if isUsed, err := v.checkIPaddresIsUnused(db, id); err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	} else if isUsed {
		errMsg := fmt.Sprintf("Failed to delete subnet: %s, ipaddress in subnet is used", id)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, NewResponseError("Delete subnet fail", errMsg, http.StatusInternalServerError))
		return
	}

	subnet := &model.Subnet{}
	if err = db.Where("id = ?", id).First(subnet).Error; err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}

	if subnet.Vlan > MAXVLAN {
		// reset subnet infor  if vlan type is vxlan
		newSubnet := &model.Subnet{
			Model:  subnet.Model,
			Name:   subnet.Name,
			Vlan:   subnet.Vlan,
			Type:   subnet.Type,
			Router: subnet.Router,
		}
		db.Save(newSubnet)
	} else {
		subnets := []*model.Subnet{}
		db.Where("uuid = ?", subnet.UUID).Find(&subnets)
		if len(subnets) == 1 {
			//reset subnet infor if just only have one subnet in network
			newSubnet := &model.Subnet{
				Model:  subnet.Model,
				Name:   subnet.Name,
				Vlan:   subnet.Vlan,
				Type:   subnet.Type,
				Router: subnet.Router,
			}
			db.Save(newSubnet)
		} else {
			db.Delete(&subnet)
		}
	}
	// start delete ip address
	err = db.Where("subnet_id = ?", subnet.ID).Delete(model.Address{}).Error
	if err != nil {
		log.Println("Database delete ip address failed, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.Status(204)
	return
}

func (v *SubnetRest) checkIPaddresIsUnused(db *gorm.DB, subnetID string) (isUsed bool, err error) {
	count := 0
	err = db.Model(&model.Address{}).Where("subnet_id = ? and allocated = ?", subnetID, true).Count(&count).Error
	if err != nil {
		log.Println("Database delete addresses failed, %v", err)
		return
	}
	if count > 0 {
		err = fmt.Errorf("Some addresses of this subnet in use")
		log.Println("There are addresses of this subnet still in use")
		return true, err
	}
	return
}
