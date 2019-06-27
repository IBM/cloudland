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
	total, subnets, err := subnetAdmin.List(offset, limit, order)
	if err != nil {
		c.JSON(500, NewResponseError("List subnets fail", err.Error(), 500))
		return
	}

	networks := &restModels.ListNetworksOKBody{}
	networkItems := []*restModels.NetworksItems{}
	for i := 0; i < len(subnets); i++ {
		network := &restModels.NetworksItems{}
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

func (v *SubnetRest) Create(c *macaron.Context, store session.Store) {
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
