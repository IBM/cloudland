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
	Add("router_status", RouterStatus)
}

func RouterStatus(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| vlan_status.sh '127' '12345 54321'
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
	for i := 0; i < len(statusList); i++ {
		ID, err := strconv.Atoi(statusList[i])
		if err != nil {
			log.Println("Invalid instance ID", err)
			continue
		}
		gateway := &model.Gateway{Model: model.Model{ID: int64(ID)}}
		err = db.Take(gateway).Error
		if (err != nil && gorm.IsRecordNotFoundError(err)) ||
			(err == nil && gateway.Hyper > 0 && gateway.Hyper != int32(hyperID) && gateway.Peer > 0 && gateway.Peer != int32(hyperID)) {
			log.Println("Invalid router", err)
		}
		if err == nil {
			if gateway.Hyper == -1 {
				gateway.Hyper = int32(hyperID)
			} else if gateway.Peer == -1 {
				gateway.Peer = int32(hyperID)
			}
			err = db.Save(gateway).Error
			if err != nil {
				log.Println("Failed to update router hyper", err)
			}
		}
	}
	return
}
