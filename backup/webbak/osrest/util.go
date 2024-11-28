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
	"time"

	"net/url"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/goombaio/namegenerator"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
	macaron "gopkg.in/macaron.v1"
)

const defaultSchema = "https"

type ResponseError struct {
	ErrorMsg `json:"error"`
}

type ErrorMsg struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Title   string `json:"title"`
}

func (e ErrorMsg) Error() string {
	return fmt.Sprintf("(%s): %d - %s", e.Title, e.Code, e.Message)
}

func formateStringToInt64(t string) (result int64, err error) {
	if t == "" {
		return result, fmt.Errorf("empty string")
	}
	changed, err := strconv.Atoi(t)
	if err != nil {
		return result, err
	}
	return int64(changed), nil
}

func NewResponseError(title, msg string, code int) *ResponseError {
	return &ResponseError{
		ErrorMsg: ErrorMsg{
			Title:   title,
			Code:    code,
			Message: msg,
		},
	}
}

func JsonSchemeCheck(schemeName string, requestBody []byte) (e *ErrorMsg) {
	schemeLocation := `../rest-api/scheme/` + schemeName
	if _, err := os.Stat(schemeLocation); os.IsNotExist(err) {
		e = &ErrorMsg{
			Title:   "Load Json Scheme Fail",
			Code:    500,
			Message: fmt.Sprintf("locate json scheme fail with path %s", schemeLocation),
		}
		return
	} else if err != nil {
		e = &ErrorMsg{
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
		e = &ErrorMsg{
			Title:   "Validate Json Scheme Internal Error",
			Code:    http.StatusInternalServerError,
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
		e = &ErrorMsg{
			Title:   "Validate Json Scheme Fail",
			Code:    http.StatusBadRequest,
			Message: errMsg,
		}
	}
	return
}

func respError(c *macaron.Context, code int) {
	c.Error(code, http.StatusText(code))
	return
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

func generateName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func CheckResIfExistByUUID(table, UUID string) (result bool, id int64, err error) {
	db := DB()
	model := &model.Model{}
	err = db.Table(table).Select("id").Where("uuid = ?", UUID).Scan(model).Error
	if err != nil {
		return
	}
	if model.ID == 0 {
		return false, id, nil
	}
	id = model.ID
	return true, id, nil
}

// check resource whether existing
// if resource is not exist, return 404 error
// if resource is exist, save ID to data with UUID as key
func CheckResWithErrorResponse(table, uuid string, c *macaron.Context) (err error) {
	result, id, err := CheckResIfExistByUUID(table, uuid)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return err
	} else if !result {
		code := http.StatusNotFound
		c.Error(
			code,
			NewResponseError(fmt.Sprintf("check %s fail", table), fmt.Sprintf("can't find %s : %s", table, uuid), code).Error(),
		)
		return fmt.Errorf("fail to check resource")
	}
	c.Data[uuid] = id
	return
}

func CheckRoleByUUID(uid, oid int64, expectRole model.Role) (passed bool, err error) {
	db := DB()
	member := &model.Member{
		UserID: uid,
		OrgID:  oid,
	}
	if err = db.First(member).Error; err != nil {
		return
	}
	if member.Role < expectRole {
		return
	}
	return true, nil
}

func ChecKPermissionWithErrorResp(expectRole model.Role, c *macaron.Context) (uid, oid int64, err error) {
	claims := c.Data[ClaimKey].(*HypercubeClaims)
	//check role
	if claims.Role < expectRole {
		// if token was issued before promote user privilige, the user need to re-apply token
		// old token will was refused
		c.Error(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		err = fmt.Errorf("the user role has been changed, need re-apply token")
		return
	}
	uid = c.Data[claims.UID].(int64)
	oid = c.Data[claims.OID].(int64)
	var result bool
	if result, err = CheckRoleByUUID(uid, oid, model.Writer); err != nil {
		code := http.StatusInternalServerError
		if gorm.IsRecordNotFoundError(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, NewResponseError("fail to check permission", err.Error(), code))
		return
	} else if !result {
		errMsg := http.StatusText(http.StatusForbidden)
		c.Error(http.StatusForbidden, errMsg)
		err = fmt.Errorf(errMsg)
		return
	}
	return
}
