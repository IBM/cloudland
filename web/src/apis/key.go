/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"context"
	"net/http"
	"strconv"

	. "web/src/common"
	"web/src/model"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

var keyAPI = &KeyAPI{}
var keyAdmin = &routes.KeyAdmin{}

type KeyAPI struct{}

type KeyResponse struct {
	*ResourceReference
	FingerPrint string `json:"finger_print"`
	PublicKey   string `json:"public_key"`
}

type KeyListResponse struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Keys   []*KeyResponse `json:"keys"`
}

type KeyPayload struct {
	Name      string `json:"name" binding:"required,min=2,max=32"`
	PublicKey string `json:"public_key" binding:"required,min=4,max=4096"`
}

type KeyPatchPayload struct {
	Name string `json:"name" binding:"required,min=2,max=32"`
}

// @Summary get a key
// @Description get a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} KeyResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys/{id} [get]
func (v *KeyAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	key, err := keyAdmin.GetKeyByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid vpc query", err)
		return
	}
	keyResp, err := v.getKeyResponse(ctx, key)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, keyResp)
}

// @Summary patch a key
// @Description patch a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   KeyPatchPayload  true   "Key patch payload"
// @Success 200 {object} KeyResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys/{id} [patch]
func (v *KeyAPI) Patch(c *gin.Context) {
	keyResp := &KeyResponse{}
	c.JSON(http.StatusOK, keyResp)
}

// @Summary delete a key
// @Description delete a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 204
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys/{id} [delete]
func (v *KeyAPI) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	key, err := keyAdmin.GetKeyByUUID(ctx, uuID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid query", err)
		return
	}
	err = keyAdmin.Delete(ctx, key)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to delete", err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary create a key
// @Description create a key
// @tags Authorization
// @Accept  json
// @Produce json
// @Param   message	body   KeyPayload  true   "Key create payload"
// @Success 200 {object} KeyResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys [post]
func (v *KeyAPI) Create(c *gin.Context) {
	ctx := c.Request.Context()
	payload := &KeyPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	key, err := keyAdmin.Create(ctx, payload.Name, payload.PublicKey)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	keyResp, err := v.getKeyResponse(ctx, key)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	c.JSON(http.StatusOK, keyResp)
}

func (v *KeyAPI) getKeyResponse(ctx context.Context, key *model.Key) (keyResp *KeyResponse, err error) {
	owner := orgAdmin.GetOrgName(key.Owner)
	keyResp = &KeyResponse{
		ResourceReference: &ResourceReference{
			ID:    key.UUID,
			Name:  key.Name,
			Owner: owner,
			CreatedAt: key.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: key.UpdatedAt.Format(TimeStringForMat),
		},
		FingerPrint: key.FingerPrint,
		PublicKey:   key.PublicKey,
	}
	return
}

// @Summary list keys
// @Description list keys
// @tags Authorization
// @Accept  json
// @Produce json
// @Success 200 {object} KeyListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /keys [get]
func (v *KeyAPI) List(c *gin.Context) {
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
	total, keys, err := keyAdmin.List(ctx, int64(offset), int64(limit), "-created_at", "")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to list vpcs", err)
		return
	}
	keyListResp := &KeyListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(keys),
	}
	keyListResp.Keys = make([]*KeyResponse, keyListResp.Limit)
	for i, key := range keys {
		keyListResp.Keys[i], err = v.getKeyResponse(ctx, key)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	c.JSON(http.StatusOK, keyListResp)
}
