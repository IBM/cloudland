/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package grpcs

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("launch_vm", LaunchVM)
}

func LaunchVM(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| launch_vm.sh '127' 'running' '3' 'reason'
	db := dbs.DB()
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	reason := ""
	errHndl := ctx.Value("error")
	if errHndl != nil {
		reason = "Resource is not enough"
		err = db.Model(instance).Updates(map[string]interface{}{
			"status": "error",
			"reason": reason}).Error
		if err != nil {
			log.Println("Failed to update instance", err)
		}
		return
	}
	err = db.Where(instance).Take(instance).Error
	if err != nil {
		log.Println("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	serverStatus := args[2]
	hyperID := -1
	if serverStatus == "running" {
		hyperID, err = strconv.Atoi(args[3])
		if err != nil {
			log.Println("Invalid hyper ID", err)
			reason = err.Error()
			return
		}
	} else if argn >= 4 {
		reason = args[4]
	}
	err = db.Model(&instance).Updates(map[string]interface{}{
		"status": serverStatus,
		"hyper":  int32(hyperID),
		"reason": reason}).Error
	if err != nil {
		log.Println("Failed to update instance", err)
		return
	}
	err = db.Model(&model.Interface{}).Where("instance = ?", instance.ID).Update(map[string]interface{}{"hyper": int32(hyperID)}).Error
	if err != nil {
		log.Println("Failed to update interface", err)
		return
	}
	return
}
