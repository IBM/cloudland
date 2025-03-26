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
	Add("migrate_vm", MigrateVM)
}

type VolumeInfo struct {
	ID      int64  `json:"id"`
	UUID    string `json:"uuid"`
	Device  string `json:"device"`
	Booting bool   `json:"booting"`
}

func execSourceMigrate(ctx context.Context, instance *model.Instance, migration *model.Migration, taskID int64, migrationType string) (err error) {
	db := DB()
	targetHyper := &model.Hyper{}
	err = db.Where("hostid = ?", migration.TargetHyper).Take(targetHyper).Error
	if err != nil {
		logger.Error("Failed to query target hyper", err)
		return
	}
	sourceHyper := &model.Hyper{}
	err = db.Where("hostid = ?", migration.SourceHyper).Take(sourceHyper).Error
	if err != nil {
		logger.Error("Failed to query source hyper", err)
		return
	}
	volumes := []*VolumeInfo{}
	for _, volume := range instance.Volumes {
		volumes = append(volumes, &VolumeInfo{
			ID:      volume.ID,
			UUID:    volume.GetOriginVolumeID(),
			Device:  volume.Target,
			Booting: volume.Booting,
		})
	}
	volumesJson, err := json.Marshal(volumes)
	if err != nil {
		logger.Error("Failed to marshal instance json data", err)
		return
	}
	if sourceHyper.Status == 1 {
		control := fmt.Sprintf("inter=%d", migration.SourceHyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/source_migration.sh '%d' '%d' '%d' '%d' '%s' '%s'<<EOF\n%s\nEOF", migration.ID, taskID, instance.ID, instance.RouterID, targetHyper.Hostname, migration.Type, volumesJson)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Source migration command execution failed", err)
			return
		}
	} else {
		err = fmt.Errorf("Source hyper is not in a valid state")
	}
	return
}

func MigrateVM(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| migrate_vm.sh '12' '2' '127' '3' 'state'
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
	status = args[5]
	taskStatus := status
	migration := &model.Migration{Model: model.Model{ID: int64(migrationID)}}
	err = db.Model(migration).Take(migration).Error
	if err != nil {
		logger.Error("Failed to get migration record", err)
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	err = db.Preload("Volumes").Take(instance).Error
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}
	if status == "completed" {
		err = db.Model(instance).Updates(map[string]interface{}{"status": "running"}).Error
		if err != nil {
			logger.Error("Failed to update instance status to running, %v", err)
			return
		}
		_, err = LaunchVM(ctx, []string{args[0], args[3], "running", args[4], "sync"})
		if err != nil {
			logger.Error("Failed to sync vm info", err)
			return
		}
	} else if status == "failed" {
		err = db.Model(instance).Updates(map[string]interface{}{"status": "unknown"}).Error
		if err != nil {
			logger.Error("Failed to update instance status to unknown, %v", err)
			return
		}
	} else if status == "target_prepared" {
		migration.TargetHyper = int32(hyperID)
		targetHyper := &model.Hyper{}
		err = db.Where("hostid = ?", hyperID).Take(targetHyper).Error
		if err != nil {
			logger.Error("Failed to query hyper", err)
			return
		}
		task2 := &model.Task{
			Name:    "Prepare_Source",
			Mission: migration.ID,
			Summary: "Prepare resources on source hypervisor",
			Status:  "in_progress",
		}
		err = db.Model(task2).Create(task2).Error
		if err != nil {
			logger.Error("Failed to create task2", err)
			return
		}
		err = execSourceMigrate(ctx, instance, migration, task2.ID, migration.Type)
		if err != nil {
			logger.Error("Failed to exec source migration", err)
			status = "failed"
		}
		taskStatus = "completed"
	} else if status == "source_prepared" {
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
		taskStatus = "completed"
	}
	if migration.Status != "completed" && migration.Status != "failed" {
		migration.Status = status
	}
	err = db.Model(migration).Save(migration).Error
	if err != nil {
		logger.Error("Failed to update migration", err)
		return
	}
	err = db.Model(&model.Task{}).Where("id = ?", taskID).Update(map[string]interface{}{"status": taskStatus}).Error
	if err != nil {
		logger.Error("Failed to update task", err)
		return
	}
	return
}
