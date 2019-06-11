/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

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

type ResponseError struct {
	Error Error `json:"error"`
}

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Title   string `json:"title"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("(%d): %d - %s", e.Title, e.Code, e.Message)
}

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
		Host:   viper.GetString("rest.listen"),
		Path:   "/identity/v3/",
	}
	rel := "self"
	fmt.Println(viper.GetString("rest.listen"))
	fmt.Println("111111")
	updatedDate, _ := strfmt.ParseDateTime(`2015-11-06T14:32:17.893797Z`)
	v.Versions = &restModels.Versions{}
	v.Versions.Values = append(
		v.Versions.Values,
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
	_, _, token, err := userAdmin.AccessToken(uid, username, organization)
	//	oid, role, token, err := userAdmin.AccessToken(uid, username, organization)

	if err != nil {
		c.JSON(403, NewResponseError("Failed to get token", err.Error(), 403))
		return
	}
	c.Header().Add(`X-Subject-Token`, token)
	c.Header().Add(`Vary`, `X-Auth-Token`)
	respInstance := restModels.PostIdentityV3AuthTokensCreatedBody{}
	respInstance.Token = &restModels.Token{
		Catalog:  catLog(),
		IsDomain: false,
		Methods:  []string{"password"},
		Roles:    []*restModels.TokenRolesItems{&restModels.TokenRolesItems{Name: "member", ID: "1841f2adad3a4b4aa6485fb4e3a3fda1"}},
		Project: &restModels.TokenProject{
			Domain: &restModels.TokenProjectDomain{
				ID:   "default",
				Name: "default",
			},
			ID:   "default",
			Name: "default",
		},
		User: &restModels.TokenUser{
			Domain: &restModels.TokenUserDomain{
				ID:   "default",
				Name: "default",
			},
			ID:   "b6c55db5d9294824bac2d2d535db92a4",
			Name: "demo",
		},
	}
	c.JSON(200, respInstance)
	return

	// response := `{"token": {"is_domain": false, "methods": ["password"], "roles": [{"id": "1841f2adad3a4b4aa6485fb4e3a3fda1", "name": "member"}, {"id": "488cb631325b48209f4070eb15770c8c", "name": "reader"}, {"id": "17f422d60638440299dc9f328530ef79", "name": "anotherrole"}], "expires_at": "2019-06-06T09:47:56.000000Z", "project": {"domain": {"id": "default", "name": "Default"}, "id": "3b6825911eaa44059c525888dc32713b", "name": "demo"}, "catalog": [{"endpoints": [{"url": "http://192.168.10.16/volume/v2/3b6825911eaa44059c525888dc32713b", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "f0509f4f384b4a8e8aa664fe9085a428"}], "type": "volumev2", "id": "01e9882c890848e79ce13616d5eafb25", "name": "cinderv2"}, {"endpoints": [{"url": "http://192.168.10.16/volume/v3/3b6825911eaa44059c525888dc32713b", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "c4d6fd85cdb643b0bde67ad891a074f6"}], "type": "block-storage", "id": "09e58e3d2207402c84578a6ff1b798cd", "name": "cinder"}, {"endpoints": [{"url": "http://192.168.10.16/compute/v2.1", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "eec0d5080b334f70bc00cd787d5269f6"}], "type": "compute", "id": "182b9192d5854c289cff7adb98415e0f", "name": "nova"}, {"endpoints": [{"url": "http://192.168.10.16:9696/", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "e825c6eafa3343aa83d10b370a6667a2"}], "type": "network", "id": "44713bed353d4684a608901dfb6f20e6", "name": "neutron"}, {"endpoints": [{"url": "http://192.168.10.16/image", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "0c04e1ff2cbc4fe29a58ae8efe743be4"}], "type": "image", "id": "58e590825bbc416fa230b6bc73344375", "name": "glance"}, {"endpoints": [{"url": "http://192.168.10.16/compute/v2/3b6825911eaa44059c525888dc32713b", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "da7e345677054d83b4c3e758f5f51ed1"}], "type": "compute_legacy", "id": "92cf651ee4b847a3af88c328a39beef0", "name": "nova_legacy"}, {"endpoints": [{"url": "http://192.168.10.16/volume/v1/3b6825911eaa44059c525888dc32713b", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "8b91e2d1477f432a8b3f46ce4d56107a"}], "type": "volume", "id": "a722d8586325401dacd321d1553ce000", "name": "cinder"}, {"endpoints": [{"url": "http://192.168.10.16/placement", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "4e8ef73044a84a4b97e443583fb69a75"}], "type": "placement", "id": "af95cfb6712c4914905737f4febecfba", "name": "placement"}, {"endpoints": [{"url": "http://192.168.10.16/identity", "interface": "admin", "region": "RegionOne", "region_id": "RegionOne", "id": "09cef1a83c36456987dd7e1c09b21014"}, {"url": "http://192.168.10.16/identity", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "814817bfd9964b64aafe5a69508057e4"}], "type": "identity", "id": "d8d0f669f8cc4ff5a5633d6ad5746e63", "name": "keystone"}, {"endpoints": [{"url": "http://192.168.10.16/volume/v3/3b6825911eaa44059c525888dc32713b", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "id": "b17ea16a471643f1abb3c7652b445dfe"}], "type": "volumev3", "id": "db9895a3f6b840d98581dbee86f49069", "name": "cinderv3"}], "user": {"password_expires_at": null, "domain": {"id": "default", "name": "Default"}, "id": "b6c55db5d9294824bac2d2d535db92a4", "name": "demo"}, "audit_ids": ["FQnlIQ2QQK2v1BZBNKhAzQ"], "issued_at": "2019-06-06T08:47:56.000000Z"}}`
	// v := &restModels.PostIdentityV3AuthTokensCreatedBody{}
	// json.Unmarshal([]byte(response), v)
	// c.JSON(200, v)
}

func JsonSchemeCheck(schemeName string, requestBody []byte) (e *Error) {
	schemeLocation := `../rest-api/scheme/` + schemeName
	if _, err := os.Stat(schemeLocation); os.IsNotExist(err) {
		e = &Error{
			Title:   "Load Json Scheme Fail",
			Code:    500,
			Message: fmt.Sprintf("locate json scheme fail with path %s", schemeLocation),
		}
		return
	} else if err != nil {
		e = &Error{
			Title:   "Load Json Scheme Fail",
			Code:    500,
			Message: err.Error(),
		}
		return
	}
	if schemeLoaders[schemeName] == nil {
		schemeLoaders[schemeName] = gojsonschema.NewReferenceLoader(`file://` + schemeLocation)
	}
	requestBodyLoader := gojsonschema.NewBytesLoader(requestBody)
	if result, err := gojsonschema.Validate(schemeLoaders[schemeName], requestBodyLoader); err != nil {
		e = &Error{
			Title:   "Validate Json Scheme Internal Error",
			Code:    500,
			Message: err.Error(),
		}
	} else if !result.Valid() {
		errMsg := ""
		for index, desc := range result.Errors() {
			if index == 0 {
				errMsg = desc.String()
				continue
			}
			errMsg = errMsg + ", " + desc.String()
		}
		e = &Error{
			Title:   "Validate Json Scheme Fail",
			Code:    400,
			Message: errMsg,
		}
	}
	return
}

func NewResponseError(title, msg string, code int) ResponseError {
	return ResponseError{
		Error: Error{
			Title:   title,
			Code:    code,
			Message: msg,
		},
	}
}

func catLog() (items []*restModels.TokenCatalogItems) {
	//hard code resrouce ID , do we need to support this function ?
	url := &url.URL{
		Scheme: "http", //todo : need to get scheme by config file
		Host:   viper.GetString("rest.listen"),
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
					URL:       url.String() + resourceEndpoints["network"],
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
