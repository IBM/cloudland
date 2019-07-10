/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"time"

	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
	macaron "gopkg.in/macaron.v1"
)

var (
	versionInstance = &Versions{}
	tokenInstance   = &Token{}
	schemeLoaders   = map[string]gojsonschema.JSONLoader{}
)

type VersionsAdmin restModels.GetIdentityMultipleChoicesBody
type tokenAdmin restModels.PostIdentityV3AuthTokensCreatedBody

type Versions struct{}
type Token struct{}

func (v *Versions) ListVersion(c *macaron.Context) {
	va := &VersionsAdmin{}
	va.v3(c)
	c.Header().Add(`Vary`, `X-Auth-Token`)
	c.JSON(300, va)
	return
}

func (v *VersionsAdmin) v3(c *macaron.Context) {
	url := &url.URL{
		Scheme: "http", //todo : need to get scheme by config file
		Host:   viper.GetString("rest.endpoint"),
		Path:   "/identity/v3/",
	}
	updatedDate, _ := strfmt.ParseDateTime(`2015-11-06T14:32:17.893797Z`)
	v.Versions = &restModels.Versions{}
	urlStr := url.String()
	rel := restModels.LinkRelSelf
	v.Versions.Values = append(
		v.Versions.Values,
		&restModels.VersionsValuesItems{
			ID: `v3.10`,
			Links: restModels.Links{
				&restModels.Link{
					Href: &urlStr,
					Rel:  &rel,
				},
			},
			MediaTypes: []*restModels.VersionsValuesItemsMediaTypesItems{
				&restModels.VersionsValuesItemsMediaTypesItems{
					Base: "application/json",
					Type: "application/vnd.openstack.identity-v3+json",
				},
			},
			Status:  `stable`,
			Updated: updatedDate,
		},
	)
}

func (t *Token) IssueTokenByPasswd(c *macaron.Context) {
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`token.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{
			Error: *err,
		})
		return
	}
	requestStruct := &restModels.PostIdentityV3AuthTokensParamsBody{}
	if err := json.Unmarshal(body, requestStruct); err != nil {
		c.JSON(500, NewResponseError("Unmarshal fail", err.Error(), 403))
	}
	//todo:
	username := requestStruct.Auth.Identity.Password.User.Name
	password := requestStruct.Auth.Identity.Password.User.Password
	user, err := userAdmin.Validate(username, password)
	if err != nil {
		c.JSON(400, NewResponseError("Authen user fail", err.Error(), 400))
		return
	}
	organization := username
	uid := user.ID
	uid, _, token, issueAt, expiresAt, err := userAdmin.AccessToken(uid, username, organization)
	//	oid, role, token, err := userAdmin.AccessToken(uid, username, organization)

	if err != nil {
		c.JSON(403, NewResponseError("Failed to get token", err.Error(), 403))
		return
	}
	c.Header().Add(`X-Subject-Token`, token)
	c.Header().Add(`Vary`, `X-Auth-Token`)
	log.Print(token)
	expire, _ := strfmt.ParseDateTime(time.Unix(expiresAt, 0).Format(time.RFC3339))
	issue, _ := strfmt.ParseDateTime(time.Unix(issueAt, 0).Format(time.RFC3339))
	respInstance := restModels.PostIdentityV3AuthTokensCreatedBody{}
	respInstance.Token = &restModels.Token{
		Catalog:   catLog(),
		ExpiresAt: expire,
		IssuedAt:  issue,
		IsDomain:  false,
		Methods:   []string{"password"},
		Roles:     []*restModels.TokenRolesItems{&restModels.TokenRolesItems{Name: "member", ID: "1841f2adad3a4b4aa6485fb4e3a3fda1"}},
		Project: &restModels.TokenProject{
			Domain: &restModels.TokenProjectDomain{
				ID:   "default",
				Name: "default",
			},
			ID:   organization,
			Name: organization,
		},
		User: &restModels.TokenUser{
			Domain: &restModels.TokenUserDomain{
				ID:   "default",
				Name: "default",
			},
			ID:   strconv.Itoa(int(uid)),
			Name: username,
		},
	}
	c.JSON(200, respInstance)
	return
}

func catLog() (items []*restModels.TokenCatalogItems) {
	//hard code resrouce ID , do we need to support this function ?
	url := &url.URL{
		Scheme: "http", //todo : need to get scheme by config file
		Host:   viper.GetString("rest.endpoint"),
	}
	// add volume endpint
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        "c4d6fd85cdb643b0bde67ad891a074f6",
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["volume"],
				},
			},
			Type: "block-storage",
			ID:   "09e58e3d2207402c84578a6ff1b798cd",
			Name: "cinder",
		},
	)
	//add compute resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        "eec0d5080b334f70bc00cd787d5269f6",
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["compute"],
				},
			},
			Type: "compute",
			ID:   "182b9192d5854c289cff7adb98415e0f",
			Name: "nova",
		},
	)
	//add image resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        "0c04e1ff2cbc4fe29a58ae8efe743be4",
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["image"],
				},
			},
			Type: "image",
			ID:   "58e590825bbc416fa230b6bc73344375",
			Name: "glance",
		},
	)
	//add network resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        "e825c6eafa3343aa83d10b370a6667a2",
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String(),
				},
			},
			Type: "network",
			ID:   "44713bed353d4684a608901dfb6f20e6",
			Name: "neutron",
		},
	)
	//add keystone resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        "09cef1a83c36456987dd7e1c09b21014",
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["identity"],
				},
			},
			Type: "identity",
			ID:   "d8d0f669f8cc4ff5a5633d6ad5746e63",
			Name: "keystone",
		},
	)
	return
}
