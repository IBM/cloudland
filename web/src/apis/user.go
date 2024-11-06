/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"

	"web/src/routes"
	"github.com/gin-gonic/gin"
)

var userAPI = &UserAPI{}
var userAdmin = &routes.UserAdmin{}

type UserAPI struct{}

type UserPayload struct {
	Username string         `json:"username,required"`
	Password string         `json:"password,required"`
	Org      *BaseReference `json:"org,omitempty"`
}

type UserPatchPayload struct {
	Password string         `json:"password,required"`
}

type UserResponse struct {
	UserInfo    *BaseReference `json:"user"`
	OrgInfo     *BaseReference `json:"org"`
	AccessToken string         `json:"token"`
	Role        string         `json:"role"`
}

type UserListResponse struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Users   []*OrgResponse `json:"users"`
}

//
// @Summary get a user
// @Description get a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /users/:id [get]
func (v *UserAPI) Get(c *gin.Context) {
	userResp := &UserResponse{}
	c.JSON(http.StatusOK, userResp)
}

//
// @Summary patch a user
// @Description patch a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   UserPatchPayload  true   "User patch payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /users/:id [patch]
func (v *UserAPI) Patch(c *gin.Context) {
	userResp := &UserResponse{}
	c.JSON(http.StatusOK, userResp)
}

//
// @Summary delete a user
// @Description delete a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /users/:id [delete]
func (v *UserAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

//
// @Summary create a user
// @Description create a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   UserPayload  true   "User create payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Bad request"
// @Failure 401 {object} APIError "Not authorized"
// @Router /users [post]
func (v *UserAPI) Create(c *gin.Context) {
	userResp := &UserResponse{}
	c.JSON(http.StatusOK, userResp)
}

//
// @Summary list users
// @Description list users
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} UserListResponse
// @Failure 401 {object} APIError "Not authorized"
// @Router /users [get]
func (v *UserAPI) List(c *gin.Context) {
	userListResp := &UserListResponse{}
	c.JSON(http.StatusOK, userListResp)
}
//
// @Summary login to get the accesstoken
// @Description get token by user name
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   UserPayload  true   "User Credential"
// @Success 200 {object} UserResponse
// @Failure 401 {object} APIError "Invalied user name or password"
// @Router /login [post]
func (v *UserAPI) LoginPost(c *gin.Context) {
	payload := &UserPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Input JSON format error", err)
		return
	}
	username := payload.Username
	password := payload.Password
	user, err := userAdmin.Validate(c.Request.Context(), username, password)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid username or password", err)
		return
	}
	orgName := username
	if payload.Org != nil {
		orgName = payload.Org.Name
	}
	org, err := orgAdmin.Get(orgName)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid organization", err)
		return
	}
	_, role, token, _, _, err := userAdmin.AccessToken(user.ID, username, orgName)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid organization with username", err)
		return
	}
	userResp := &UserResponse{
		UserInfo: &BaseReference{
			Name: username,
			ID:   user.UUID,
		},
		OrgInfo: &BaseReference{
			Name: orgName,
			ID:   org.UUID,
		},
		AccessToken: token,
		Role:        role.String(),
	}
	c.JSON(http.StatusOK, userResp)
	return
}
