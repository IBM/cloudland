/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

var (
	resourceEndpoints = map[string]string{
		"volume":   `/volume/v3`,
		"compute":  `/compute/v2.1`,
		"image":    `/image`,
		"network":  `/network`,
		"identity": `/identity`,
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
	//	m.Use(macaron.Renderer())
	m.Get(resourceEndpoints["identity"], versionInstance.ListVersion)
	m.Post("/identity/v3/auth/tokens", tokenInstance.IssueTokenByPasswd)
	return
}
