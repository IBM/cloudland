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
	Add("create_net", CreateNet)
}

func CreateNet(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| create_router.sh 5 277 MASTER
	db := dbs.DB()
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	vlan, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid vlan ID", err)
		return
	}
	netlink := &model.Network{Vlan: int64(vlan)}
	err = db.Where(netlink).Take(netlink).Error
	if err != nil {
		log.Println("DB failed to query network", err)
		return
	}
	hyperID := -1
	hyperID, err = strconv.Atoi(args[2])
	if err != nil {
		log.Println("Invalid hyper ID", err)
		return
	}
	if args[3] == "FIRST" {
		err = db.Model(&netlink).Updates(map[string]interface{}{"hyper": int32(hyperID)}).Error
	} else if args[3] == "SECOND" {
		err = db.Model(&netlink).Updates(map[string]interface{}{"peer": int32(hyperID)}).Error
	}
	if err != nil {
		log.Println("Update hyper/Peer ID failed", err)
		return
	}
	return
}
