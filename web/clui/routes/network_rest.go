/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/go-openapi/strfmt"
	"github.ibm.com/cland/dbs"
	macaron "gopkg.in/macaron.v1"
)

const MAXVLAN int64 = 4096
const MINVLAN int64 = 1

var (
	networkInstance = &NetworkRest{}
	networkAdmin    = NetworkAdmin{}
)

type NetworkRest struct{}
type NetworkAdmin struct{}

type networkType string

const VLAN networkType = `vlan`
const VXLAN networkType = `vxlan`

var totalSubnets int64 = 50

func (v networkType) String() string {
	return string(v)
}

func (v *NetworkRest) ListNetworks(c *macaron.Context) {
	_, oid, err := ChecPermissionWithErrorResp(model.Writer, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
	// offset := c.QueryInt64("marker")
	// limit := c.QueryInt64("limit")
	reverse := c.QueryBool("page_reverse")
	order := "uuid,created_at"
	if reverse {
		order = "uuid,-created_at"
	}
	var networks []*model.Network
	_, networks, err = networkAdmin.List(oid, 0, totalSubnets, order)
	if err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("List network fail", err.Error(), code))
		return
	}
	networksOK := &restModels.ListNetworksOKBody{
		Networks: restModels.Networks{},
	}
	for _, network := range networks {
		creatAt, _ := strfmt.ParseDateTime(network.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(network.UpdatedAt.Format(time.RFC3339))
		networkItem := &restModels.Network{
			AdminStateUp:           true,
			CreatedAt:              creatAt,
			AvailabilityZones:      []string{"cloudland"},
			ID:                     network.UUID,
			Name:                   network.Name,
			Status:                 "Active",
			UpdatedAt:              updateAt,
			ProviderNetworkType:    network.Type,
			ProviderSegmentationID: strconv.FormatInt(network.Vlan, 10),
			Subnets:                []string{},
		}
		for _, subnet := range network.Subnets {
			networkItem.Subnets = append(networkItem.Subnets, subnet.UUID)
		}
		networksOK.Networks = append(networksOK.Networks, networkItem)
	}
	c.JSON(200, networksOK)
}

func (v *NetworkRest) DeleteNetwork(c *macaron.Context) {
	_, _, err := ChecPermissionWithErrorResp(model.Writer, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
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

func (v *NetworkRest) CreateNetwork(c *macaron.Context) {
	db := DB()
	uid, oid, err := ChecPermissionWithErrorResp(model.Writer, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`network.json`, body); err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{ErrorMsg: *err})
		return
	}
	requestData := &restModels.CreateNetworkParamsBody{}
	if err := json.Unmarshal(body, requestData); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			NewResponseError("Unmarshal fail", err.Error(), http.StatusInternalServerError),
		)
		return
	}
	vlanID, err := formateStringToInt64(requestData.Network.ProviderSegmentationID)
	if err != nil {
		code := http.StatusBadRequest
		c.JSON(code, NewResponseError("invalid vlan ID", err.Error(), code))
		return
	}
	if requestData.Network.ProviderNetworkType == "" {
		requestData.Network.ProviderNetworkType = restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVxlan
	}
	// for vlan type , the vlan ID must was provide by user
	if requestData.Network.ProviderNetworkType == restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVxlan && vlanID == 0 {
		vlanNo, err := getValidVni()
		if err != nil {
			code := http.StatusInternalServerError
			c.JSON(code, NewResponseError("can't resever vni", err.Error(), code))
			return
		}
		vlanID = int64(vlanNo)
	}
	if requestData.Network.ProviderNetworkType == restModels.CreateNetworkParamsBodyNetworkProviderNetworkTypeVlan {
		code := http.StatusBadRequest
		if requestData.Network.ProviderSegmentationID == "" {
			c.JSON(code, NewResponseError("must provide vlan ID", "empty vlan id", code))
			return
		}
		if vlanID > MAXVLAN || vlanID < MINVLAN {
			c.JSON(code, NewResponseError("invalid vlan range", "vlan range must be in 1-4096", code))
			return
		}
	}

	//check vni and vlan network whether has been created priviously
	if result, err := checkIfExistVni(vlanID); err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("check vni fail", err.Error(), code))
		return
	} else if result {
		code := http.StatusConflict
		c.JSON(
			code,
			NewResponseError(
				"duplicate vni",
				fmt.Sprintf("the vni %s has been used", requestData.Network.ProviderSegmentationID),
				code,
			),
		)
		return
	}
	//save data to db
	network := &model.Network{
		Model: model.Model{
			Creater: uid,
			Owner:   oid,
		},
		Name:     requestData.Network.Name,
		Vlan:     vlanID,
		Type:     requestData.Network.ProviderNetworkType,
		External: requestData.Network.RouterExternal,
	}
	err = db.Create(network).Error
	if err != nil {
		log.Println("Database create network failed, %v", err)
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("create network fail", err.Error(), code))
		return
	}
	//create response body
	creatAt, _ := strfmt.ParseDateTime(network.CreatedAt.Format(time.RFC3339))
	updateAt, _ := strfmt.ParseDateTime(network.UpdatedAt.Format(time.RFC3339))
	responseBody := &restModels.CreateNetworkOKBody{
		Network: &restModels.Network{
			AdminStateUp:           true,
			AvailabilityZones:      []string{"cloudland"},
			CreatedAt:              creatAt,
			ID:                     network.UUID,
			IsDefault:              false,
			Mtu:                    1500,
			Name:                   network.Name,
			PortSecurityEnabled:    false,
			ProviderNetworkType:    requestData.Network.ProviderNetworkType,
			ProviderSegmentationID: strconv.FormatInt(network.Vlan, 10),
			RouterExternal:         requestData.Network.RouterExternal,
			Shared:                 false,
			Status:                 "ACTIVE",
			Subnets:                []string{},
			UpdatedAt:              updateAt,
			VlanTransparent:        false,
		},
	}
	c.JSON(200, responseBody)
}

func (a *NetworkAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	subnet := &model.Subnet{Model: model.Model{ID: id}}
	err = db.Preload("Netlink").Take(subnet).Error
	if err != nil {
		log.Println("Database failed to query subnet", err)
		return
	}
	if subnet.Router > 0 {
		err = fmt.Errorf("Subnet belongs to a gateway")
		log.Println("Subnet belongs to a gateway", err)
		return
	}
	count := 0
	err = db.Model(&model.Address{}).Where("subnet_id = ? and allocated = ?", id, true).Count(&count).Error
	if err != nil {
		log.Println("Database delete addresses failed, %v", err)
		return
	}
	if count > 2 {
		err = fmt.Errorf("Some addresses of this subnet in use")
		log.Println("There are addresses of this subnet still in use")
		return
	}
	err = db.Model(&model.Subnet{}).Where("vlan = ?", subnet.Vlan).Count(&count).Error
	if err != nil {
		log.Println("Database failed to count subnet", err)
		return
	}
	err = db.Delete(subnet).Error
	if err != nil {
		log.Println("Database delete subnet failed, %v", err)
		return
	}
	//delete ip address
	err = db.Where("subnet_id = ?", id).Delete(model.Address{}).Error
	if err != nil {
		log.Println("Database delete ip address failed, %v", err)
		return
	}
	netlink := subnet.Netlink
	if count <= 1 && netlink != nil {
		err = db.Where("dhcp = ?", netlink.ID).Delete(&model.Interface{}).Error
		if err != nil {
			log.Println("Failed to delete dhcp interfaces")
			return
		}
		err = db.Delete(netlink).Error
		if err != nil {
			log.Println("Failed to delete network")
			return
		}
		control := ""
		if netlink.Hyper >= 0 {
			control = fmt.Sprintf("toall=vlan-%d:%d", subnet.Vlan, netlink.Hyper)
			if netlink.Peer >= 0 {
				control = fmt.Sprintf("%s,%d", control, netlink.Peer)
			}
		} else if netlink.Peer >= 0 {
			control = fmt.Sprintf("inter=%d", netlink.Peer)
		} else {
			log.Println("Network has no valid hypers")
			return
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_net.sh %d %s %d", netlink.Vlan, subnet.Network, subnet.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete interface failed")
			return
		}
	}
	return
}

func (a *NetworkAdmin) List(orgID int64, offset, limit int64, order string) (total int64, networks []*model.Network, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}
	if order == "" {
		order = "created_at"
	}
	networks = []*model.Network{}
	if err = db.Model(&model.Network{}).Where("owner = ?", orgID).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where("owner = ?", orgID).Find(&networks).Error; err != nil {
		return
	}
	return
}
