/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/model"

	"github.com/jinzhu/gorm"
)

func init() {
	Add("inst_status", InstanceStatus)
}

func InstanceStatus(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| launch_vm.sh '3' '5 running 7 running 9 shut_off'
	db := DB()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	hyperID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid hypervisor ID", err)
		return
	}
	hyper := &model.Hyper{Hostid: int32(hyperID)}
	err = db.Where(hyper).Take(hyper).Error
	if err != nil {
		logger.Error("Failed to query hyper", err)
		return
	}
	statusList := strings.Split(args[2], " ")
	for i := 0; i < len(statusList); i += 2 {
		instID, err := strconv.Atoi(statusList[i])
		if err != nil {
			logger.Error("Invalid instance ID", err)
			continue
		}
		status := statusList[i+1]
		instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
		err = db.Unscoped().Take(instance).Error
		if err != nil {
			logger.Error("Invalid instance ID", err)
			if gorm.IsRecordNotFoundError(err) {
				instance.Hostname = "unknown"
				instance.Status = status
				instance.Hyper = int32(hyperID)
				err = db.Create(instance).Error
				if err != nil {
					logger.Error("Failed to create unknown instance", err)
				}
			}
			continue
		}
		if instance.Status == "migrating" {
			continue
		}
		if instance.Status != status {
			err = db.Unscoped().Model(instance).Update(map[string]interface{}{
				"status": status,
			}).Error
			if err != nil {
				logger.Error("Failed to update status", err)
			}
		}
		if instance.Hyper != int32(hyperID) {
			instance.Hyper = int32(hyperID)
			err = db.Unscoped().Model(instance).Update(map[string]interface{}{
				"hyper": int32(hyperID),
			}).Error
			if err != nil {
				logger.Error("Failed to hypervisor", err)
			}
			err = db.Unscoped().Model(&model.Interface{}).Where("instance = ?", instance.ID).Update(map[string]interface{}{
				"hyper":   int32(hyperID),
				"zone_id": hyper.ZoneID,
			}).Error
			if err != nil {
				logger.Error("Failed to update interface", err)
				continue
			}
		}
	}
	return
}
