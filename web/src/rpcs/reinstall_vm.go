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
	Add("reinstall_vm", ReinstallVM)
}

func ReinstallVM(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| reinstall_vm.sh '127' 'error' '3' '10' 'reason'
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()

	argn := len(args)
	if argn < 5 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}

	instID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	err = db.Take(instance).Error
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}

	serverStatus := args[2]

	hyperID, err := strconv.Atoi(args[3])
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

	volumeID, err := strconv.Atoi(args[4])
	if err != nil {
		logger.Error("Invalid volume ID", err)
		return
	}
	volume := &model.Volume{Model: model.Model{ID: int64(volumeID)}}
	err = db.Take(volume).Error
	if err != nil {
		logger.Error("Invalid volume ID", err)
		return
	}

	reason := args[5]
	if serverStatus == "error" {
		err = db.Model(&instance).Updates(map[string]interface{}{
			"status": serverStatus,
			"hyper":  int32(hyperID),
			"reason": reason,
		}).Error
		if err != nil {
			logger.Error("Failed to update instance", err)
			return
		}

		err = db.Model(&volume).Updates(map[string]interface{}{
			"status": serverStatus,
		}).Error
		if err != nil {
			logger.Error("Failed to update volume", err)
			return
		}
	}
	return
}
