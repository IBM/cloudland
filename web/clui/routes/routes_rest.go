/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"fmt"
	"html/template"
	"log"
	"strings"

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
	}
	unAuthenResources = []string{
		resourceEndpoints["identityToken"],
		resourceEndpoints["identity"],
	}
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
	m.Use(TokenProcess)
	//	m.Use(macaron.Renderer())
	m.Get(resourceEndpoints["identity"], versionInstance.ListVersion)
	m.Post(resourceEndpoints["identityToken"], tokenInstance.IssueTokenByPasswd)
	//neutron network api
	m.Get(resourceEndpoints["network"], subnetInstance.ListNetworks)
	m.Post(resourceEndpoints["network"], subnetInstance.CreateNetwork)
	m.Delete(resourceEndpoints["network"]+`/:id`, subnetInstance.DeleteNetwork)
	//neutron subnet API
	m.Get(resourceEndpoints["subnet"], subnetInstance.ListSubnet)
	m.Post(resourceEndpoints["subnet"], subnetInstance.CreateSubnet)
	m.Delete(resourceEndpoints["subnet"]+`/:id`, subnetInstance.DeleteSubnet)

	return
}

func TokenProcess(c *macaron.Context) {
	for _, unAuthenResource := range unAuthenResources {
		if unAuthenResource == c.Req.RequestURI {
			return
		}
	}
	if claims, err := ParseToken(c.Req.Header.Get("X-Auth-Token")); err != nil {
		log.Println(err.Error())
		c.Error(403, "Unauthenticatins")
		return
	} else {
		log.Println(fmt.Sprintf("%+v", claims))
		// TODO: check token's expire time
		// TODO: save claims data to context  and check action and authority base on userID and orgID
		c.Data["claims"] = claims
	}

	return
}
