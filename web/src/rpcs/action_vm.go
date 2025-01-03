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
	Add("action_vm", ActionVM)
}

func ActionVM(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| clear_vm.sh '127'
	db := DB()
	argn := len(args)
	if argn < 2 {
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
	status = args[2]
	err = db.Model(&instance).Updates(map[string]interface{}{
		"status": status,
	}).Error
	if err != nil {
		return
	}
	return
}
