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
	"github.com/go-macaron/session"
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

type VersionsAdmin restModels.Versions
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

func (v *Versions) ListVersion(c *macaron.Context, store session.Store) {
	va := &VersionsAdmin{}
	va.v3(c)
	c.JSON(200, va)
	return
}

func (v *VersionsAdmin) v3(c *macaron.Context) {
	url := &url.URL{
		Scheme: "http", //todo : need to get scheme by config file
		Host:   viper.GetString("api.endpoint"),
		Path:   "/identity/v3/",
	}
	rel := "remote"
	if hostname, _ := os.Hostname(); c.Req.Host == hostname {
		rel = "self"
	}

	updatedDate, _ := strfmt.ParseDateTime(`2015-11-06T14:32:17.893797Z`)
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
			Updated: updatedDate,
		},
	)
}

func (t *Token) IssueTokenByPasswd(c *macaron.Context, store session.Store) {
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
	c.JSON(200, restModels.Token{
		Catalog: []*restModels.TokenCatalogItems{
			&restModels.TokenCatalogItems{
				Endpoints: []*restModels.TokenCatalogItemsEndpointsItems{
					&restModels.TokenCatalogItemsEndpointsItems{
						ID:        `d111111`,
						Interface: `public`,
						Region:    `RegionOne`,
						RegionID:  `RegionOne`,
						URL:       `self`,
					},
				},
			},
		},
	})
	return
}

func JsonSchemeCheck(schemeName string, requestBody []byte) (e *Error) {
	schemeLocation := `./web/rest-api/scheme` + schemeName
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
		return
	} else if !result.Valid() {
		e = &Error{
			Title:   "Validate Json Scheme Fail",
			Code:    400,
			Message: err.Error(),
		}
		return
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
