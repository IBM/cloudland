/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"net/url"

	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
	macaron "gopkg.in/macaron.v1"
)

func formateStringToInt64(c *macaron.Context, t string) (result int64, err error) {
	if t == "" {
		return result, nil
	}
	changed, err := strconv.Atoi(t)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return result, err
	}
	return int64(changed), nil
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

func respError(c *macaron.Context, code int) {
	c.Error(code, http.StatusText(code))
	return
}

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

func getRestEndpoint() (*url.URL, error) {
	endpoint := viper.GetString("rest.endpoint")
	if endpoint == "" {
		return nil, fmt.Errorf("fail to get URL")
	}
	if !strings.Contains(endpoint, `//`) {
		endpoint = "//" + endpoint
	}
	urlparsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(strings.TrimSpace(viper.GetString("rest.scheme"))) == "https" {
		urlparsed.Scheme = "https"
	} else {
		urlparsed.Scheme = "http"
	}
	return urlparsed, err
}
