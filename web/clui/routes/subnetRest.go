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

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/go-macaron/session"
	"github.com/go-openapi/strfmt"
	macaron "gopkg.in/macaron.v1"
)

var (
	subnetInstance = &SubnetRest{}
)

type SubnetRest struct{}

func (v *SubnetRest) List(c *macaron.Context) {
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
	networkItems := []*restModels.NetworksItems{}
	for _, subnet := range subnets {
		creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.String())
		updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.String())
		network := &restModels.NetworksItems{
			AdminStateUp:           true,
			CreatedAt:              creatAt,
			AvailabilityZones:      []string{"nova"},
			ID:                     strconv.FormatInt(subnet.ID, 10),
			Name:                   subnet.Name,
			Status:                 "Active",
			UpdatedAt:              updateAt,
			ProviderNetworkType:    "vxlan",
			ProviderSegmentationID: &subnet.Vlan,
			Subnets: []string{
				strconv.FormatInt(subnet.ID, 10),
			},
		}
		networkItems = append(networkItems, network)
	}
	networks.Networks = networkItems
	c.JSON(200, networks)
}

func (v *SubnetRest) Delete(c *macaron.Context, store session.Store) (err error) {
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

func (v *SubnetRest) Create(c *macaron.Context) {
	db := DB()
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`token.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{
			Error: *err,
		})
		return
	}
	requestData := &restModels.CreateNetworkParamsBodyNetwork{}
	if err := json.Unmarshal(body, requestData); err != nil {
		c.JSON(500, NewResponseError("Unmarshal fail", err.Error(), 500))
		return
	}
	if result, err := checkIfExistVni(requestData.ProviderSegmentationID); err != nil {
		c.JSON(500, NewResponseError("check vni fail", err.Error(), 500))
	} else if result {
		c.JSON(
			400,
			NewResponseError(
				"duplicate vni",
				fmt.Sprintf("the vni %d has been used", requestData.ProviderSegmentationID),
				400,
			),
		)
		return
	}
	subnet := &model.Subnet{Name: requestData.Name, Vlan: requestData.ProviderSegmentationID, Type: "internal"}
	err := db.Create(subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		c.JSON(500, NewResponseError("create network fail", err.Error(), 500))
		return
	}

}
