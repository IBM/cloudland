/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

var (
	resourceEndpoints = map[string]string{
		"volume":        `/volume/v3`,
		"compute":       `/compute/v2.1`,
		"image":         `/image`,
		"network":       `/v2.0/networks`,
		"identity":      `/identity`,
		"subnet":        `/v2.0/subnets`,
		"identityToken": "/identity/v3/auth/tokens",
		"flavor":        "/compute/v2.1/flavors",
	}
	unAuthenResources = []string{
		resourceEndpoints["identityToken"],
		resourceEndpoints["identity"],
	}
	ReqIDKey = "x-openstack-request-id"
	ClaimKey = "claims"
)

func RunRest() (err error) {
	Rest().Run(runArgs("rest.listen")...)
	return
}

func Rest() (m *macaron.Macaron) {
	m = macaron.Classic()
	m.Use(macaron.Renderer(
		macaron.RenderOptions{
			Funcs: []template.FuncMap{
				template.FuncMap{
					"GetString": viper.GetString,
					"Title":     func(v interface{}) string { return strings.Title(fmt.Sprint(v)) },
				},
			},
		},
	))
	//token middleware
	m.Use(AddRequestID)
	m.Use(TokenProcess)
	//	m.Use(macaron.Renderer())
	m.Get(resourceEndpoints["identity"], versionInstance.ListVersion)
	m.Post(resourceEndpoints["identityToken"], tokenInstance.IssueTokenByPasswd)
	//neutron network api
	m.Get(resourceEndpoints["network"], networkInstance.ListNetworks)
	m.Post(resourceEndpoints["network"], networkInstance.CreateNetwork)
	m.Delete(resourceEndpoints["network"]+`/:id`, networkInstance.DeleteNetwork)
	//neutron subnet API
	m.Get(resourceEndpoints["subnet"], subnetInstance.ListSubnets)
	m.Post(resourceEndpoints["subnet"], subnetInstance.CreateSubnet)
	m.Delete(resourceEndpoints["subnet"]+`/:id`, subnetInstance.DeleteSubnet)
	//nova flavor
	m.Get(resourceEndpoints["flavor"]+`/detail`, flavorInstance.ListFlavorsDetail)
	m.Get(resourceEndpoints["flavor"], flavorInstance.ListFlavors)
	m.Post(resourceEndpoints["flavor"], flavorInstance.Create)
	m.Delete(resourceEndpoints["flavor"], flavorInstance.Delete)
	return
}

func TokenProcess(c *macaron.Context) {
	for _, unAuthenResource := range unAuthenResources {
		if unAuthenResource == c.Req.RequestURI {
			return
		}
	}
	claims, err := ParseToken(c.Req.Header.Get("X-Auth-Token"))
	if err != nil {
		log.Println(err.Error())
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else {
		c.Data[ClaimKey] = claims
	}
	// check org whether existing
	if err = CheckResWithErrorResponse(model.Organization{}.TableName(), claims.OID, c); err != nil {
		// org is not exist, return error response
		return
	}
	// check user whether existing
	if err = CheckResWithErrorResponse(model.User{}.TableName(), claims.UID, c); err != nil {
		// user is not exiit, return error response
		return
	}
	memberShip, err := GetDBMemberShip(c.Data[claims.OID].(int64), c.Data[claims.UID].(int64))
	if err != nil {
		return
	}
	c.Req.Request = c.Req.WithContext(memberShip.SetContext(c.Req.Context()))
	return
}

func AddRequestID(c *macaron.Context) {
	requestID := c.Req.Header.Get(ReqIDKey)
	if requestID == "" {
		requestID = `req-` + uuid.New().String()
		c.Req.Header.Set(ReqIDKey, requestID)
		c.Resp.Header().Set(ReqIDKey, requestID)
	} else {
		c.Resp.Header().Set(ReqIDKey, requestID)
	}
}
