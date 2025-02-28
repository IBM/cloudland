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
	migrationAdmin = &MigrationAdmin{}
	migrationView  = &MigrationView{}
)

type MigrationAdmin struct{}
type MigrationView struct{}

func (a *MigrationAdmin) Create(ctx context.Context, name, instance *model.Instance, force bool, toHyper *model.Hyper) (migration *model.Migration, err error) {
	logger.Debugf("Start migrating %s for instance %d from %d to %d", name, instance.ID, instance.Hyper, toHyper)
	memberShip := GetMemberShip(ctx)
	permit = memberShip.CheckPermission(model.Admin)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized for this operation")
		return
	}
	if instance.Hyper == toHyper.Hostid {
		logger.Error("No need to migrate if source and target hypervisors are the same")
		err = fmt.Errorf("No need to migrate if source and target hypervisors are the same")
		return
	}
	if toHyper.Status != 1 {
		logger.Error("Target hypervisors is not available")
		err = fmt.Errorf("Target hypervisor is not available")
		return
	}
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	hyper := &model.Hyper{Hostid: instance.Hyper)
	err = db.Where(hyper).Take(hyper).Error
	if err != nil {
		logger.Error("Failed to query hyper", err)
		return
	}
	task1 := &model.Task{
		Name: "Migration_Step1",
		Summary: "Processing resources on source hypervisor",
	}
	shutdown := false
	if force {
		if fromHyper.Status == 1 {
			task1.Status = "in_progress"
			shutdown = true
		} else {
			task1.Status = "not_doing"
		}
	}
	migration = &model.Migration{
		Model:        model.Model{Creater: memberShip.UserID},
		Name:         name,
		Force:        force,
		FromHyper:    instance.Hyper,
		ToHyper:      toHyper.Hostid,
		Phases:       []*Tasks{task1},
	}
	logger.Debugf("Creating migration %+v", migration)
	err = db.Create(migration).Error
	if err != nil {
		logger.Error("DB create migration failed, %v", err)
	}
	if shutdown {
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/action_vm.sh '%d' '%s'", instance.ID, "shutdown")
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Shutting down instance failed", err)
			return
		}
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_migration.sh '%d' '%s' '%s'", migration.ID, prefix, url)
	control = fmt.Sprintf("inter=%d", instance.Hyper)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Create migration command execution failed", err)
		return
	}
	return
}

func (a *MigrationAdmin) GetMigrationByUUID(ctx context.Context, uuID string) (migration *model.Migration, err error) {
	db := DB()
	migration = &model.Migration{}
	err = db.Where("uuid = ?", uuID).Take(migration).Error
	if err != nil {
		logger.Error("Failed to query migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *MigrationAdmin) GetMigrationByName(ctx context.Context, name string) (migration *model.Migration, err error) {
	db := DB()
	migration = &model.Migration{}
	err = db.Where("name = ?", name).Take(migration).Error
	if err != nil {
		logger.Error("Failed to query migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *MigrationAdmin) Get(ctx context.Context, id int64) (migration *model.Migration, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid migration ID: %d", id)
		logger.Error(err)
		return
	}
	db := DB()
	migration = &model.Migration{Model: model.Model{ID: id}}
	err = db.Take(migration).Error
	if err != nil {
		logger.Error("DB failed to query migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Error("Not authorized to get migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *MigrationAdmin) GetMigration(ctx context.Context, reference *BaseReference) (migration *model.Migration, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Migration base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		migration, err = a.GetMigrationByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		migration, err = a.GetMigrationByName(ctx, reference.Name)
		return
	}
	return
}

func (a *MigrationAdmin) Delete(ctx context.Context, migration *model.Migration) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, migration.Owner)
	if !permit {
		logger.Error("Not authorized to delete migration")
		err = fmt.Errorf("Not authorized")
		return
	}
	refCount := 0
	err = db.Model(&model.Instance{}).Where("migration_id = ?", migration.ID).Count(&refCount).Error
	if err != nil {
		logger.Error("Failed to count the number of instances using the migration", err)
		return
	}
	if refCount > 0 {
		logger.Error("Migration can not be deleted if there are instances using it")
		err = fmt.Errorf("The migration can not be deleted if there are instances using it")
		return
	}
	if migration.Status == "available" {
		prefix := strings.Split(migration.UUID, "-")[0]
		control := "inter="
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_migration.sh '%d' '%s' '%s'", migration.ID, prefix, migration.Format)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Clear migration command execution failed", err)
			return
		}
	}
	if err = db.Delete(migration).Error; err != nil {
		return
	}
	return
}

func (a *MigrationAdmin) List(offset, limit int64, order, query string) (total int64, migrations []*model.Migration, err error) {
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
	migrations = []*model.Migration{}
	if err = db.Model(&model.Migration{}).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(query).Find(&migrations).Error; err != nil {
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
	total, migrations, err := migrationAdmin.List(offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusInternalServerError)
		return
	}
	pages := GetPages(total, limit)
	c.Data["Migrations"] = migrations
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.HTML(200, "migrations")
}

func (v *MigrationView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is empty"
		c.Error(http.StatusBadRequest)
		return
	}
	migrationID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	migration, err := migrationAdmin.Get(ctx, int64(migrationID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = migrationAdmin.Delete(ctx, migration)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "migrations",
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
	c.HTML(200, "migrations_new")
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
	redirectTo := "../migrations"
	osCode := c.QueryTrim("osCode")
	name := c.QueryTrim("name")
	url := c.QueryTrim("url")
	instance := c.QueryInt64("instance")
	osVersion := c.QueryTrim("osVersion")
	virtType := "kvm-x86_64"
	userName := c.QueryTrim("userName")
	qaEnabled := true
	architecture := "x86_64"
	_, err := migrationAdmin.Create(c.Req.Context(), osCode, name, osVersion, virtType, userName, url, architecture, qaEnabled, instance)
	if err != nil {
		logger.Error("Create migration failed", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}
