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

func CreateVolumeLocal(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_volume.sh 5 /volume-12.disk available
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

	if volume.Booting && status == "error" {
		reason := path
		instanceId := volume.InstanceID
		instance := &model.Instance{Model: model.Model{ID: instanceId}}
		err = db.Take(&instance).Error
		if err != nil {
			logger.Error("Invalid instance ID", err)
			return
		}

		instance.Status = status
		instance.Reason = reason
		err = db.Save(&instance).Error
		if err != nil {
			logger.Error("Update instance status failed", err)
			return
		}
	}
	return
}

func CreateVolumeWDSVhost(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_volume_wds_vhost.sh 5 available wds_vhost://1/2
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

	if volume.Booting && status == "error" {
		reason := path
		instanceId := volume.InstanceID
		instance := &model.Instance{Model: model.Model{ID: instanceId}}
		err = db.Take(&instance).Error
		if err != nil {
			logger.Error("Invalid instance ID", err)
			return
		}

		instance.Status = status
		instance.Reason = reason
		err = db.Save(&instance).Error
		if err != nil {
			logger.Error("Update instance status failed", err)
			return
		}
	}
	return
}
