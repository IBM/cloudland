/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
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

func (a *MigrationAdmin) Create(ctx context.Context, name string, instances []*model.Instance, force bool, tgtHyper int32) (migrations []*model.Migration, err error) {
	logger.Debugf("Start migrating instances to %d", tgtHyper)
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized for this operation")
		return
	}
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	if tgtHyper > -1 {
		targetHyper := &model.Hyper{Hostid: tgtHyper}
		err = db.Where(targetHyper).Take(targetHyper).Error
		if err != nil {
			logger.Error("Failed to query hyper", err)
			return
		}
		if targetHyper.Status != 1 {
			err = fmt.Errorf("Target hypvervisor is in wrong state")
			logger.Error("Target hypvervisor is in wrong state")
			return
		}
	}
	for _, instance := range instances {
		if instance.Status == "migrating" {
			continue
		}
		sourceHyper := &model.Hyper{Hostid: instance.Hyper}
		err = db.Where(sourceHyper).Take(sourceHyper).Error
		if err != nil {
			logger.Error("Failed to query hyper", err)
			return
		}
		status := "in_progress"
		migrationType := "cold"
		if sourceHyper.Status == 1 && !force {
			migrationType = "warm"
		}
		if instance.Hyper == tgtHyper {
			logger.Error("No need to migrate if source and target hypervisors are the same")
			continue
		}
		task1 := &model.Task{
			Name:    "Prepare_Target",
			Summary: "Prepare resources on target hypervisor",
			Status:  status,
		}
		migration := &model.Migration{
			Model:       model.Model{Creater: memberShip.UserID},
			Name:        name,
			InstanceID:  instance.ID,
			Type:        migrationType,
			Force:       force,
			SourceHyper: instance.Hyper,
			TargetHyper: tgtHyper,
			Phases:      []*model.Task{task1},
			Status:      status,
		}
		migration.Instance = instance
		logger.Debugf("Creating migration %+v", migration)
		err = db.Create(migration).Error
		if err != nil {
			logger.Error("DB create migration failed, %v", err)
			return
		}
		var metadata string
		metadata, err = instanceAdmin.GetMetadata(ctx, instance, "")
		if err != nil {
			logger.Error("Failed to get metadata")
			return
		}
		control := fmt.Sprintf("inter=%d", tgtHyper)
		if tgtHyper == -1 {
			var hyperGroup string
			hyperGroup, err = instanceAdmin.GetHyperGroup(ctx, instance.ZoneID, instance.Hyper)
			if err != nil {
				task1.Summary = "No qualified target"
				task1.Status = "not_doing"
				migration.Status = "not_doing"
				err = db.Model(migration).Save(migration).Error
				if err != nil {
					logger.Error("Failed to update save migration, %v", err)
					return
				}
				continue
			}
			control = "select=" + hyperGroup
		}
		err = db.Model(instance).Update("status", "migrating").Error
		if err != nil {
			logger.Error("Failed to update instance status to migrating, %v", err)
			return
		}
		flavor := instance.Flavor
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/target_migration.sh '%d' '%d' '%d' '%s' '%d' '%d' '%d' '%s' '%s'<<EOF\n%s\nEOF", migration.ID, task1.ID, instance.ID, instance.Hostname, flavor.Cpu, flavor.Memory, flavor.Disk, sourceHyper.Hostname, migrationType, base64.StdEncoding.EncodeToString([]byte(metadata)))
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Target migration command execution failed", err)
			return
		}
		migrations = append(migrations, migration)
	}
	return
}

func (a *MigrationAdmin) GetMigrationByUUID(ctx context.Context, uuID string) (migration *model.Migration, err error) {
	db := DB()
	migration = &model.Migration{}
	err = db.Preload("Instance").Preload("Phases").Where("uuid = ?", uuID).Take(migration).Error
	if err != nil {
		logger.Error("Failed to query migration, %v", err)
		return
	}
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Admin)
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
	permit := memberShip.CheckPermission(model.Admin)
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
	permit := memberShip.CheckPermission(model.Admin)
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
	if err = db.Preload("Instance").Preload("Phases").Where(query).Find(&migrations).Error; err != nil {
		return
	}

	return
}

func (v *MigrationView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
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
	hypers := []*model.Hyper{}
	err = DB().Where("hostid >= 0").Find(&hypers).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Instances"] = instances
	c.Data["Hypers"] = hypers
	c.HTML(200, "migrations_new")
}

func (v *MigrationView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "../migrations"
	name := c.QueryTrim("name")
	instList := c.QueryTrim("instances")
	var instances []*model.Instance
	instArray := strings.Split(instList, ",")
	for _, inst := range instArray {
		instID, err := strconv.Atoi(inst)
		if err != nil {
			logger.Error("Invalid instance ID", err)
			continue
		}
		var instance *model.Instance
		instance, err = instanceAdmin.Get(ctx, int64(instID))
		if err != nil {
			logger.Error("Failed to get instance", err)
			c.Data["ErrorMsg"] = "Failed to get instance"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		instances = append(instances, instance)
	}
	tgthyper := c.QueryInt("hyper")
	forceStr := c.QueryTrim("force")
	force := false
	if forceStr == "yes" {
		force = true
	}
	_, err := migrationAdmin.Create(ctx, name, instances, force, int32(tgthyper))
	if err != nil {
		logger.Error("Create migration failed", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Redirect(redirectTo)
}
