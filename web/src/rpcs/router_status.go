/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/src/model"
	"github.com/IBM/cloudland/web/src/dbs"
	"github.com/jinzhu/gorm"
)

func init() {
	Add("router_status", RouterStatus)
}

func RouterStatus(ctx context.Context, args []string) (status string, err error) {
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
		router := &model.Router{Model: model.Model{ID: int64(ID)}}
		err = db.Take(router).Error
		if (err != nil && gorm.IsRecordNotFoundError(err)) ||
			(err == nil && router.Hyper > 0 && router.Hyper != int32(hyperID) && router.Peer > 0 && router.Peer != int32(hyperID)) {
			log.Println("Invalid router", err)
		}
		if err == nil {
			if router.Hyper == -1 {
				router.Hyper = int32(hyperID)
			} else if router.Peer == -1 {
				router.Peer = int32(hyperID)
			}
			err = db.Save(router).Error
			if err != nil {
				log.Println("Failed to update router hyper", err)
			}
		}
	}
	return
}
