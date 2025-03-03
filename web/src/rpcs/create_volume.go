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
	Add("create_volume_local", CreateVolumeLocal)
	Add("create_volume_wds_vhost", CreateVolumeWDSVhost)
}

func updateInstance(volume *model.Volume, status string, reason string) (err error) {
	db := DB()
	if volume.Booting && status == "error" {
		instance := &model.Instance{Model: model.Model{ID: volume.InstanceID}}
		if err = db.Take(&instance).Error; err != nil {
			logger.Error("Invalid instance ID", err)
			return err
		}

		instance.Status = status
		instance.Reason = reason
		if err = db.Save(&instance).Error; err != nil {
			logger.Error("Update instance status failed", err)
			return err
		}
	}
	return
}

func CreateVolumeLocal(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_volume.sh 5 /volume-12.disk available reason
	logger.Debug("CreateVolumeLocal", args)
	db := DB()
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	volID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid volume ID", err)
		return
	}
	volume := &model.Volume{Model: model.Model{ID: int64(volID)}}
	err = db.Where(volume).Take(volume).Error
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}
	path := args[2]
	status = args[3]
	err = db.Model(&volume).Updates(map[string]interface{}{"path": path, "status": status}).Error
	if err != nil {
		logger.Error("Update volume status failed", err)
		return
	}
	if err = updateInstance(volume, status, args[4]); err != nil {
		logger.Error("Update instance status failed", err)
		return
	}
	return
}

func CreateVolumeWDSVhost(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_volume_wds_vhost.sh 5 available wds_vhost://1/2 reason
	logger.Debug("CreateVolumeWDSVhost", args)
	db := DB()
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	volID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid volume ID", err)
		return
	}
	volume := &model.Volume{Model: model.Model{ID: int64(volID)}}
	err = db.Where(volume).Take(volume).Error
	if err != nil {
		logger.Error("Invalid volume ID", err)
		return
	}
	status = args[2]
	path := args[3]
	err = db.Model(&volume).Updates(map[string]interface{}{"path": path, "status": status}).Error
	if err != nil {
		logger.Error("Update volume status failed", err)
		return
	}
	if err = updateInstance(volume, status, args[4]); err != nil {
		logger.Error("Update instance status failed", err)
		return
	}
	return
}
