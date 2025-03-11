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

var migrationAPI = &MigrationAPI{}
var migrationAdmin = &routes.MigrationAdmin{}

type MigrationAPI struct{}

type TaskResponse struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type MigrationResponse struct {
	*ResourceReference
	Instance    *InstanceInfo   `json:"instance"`
	SourceHyper int32           `json:"source_hyper"`
	TargetHyper int32           `json:"target_hyper"`
	Force       bool            `json:"force"`
	Type        string          `json:"type"`
	Phases      []*TaskResponse `json:"phases"`
	Status      string          `json:"status"`
}

type MigrationListResponse struct {
	Offset     int                  `json:"offset"`
	Total      int                  `json:"total"`
	Limit      int                  `json:"limit"`
	Migrations []*MigrationResponse `json:"migrations"`
}

type MigrationPayload struct {
	Name        string    `json:"name" binding:"required,min=2,max=32"`
	Instances   []*BaseID `json:"instances" binding:"required,gte=1"`
	Force       bool      `json:"force" binding:"omitempty"`
	TargetHyper *int32     `json:"target_hyper" binding:"omitempty"`
}

// @Summary get a migration
// @Description get a migration
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} MigrationResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /migrations/{id} [get]
func (v *MigrationAPI) Get(c *gin.Context) {
	ctx := c.Request.Context()
	uuID := c.Param("id")
	logger.Debugf("Get migration %s", uuID)
	migration, err := migrationAdmin.GetMigrationByUUID(ctx, uuID)
	if err != nil {
		logger.Errorf("Failed to get migration %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid migration query", err)
		return
	}
	migrationResp, err := v.getMigrationResponse(ctx, migration)
	if err != nil {
		logger.Errorf("Failed to create migration response %s, %+v", uuID, err)
		ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
		return
	}
	logger.Debugf("Get migration %s success, response: %+v", uuID, migrationResp)
	c.JSON(http.StatusOK, migrationResp)
}

// @Summary create a migration
// @Description create a migration
// @tags Compute
// @Accept  json
// @Produce json
// @Param   message	body   MigrationPayload  true   "Migration create payload"
// @Success 200 {array} MigrationResponse
// @Failure 400 {object} common.APIError "Bad request"
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /migrations [post]
func (v *MigrationAPI) Create(c *gin.Context) {
	logger.Debugf("Create migration")
	ctx := c.Request.Context()
	payload := &MigrationPayload{}
	err := c.ShouldBindJSON(payload)
	if err != nil {
		logger.Errorf("Invalid input JSON %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid input JSON", err)
		return
	}
	var instances []*model.Instance
	for _, instRef := range payload.Instances {
		var instance *model.Instance
		instance, err = instanceAdmin.GetInstanceByUUID(ctx, instRef.ID)
		if err != nil {
			logger.Errorf("Failed to get instance %s, %+v", instRef.ID, err)
			ErrorResponse(c, http.StatusBadRequest, "Invalid input, specified instance does not exist", err)
			return
		}
		instances = append(instances, instance)
	}
	targetHyper := int32(-1)
	if payload.TargetHyper != nil {
		targetHyper = *payload.TargetHyper
	}
	logger.Debugf("Creating migration with payload %+v", payload)
	migrations, err := migrationAdmin.Create(ctx, payload.Name, instances, payload.Force, targetHyper)
	if err != nil {
		logger.Errorf("Not able to create migration %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Not able to create", err)
		return
	}
	migrationsResp := make([]*MigrationResponse, len(migrations))
	for i, migration := range migrations {
		migrationsResp[i], err = v.getMigrationResponse(ctx, migration)
		if err != nil {
			logger.Errorf("Failed to create migration response %+v", err)
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	logger.Debugf("Create migration success, response: %+v", migrationsResp)
	c.JSON(http.StatusOK, migrationsResp)
}

func (v *MigrationAPI) getMigrationResponse(ctx context.Context, migration *model.Migration) (migrationResp *MigrationResponse, err error) {
	migrationResp = &MigrationResponse{
		ResourceReference: &ResourceReference{
			ID:        migration.UUID,
			Name:      migration.Name,
			CreatedAt: migration.CreatedAt.Format(TimeStringForMat),
			UpdatedAt: migration.UpdatedAt.Format(TimeStringForMat),
		},
		Force:       migration.Force,
		Type:        migration.Type,
		SourceHyper: migration.SourceHyper,
		TargetHyper: migration.TargetHyper,
		Status:      migration.Status,
	}
	if migration.Instance != nil {
		migrationResp.Instance = &InstanceInfo{
			ResourceReference: &ResourceReference{
				ID: migration.Instance.UUID,
			},
			Hostname: migration.Instance.Hostname,
		}
	}
	migrationResp.Phases = make([]*TaskResponse, len(migration.Phases))
	for i, task := range migration.Phases {
		migrationResp.Phases[i] = &TaskResponse{
			Name:    task.Name,
			Summary: task.Summary,
			Status:  task.Status,
		}
	}
	return
}

// @Summary list migrations
// @Description list migrations
// @tags Compute
// @Accept  json
// @Produce json
// @Success 200 {object} MigrationListResponse
// @Failure 401 {object} common.APIError "Not authorized"
// @Router /migrations [get]
func (v *MigrationAPI) List(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	queryStr := c.DefaultQuery("query", "")
	logger.Debugf("List migrations with offset %s, limit %s, query %s", offsetStr, limitStr, queryStr)
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		logger.Errorf("Invalid query offset %s, %+v", offsetStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset: "+offsetStr, err)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logger.Errorf("Invalid query limit %s, %+v", limitStr, err)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query limit: "+limitStr, err)
		return
	}
	if offset < 0 || limit < 0 {
		logger.Errorf("Invalid query offset or limit %d, %d", offset, limit)
		ErrorResponse(c, http.StatusBadRequest, "Invalid query offset or limit", err)
		return
	}
	total, migrations, err := migrationAdmin.List(int64(offset), int64(limit), "-created_at", queryStr)
	if err != nil {
		logger.Errorf("Failed to list migrations %+v", err)
		ErrorResponse(c, http.StatusBadRequest, "Failed to list migrations", err)
		return
	}
	migrationListResp := &MigrationListResponse{
		Total:  int(total),
		Offset: offset,
		Limit:  len(migrations),
	}
	migrationListResp.Migrations = make([]*MigrationResponse, migrationListResp.Limit)
	for i, migration := range migrations {
		migrationListResp.Migrations[i], err = v.getMigrationResponse(ctx, migration)
		if err != nil {
			logger.Errorf("Failed to create migration response %+v", err)
			ErrorResponse(c, http.StatusInternalServerError, "Internal error", err)
			return
		}
	}
	logger.Debugf("List migrations success, response: %+v", migrationListResp)
	c.JSON(http.StatusOK, migrationListResp)
}
