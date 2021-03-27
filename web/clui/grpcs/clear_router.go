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
	Add("clear_router", ClearRouter)
}

func ClearRouter(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| clear_router.sh 5 277 MASTER
	db := dbs.DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	gwID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid gateway ID", err)
		return
	}
	gateway := &model.Gateway{Model: model.Model{ID: int64(gwID)}}
	err = db.Preload("Interfaces").Preload("Interfaces.Address").Preload("Interfaces.Address.Subnet").Where(gateway).Take(gateway).Error
	if err != nil {
		log.Println("Invalid gateway ID", err)
		return
	}
	hyperID := -1
	hyperID, err = strconv.Atoi(args[2])
	if err != nil {
		log.Println("Invalid hyper ID", err)
		return
	}
	err = sendFdbRules(ctx, gateway.Interfaces, int32(hyperID), "/opt/cloudland/scripts/backend/del_fwrule.sh")
	if err != nil {
		log.Println("Failed to send fdb rules", err)
		return
	}
	return
}
