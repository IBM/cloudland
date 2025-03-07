/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"fmt"
	"strconv"

	. "web/src/common"
	"web/src/model"
)

func init() {
	Add("migrate_vm", MigrateVM)
}

func MigrateVM(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| target_migrate.sh '12' '2' '127' '3' 'state'
	db := DB()
	argn := len(args)
	if argn < 5 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	migrationID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid migration ID", err)
		return
	}
	taskID, err := strconv.Atoi(args[2])
	if err != nil {
		logger.Error("Invalid task ID", err)
		return
	}
	instID, err := strconv.Atoi(args[3])
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}
	hyperID, err := strconv.Atoi(args[4])
	if err != nil {
		logger.Error("Invalid hyper ID", err)
		return
	}
	migration := &model.Migration{Model: model.Model{ID: int64(migrationID)}}
	err = db.Model(migration).Take(migration).Error
	if err != nil {
		logger.Error("Failed to get migration record", err)
		return
	}
	status = args[5]
	if status == "completed" {
		_, err = LaunchVM(ctx, []string{args[0], args[3], "running", args[4], "sync"})
		if err != nil {
			logger.Error("Failed to create system router", err)
			return
		}
	} else if status == "source_prepared" {
		instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
		err = db.Take(instance).Error
		if err != nil {
			logger.Error("Invalid instance ID", err)
			return
		}
		err = db.Preload("Address").Preload("Address.Subnet").Preload("Address.Subnet.Router").Where("instance = ?", instID).Find(&instance.Interfaces).Error
		if err != nil {
			logger.Error("Failed to get interfaces", err)
			return
		}
		var primaryIface *model.Interface
		for i, iface := range instance.Interfaces {
			if iface.PrimaryIf {
				primaryIface = instance.Interfaces[i]
				break
			}
		}
		err = db.Where("instance_id = ?", instance.ID).Find(&instance.FloatingIps).Error
		if err != nil {
			logger.Errorf("Failed to query floating ip(s), %v", err)
			return
		}
		if instance.FloatingIps != nil {
			for _, fip := range instance.FloatingIps {
				control := fmt.Sprintf("inter=%d", migration.SourceHyper)
				command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_floating.sh '%d' '%s' '%s' '%d' '%d'", fip.RouterID, fip.FipAddress, fip.IntAddress, primaryIface.Address.Subnet.Vlan, fip.ID)
				err = HyperExecute(ctx, control, command)
				if err != nil {
					logger.Error("Execute floating ip failed", err)
					return
				}
			}
		}
	}
	err = db.Model(migration).Update(map[string]interface{}{"status": status, "target_hyper": hyperID}).Error
	if err != nil {
		logger.Error("Failed to update migration", err)
		return
	}
	task := &model.Instance{Model: model.Model{ID: int64(taskID)}}
	err = db.Model(task).Update(map[string]interface{}{"status": status}).Error
	if err != nil {
		logger.Error("Failed to update task", err)
		return
	}
	return
}
