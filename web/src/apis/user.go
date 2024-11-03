/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"

	"github.com/IBM/cloudland/web/src/routes"
	"github.com/gin-gonic/gin"
)

var userAPI = &UserAPI{}
var userAdmin = &routes.UserAdmin{}

type UserAPI struct{}

type UserPayload struct {
	Name     string        `json:"name,required"`
	Password string        `json:"password,required"`
	Org      *Organization `json:"org,omitempty"`
}

type UserInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type User struct {
	UserInfo    *UserInfo     `json:"user_info"`
	OrgInfo     *Organization `json:"org_info"`
	AccessToken string        `json:"access_token"`
	Role        string        `json:"role"`
}

//
// @Summary login to get the accesstoken
// @Description get token by user name
// @tags Authorities
// @Accept  json
// @Produce json
// @Param   message	body   UserPayload  true   "User Credential"
// @Success 200 {object} User
// @Failure 401 {object} APIError "Invalied user name or password"
// @Router /login [post]
func (v *UserAPI) LoginPost(c *gin.Context) {
	payload := &UserPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, &APIError{ErrorMessage: "Input JSON format error"})
		return
	}
	username := payload.Name
	password := payload.Password
	user, err := userAdmin.Validate(c.Request.Context(), username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, &APIError{ErrorMessage: "Invalid username or password"})
		return
	}
	orgName := username
	if payload.Org != nil {
		orgName = payload.Org.Name
	}
	org, err := orgAdmin.Get(orgName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, &APIError{ErrorMessage: "Invalid orgnazation"})
		return
	}
	_, role, token, _, _, err := userAdmin.AccessToken(user.ID, username, orgName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, &APIError{ErrorMessage: "Invalid org with username"})
	}
	userResp := &User{
		UserInfo: &UserInfo{
			Name: username,
			ID:   user.UUID,
		},
		OrgInfo: &Organization{
			Name: orgName,
			ID:   org.UUID,
		},
		AccessToken: token,
		Role:        role.String(),
	}
	c.JSON(http.StatusOK, userResp)
	return
}
