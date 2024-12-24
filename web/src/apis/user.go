/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var userAPI = &UserAPI{}
var userAdmin = &routes.UserAdmin{}

type UserAPI struct{}

type UserPayload struct {
	Username string         `json:"username,required" binding:"required,min=2"`
	Password string         `json:"password,required" binding:"required,min=8,max=32"`
	Org      *BaseReference `json:"org,omitempty" binding:"omitempty"`
}

type UserPatchPayload struct {
	Password string `json:"password,required" binding:"required,min=6"`
}

type UserResponse struct {
	UserInfo    *ResourceReference `json:"user"`
	OrgInfo     *ResourceReference `json:"org,omitempty"`
	AccessToken string             `json:"token,omitempty"`
	Role        string             `json:"role,omitempty"`
}

type UserListResponse struct {
	Offset int             `json:"offset"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Users  []*UserResponse `json:"users"`
}

// @Summary get a user
// @Description get a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /users/{id} [get]
func (v *UserAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	user, err := userAdmin.GetUserByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	userResp := &UserResponse{
		UserInfo: &ResourceReference{
			ID:   user.UUID,
			Name: user.Username,
		},
	}
	c.JSON(http.StatusOK, userResp)
}

// @Summary patch a user
// @Description patch a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   UserPatchPayload  true   "User patch payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /users/{id} [patch]
func (v *UserAPI) Patch(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	payload := &UserPatchPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	user, err := userAdmin.GetUserByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	user, err = userAdmin.Update(ctx, user.ID, payload.Password, nil)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	userResp := &UserResponse{
		UserInfo: &ResourceReference{
			ID:   user.UUID,
			Name: user.Username,
			CreatedAt: user.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: user.UpdatedAt.Format(TimeStringForMat),
		},
	}
	c.JSON(http.StatusOK, userResp)
}

// @Summary delete a user
// @Description delete a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /users/{id} [delete]
func (v *UserAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	user, err := userAdmin.GetUserByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = userAdmin.Delete(ctx, user)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a user
// @Description create a user
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   UserPayload  true   "User create payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /users [post]
func (v *UserAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &UserPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	username := payload.Username
	password := payload.Password
	user, err := userAdmin.Create(ctx, username, password)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err)
		return
	}
	org, err := orgAdmin.Create(ctx, username, username)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create org", err)
		return
	}
	userResp := &UserResponse{
		UserInfo: &ResourceReference{
			ID:   user.UUID,
			Name: username,
			CreatedAt: user.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: user.UpdatedAt.Format(TimeStringForMat),
		},
		OrgInfo: &ResourceReference{
			ID:   org.UUID,
			Name: username,
			CreatedAt: org.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: org.UpdatedAt.Format(TimeStringForMat),
		},
		Role: model.Owner.String(),
	}
	c.JSON(http.StatusOK, userResp)
}

// @Summary list users
// @Description list users
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} UserListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /users [get]
func (v *UserAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, users, err := userAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list vpcs", err)
		return
	}
	userListResp := &UserListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(users),
	}
	userListResp.Users = make([]*UserResponse, userListResp.Limit)
	for i, user := range users {
		userListResp.Users[i] = &UserResponse{
			UserInfo: &ResourceReference{
				ID:   user.UUID,
				Name: user.Username,
				CreatedAt: user.CreatedAt.Format(TimeStringForMat),
				UpdatedAt: user.UpdatedAt.Format(TimeStringForMat),
			},
		}
	}
	c.JSON(http.StatusOK, userListResp)
}

// @Summary login to get the access token
// @Description get token by user name
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   UserPayload  true   "User Credential"
// @Success 200 {object} UserResponse
// @Failure 401 {object} common.APIError "Invalied user name or password"
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
	org, err := orgAdmin.GetOrgByName(orgName)
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
		UserInfo: &ResourceReference{
			Name: username,
			ID:   user.UUID,
		},
		OrgInfo: &ResourceReference{
			Name: orgName,
			ID:   org.UUID,
		},
		AccessToken: token,
		Role:        role.String(),
	}
	c.JSON(http.StatusOK, userResp)
	return
}
