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
	Add("clear_vm", ClearVM)
}

func ClearVM(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| clear_vm.sh '127'
	db := dbs.DB()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}
	reason := ""
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	err = db.Preload("Interfaces").Preload("Interfaces.Address").Preload("Interfaces.Address.Subnet").Take(instance).Error
	if err != nil {
		log.Println("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	err = db.Model(&instance).Updates(map[string]interface{}{
		"status": "deleted",
		"reason": reason}).Error
	if err != nil {
		return
	}
	err = sendFdbRules(ctx, instance.Interfaces, instance.Hyper, "/opt/cloudland/scripts/backend/del_fwrule.sh")
	if err != nil {
		log.Println("Failed to send clear fdb rules", err)
		return
	}
	return
}
