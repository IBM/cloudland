/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/jinzhu/gorm"
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

// ListVersion : json home
func (v *Versions) ListVersion(c *macaron.Context) {
	va := &VersionsAdmin{}
	va.v3(c)
	c.Header().Add(`Vary`, `X-Auth-Token`)
	c.JSON(http.StatusMultipleChoices, va)
	return
}

func (v *VersionsAdmin) v3(c *macaron.Context) {
	scheme := viper.GetString("rest.scheme")
	if scheme == "" {
		scheme = defaultSchema
	}
	url := &url.URL{
		Scheme: scheme,
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
	db := DB()
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`token.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{ErrorMsg: *err})
		return
	}
	requestStruct := &restModels.PostIdentityV3AuthTokensParamsBody{}
	if err := json.Unmarshal(body, requestStruct); err != nil {
		c.JSON(http.StatusInternalServerError, NewResponseError("Unmarshal fail", err.Error(), http.StatusInternalServerError))
		return
	}
	username := requestStruct.Auth.Identity.Password.User.Name
	password := requestStruct.Auth.Identity.Password.User.Password
	org := requestStruct.Auth.Scope.Project.Name
	user, err := userAdmin.Validate(c.Req.Context(), username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewResponseError("Authen user fail", err.Error(), http.StatusUnauthorized))
		return
	}
	uid := user.ID
	oid, role, token, issueAt, expiresAt, err := userAdmin.AccessToken(uid, user.Username, org)
	if err != nil {
		c.JSON(http.StatusNotFound, NewResponseError("Failed to get token", err.Error(), http.StatusNotFound))
		return
	}
	c.Header().Add(`X-Subject-Token`, token)
	c.Header().Add(`Vary`, `X-Auth-Token`)
	expire, _ := strfmt.ParseDateTime(time.Unix(expiresAt, 0).Format(time.RFC3339))
	issue, _ := strfmt.ParseDateTime(time.Unix(issueAt, 0).Format(time.RFC3339))
	respInstance := restModels.PostIdentityV3AuthTokensCreatedBody{}

	orgInstance := &model.Organization{Model: model.Model{ID: oid}}
	if err := db.First(orgInstance).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, NewResponseError("Failed to get token", err.Error(), http.StatusNotFound))
			return
		}
		c.JSON(http.StatusInternalServerError, NewResponseError("Failed to get token", err.Error(), http.StatusInternalServerError))
		return
	}
	respInstance.Token = &restModels.Token{
		Catalog:   catLog(orgInstance.UUID),
		ExpiresAt: expire,
		IssuedAt:  issue,
		IsDomain:  false,
		Methods:   []string{"password"},
		Roles:     []*restModels.TokenRolesItems{&restModels.TokenRolesItems{Name: role.String(), ID: orgInstance.UUID}},
		Project: &restModels.TokenProject{
			Domain: &restModels.TokenProjectDomain{
				ID:   "default",
				Name: "default",
			},
			ID:   orgInstance.UUID,
			Name: org,
		},
		User: &restModels.TokenUser{
			Domain: &restModels.TokenUserDomain{
				ID:   "default",
				Name: "default",
			},
			ID:   user.UUID,
			Name: username,
		},
	}
	c.JSON(http.StatusCreated, respInstance)
	return
}

func catLog(orgID string) (items []*restModels.TokenCatalogItems) {
	//hard code resrouce ID , do we need to support this function ?
	url, _ := getRestEndpoint()
	// add volume endpint
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        orgID,
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["volume"],
				},
			},
			Type: "block-storage",
			ID:   orgID,
			Name: "cinder",
		},
	)
	//add compute resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        orgID,
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["compute"],
				},
			},
			Type: "compute",
			ID:   orgID,
			Name: "nova",
		},
	)
	//add image resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        orgID,
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["image"],
				},
			},
			Type: "image",
			ID:   orgID,
			Name: "glance",
		},
	)
	//add network resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        orgID,
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String(),
				},
			},
			Type: "network",
			ID:   orgID,
			Name: "neutron",
		},
	)
	//add keystone resource
	items = append(
		items,
		&restModels.TokenCatalogItems{
			Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
				&restModels.TokenCatalogItemsEndpointsItems{
					ID:        orgID,
					Interface: "public",
					Region:    "default",
					RegionID:  "default",
					URL:       url.String() + resourceEndpoints["identity"],
				},
			},
			Type: "identity",
			ID:   orgID,
			Name: "keystone",
		},
	)
	return
}
