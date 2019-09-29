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
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/jinzhu/gorm"
)

func init() {
	Add("inst_status", InstanceStatus)
}

func InstanceStatus(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| launch_vm.sh '3' '5 running 7 running 9 shut_off'
	db := dbs.DB()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	hyperID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid hypervisor ID", err)
		return
	}
	statusList := strings.Split(args[2], " ")
	for i := 0; i < len(statusList); i += 2 {
		instID, err := strconv.Atoi(statusList[i])
		if err != nil {
			log.Println("Invalid instance ID", err)
			continue
		}
		status := statusList[i+1]
		instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
		err = db.Unscoped().Take(instance).Error
		if err != nil {
			log.Println("Invalid instance ID", err)
			if gorm.IsRecordNotFoundError(err) {
				instance.Hostname = "unknown"
				instance.Status = status
				instance.Hyper = int32(hyperID)
				err = db.Create(instance).Error
				if err != nil {
					log.Println("Failed to create unknown instance", err)
				}
			}
			continue
		}
		if instance.Status != status || instance.Hyper != int32(hyperID) {
			instance.Status = status
			instance.Hyper = int32(hyperID)
			instance.DeletedAt = nil
			err = db.Unscoped().Save(instance).Error
			if err != nil {
				log.Println("Failed to update status", err)
				continue
			}
		}
	}
	return
}
