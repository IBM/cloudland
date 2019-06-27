/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"log"
	"net/http"
	"strconv"

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
			AdminStateUp:      true,
			CreatedAt:         creatAt,
			AvailabilityZones: []string{"nova"},
			ID:                strconv.FormatInt(subnet.ID, 10),
			Name:              subnet.Name,
			Status:            "Active",
			UpdatedAt:         updateAt,
			Provider: &restModels.NetworksItemsProvider{
				NetworkType:    "vxlan",
				SegmentationID: &subnet.Vlan,
			},
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
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`token.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{
			Error: *err,
		})
		return
	}
	requestStruct := &restModels.{}
	if err := json.Unmarshal(body, requestStruct); err != nil {
		c.JSON(500, NewResponseError("Unmarshal fail", err.Error(), 403))
	}

	redirectTo := "../subnets"
	name := c.Query("name")
	vlan := c.Query("vlan")
	rtype := c.Query("rtype")
	network := c.Query("network")
	netmask := c.Query("netmask")
	gateway := c.Query("gateway")
	start := c.Query("start")
	end := c.Query("end")
	_, err := subnetAdmin.Create(name, vlan, network, netmask, gateway, start, end, rtype)
	if err != nil {
		log.Println("Create subnet failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
