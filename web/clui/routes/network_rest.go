/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-openapi/strfmt"
	"github.com/jinzhu/gorm"
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
	_, oid, err := ChecKPermissionWithErrorResp(model.Reader, c)
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
	db := DB()
	_, oid, err := ChecKPermissionWithErrorResp(model.Writer, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
	uuid := c.Params("id")
	if uuid == "" {
		log.Println("empty network uuid")
		code := http.StatusBadRequest
		c.JSON(code, NewResponseError("empty network uuid", "uuid is empty", code))
		return
	}
	network := &model.Network{
		Model: model.Model{Owner: oid, UUID: uuid},
	}
	err = db.Preload("Subnets").Where(network).Take(network).Error
	if err != nil {
		code := http.StatusInternalServerError
		if gorm.IsRecordNotFoundError(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, NewResponseError(fmt.Sprint("fail delete network: %d", uuid), err.Error(), code))
		return
	}
	//check network without subnet attached
	if len(network.Subnets) != 0 {
		code := http.StatusConflict
		c.JSON(code, NewResponseError(fmt.Sprint("fail delete network: %d", uuid), "network is not empty", code))
		return
	}
	//start delete network
	err = db.Delete(network).Error
	if err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError(fmt.Sprint("fail delete network: %d", uuid), err.Error(), code))
		return
	}
	c.Status(http.StatusNoContent)
	return
}

func (v *NetworkRest) CreateNetwork(c *macaron.Context) {
	db := DB()
	uid, oid, err := ChecKPermissionWithErrorResp(model.Writer, c)
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
	//subnet := &model.Subnet{}
	if err = db.Preload("Subnets").Where("owner = ?", orgID).Find(&networks).Error; err != nil {
		return
	}
	return
}
