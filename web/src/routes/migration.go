/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	MigrationAdmin = &MigrationAdmin{}
	MigrationView  = &MigrationView{}
)

type MigrationAdmin struct{}
type MigrationView struct{}

func FileExist(filename string) bool {
	_, err := os.Lstat(filename)
	return !os.IsNotExist(err)
}

func (a *MigrationAdmin) Create(ctx context.Context, osCode, name, osVersion, virtType, userName, url, architecture string, qaEnabled bool, instID int64) (Migration *model.Migration, err error) {
	logger.Debugf("Creating Migration %s %s %s %s %s %s %s %t %d", osCode, name, osVersion, virtType, userName, url, architecture, qaEnabled, instID)
	memberShip := GetMemberShip(ctx)
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	var instance *model.Instance
	if instID > 0 {
		instance = &model.Instance{Model: model.Model{ID: instID}}
		err = db.Preload("Migration").Preload("Volumes").Take(instance).Error
		if err != nil {
			logger.Error("DB failed to query instance", err)
			return
		}
		if instance.Status != "shut_off" {
			err = fmt.Errorf("instance [%s] is running, shut it down first before capturing", instance.Hostname)
			logger.Error(err)
			return
		}
		Migration = instance.Migration.Clone()
		Migration.Model = model.Model{Creater: memberShip.UserID}
		Migration.Owner = memberShip.OrgID
		Migration.Name = name
		Migration.Status = "creating"
		Migration.CaptureFromInstanceID = instance.ID
		Migration.CaptureFromInstance = instance
	} else {
		Migration = &model.Migration{
			Model:        model.Model{Creater: memberShip.UserID},
			Owner:        memberShip.OrgID,
			OsVersion:    osVersion,
			VirtType:     virtType,
			UserName:     userName,
			Name:         name,
			OSCode:       osCode,
			Status:       "creating",
			Architecture: architecture,
			QAEnabled:    qaEnabled,
		}
	}
	logger.Debugf("Creating Migration %+v", Migration)
	err = db.Create(Migration).Error
	if err != nil {
		logger.Error("DB create Migration failed, %v", err)
	}
	prefix := strings.Split(Migration.UUID, "-")[0]
	control := "inter="
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_Migration.sh '%d' '%s' '%s'", Migration.ID, prefix, url)
	if instID > 0 {
		bootVolumeUUID := ""
		if instance.Volumes != nil {
			for _, volume := range instance.Volumes {
				if volume.Booting {
					bootVolumeUUID = volume.GetOriginVolumeID()
					break
				}
			}
		}
		control = fmt.Sprintf("inter=%d", instance.Hyper)
		command = fmt.Sprintf("/opt/cloudland/scripts/backend/capture_Migration.sh '%d' '%s' '%d' '%s'", Migration.ID, prefix, instance.ID, bootVolumeUUID)
	}
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Create Migration command execution failed", err)
		return
	}
	return
}

func (a *MigrationAdmin) GetMigrationByUUID(ctx context.Context, uuID string) (Migration *model.Migration, err error) {
	db := DB()
	Migration = &model.Migration{}
	err = db.Where("uuid = ?", uuID).Take(Migration).Error
	if err != nil {
		logger.Error("Failed to query Migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get Migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *MigrationAdmin) GetMigrationByName(ctx context.Context, name string) (Migration *model.Migration, err error) {
	db := DB()
	Migration = &model.Migration{}
	err = db.Where("name = ?", name).Take(Migration).Error
	if err != nil {
		logger.Error("Failed to query Migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get Migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *MigrationAdmin) Get(ctx context.Context, id int64) (Migration *model.Migration, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid Migration ID: %d", id)
		logger.Error(err)
		return
	}
	db := DB()
	Migration = &model.Migration{Model: model.Model{ID: id}}
	err = db.Take(Migration).Error
	if err != nil {
		logger.Error("DB failed to query Migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get Migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *MigrationAdmin) GetMigration(ctx context.Context, reference *BaseReference) (Migration *model.Migration, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Migration base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		Migration, err = a.GetMigrationByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		Migration, err = a.GetMigrationByName(ctx, reference.Name)
		return
	}
	return
}

func (a *MigrationAdmin) Delete(ctx context.Context, Migration *model.Migration) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, Migration.Owner)
	if !permit {
		logger.Error("Not authorized to delete Migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	refCount := 0
	err = db.Model(&model.Instance{}).Where("Migration_id = ?", Migration.ID).Count(&refCount).Error
	if err != nil {
		logger.Error("Failed to count the number of instances using the Migration", err)
		return
	}
	if refCount > 0 {
		logger.Error("Migration can not be deleted if there are instances using it")
		err = fmt.Errorf("The Migration can not be deleted if there are instances using it")
		return
	}
	if Migration.Status == "available" {
		prefix := strings.Split(Migration.UUID, "-")[0]
		control := "inter="
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_Migration.sh '%d' '%s' '%s'", Migration.ID, prefix, Migration.Format)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Clear Migration command execution failed", err)
			return
		}
	}
	if err = db.Delete(Migration).Error; err != nil {
		return
	}
	return
}

func (a *MigrationAdmin) List(offset, limit int64, order, query string) (total int64, Migrations []*model.Migration, err error) {
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}
	Migrations = []*model.Migration{}
	if err = db.Model(&model.Migration{}).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(query).Find(&Migrations).Error; err != nil {
		return
	}

	return
}

func (v *MigrationView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, Migrations, err := MigrationAdmin.List(offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusInternalServerError)
		return
	}
	pages := GetPages(total, limit)
	c.Data["Migrations"] = Migrations
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.HTML(200, "Migrations")
}

func (v *MigrationView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	MigrationID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	Migration, err := MigrationAdmin.Get(ctx, int64(MigrationID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = MigrationAdmin.Delete(ctx, Migration)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "Migrations",
	})
	return
}

func (v *MigrationView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, instances, err := instanceAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusInternalServerError)
		return
	}
	c.Data["Instances"] = instances
	c.HTML(200, "Migrations_new")
}

func (v *MigrationView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../Migrations"
	osCode := c.QueryTrim("osCode")
	name := c.QueryTrim("name")
	url := c.QueryTrim("url")
	instance := c.QueryInt64("instance")
	osVersion := c.QueryTrim("osVersion")
	virtType := "kvm-x86_64"
	userName := c.QueryTrim("userName")
	qaEnabled := true
	architecture := "x86_64"
	_, err := MigrationAdmin.Create(c.Req.Context(), osCode, name, osVersion, virtType, userName, url, architecture, qaEnabled, instance)
	if err != nil {
		logger.Error("Create Migration failed", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}
