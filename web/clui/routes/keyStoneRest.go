/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"net/url"
	"os"

	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/go-macaron/session"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
	macaron "gopkg.in/macaron.v1"
)

var (
	versionInstance = &Versions{}
)

type Versions restModels.Versions

func (v *Versions) ListVersion(c *macaron.Context, store session.Store) {
	versionInstance := &Versions{}
	versionInstance.v3(c, store)
	c.JSON(200, versionInstance)
}

func (v *Versions) v3(c *macaron.Context, store session.Store) {
	url := &url.URL{
		Scheme: "http",
		Host:   viper.GetString("api.endpoint"),
		Path:   "/identity/v3/",
	}
	rel := "remote"
	if hostname, _ := os.Hostname(); c.Req.Host == hostname {
		rel = "self"
	}

	updateDate, _ := strfmt.ParseDateTime(`2015-11-06T14:32:17.893797Z`)
	v.Values = append(
		v.Values,
		&restModels.VersionsValuesItems{
			ID: `v3.10`,
			Links: []*restModels.VersionsValuesItemsLinksItems{
				&restModels.VersionsValuesItemsLinksItems{
					Href: url.String(),
					Rel:  rel,
				},
			},
			MediaTypes: []*restModels.VersionsValuesItemsMediaTypesItems{
				&restModels.VersionsValuesItemsMediaTypesItems{
					Base: "application/json",
					Type: "application/vnd.openstack.identity-v3+json",
				},
			},
			Status:  `stable`,
			Updated: updateDate,
		},
	)
}
