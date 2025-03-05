/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	. "web/src/common"
	"web/src/model"
)

func init() {
	Add("migrate", TargetMigrate)
}

func Migrate(ctx context.Context, args []string) (status string, err error) {
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
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	err = db.Take(instance).Error
	if err != nil {
		logger.Error("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	err = db.Preload("Address").Preload("Address.Subnet").Preload("Address.Subnet.Router").Where("instance = ?", instID).Find(&instance.Interfaces).Error
	if err != nil {
		logger.Error("Failed to get interfaces", err)
		reason = err.Error()
		return
	}
	status := args[5]
	if status == "completed" {
		instance.Hyper = int32(hyperID)
		err = db.Model(&instance).Updates(map[string]interface{}{
			"status": running,
			"hyper":  int32(hyperID)).Error
			if err != nil {
				logger.Error("Failed to update instance", err)
				return
			}
		}
	} else if status == "target_prepared" {
	} else if status == "source_prepared" {
	}
	migration := &model.Instance{Model: model.Model{ID: int64(migrationID)}}
	err = db.Model(migration).Update(map[string]interface{}{"status": status).Error
	if err != nil {
		logger.Error("Failed to update migration", err)
		return
	}
	task := &model.Instance{Model: model.Model{ID: int64(taskID)}}
	err = db.Model(task).Update(map[string]interface{}{"status": status).Error
	if err != nil {
		logger.Error("Failed to update task", err)
		return
	}
	return
}
