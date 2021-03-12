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
	Add("create_router", CreateRouter)
}

func CreateRouter(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| create_router.sh 5 277 MASTER
	db := dbs.DB()
	argn := len(args)
	if argn < 5 {
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
	err = db.Preload("Subnets").Where(gateway).Take(gateway).Error
	if err != nil {
		log.Println("Invalid gateway ID", err)
		return
	}
	devIfaces := []*model.Interface{}
	for _, subnet := range gateway.Subnets {
		iface := &model.Interface{}
		err = db.Where("subnet = ? and device = ? and type='gateway'", subnet.ID, gateway.ID).Preload("Address").Preload("Address.Subnet").Take(iface).Error
		if err != nil {
			log.Println("Failed to query interface", err)
			return
		}
		devIfaces = append(devIfaces, iface)
	}
	hyperID := -1
	hyperID, err = strconv.Atoi(args[2])
	if err != nil {
		log.Println("Invalid hyper ID", err)
		return
	}
	if args[3] == "MASTER" {
		err = db.Model(&gateway).Updates(map[string]interface{}{"hyper": int32(hyperID), "status": "active"}).Error
		if err != nil {
			log.Println("Update hyper ID failed", err)
			return
		}
		if args[4] == "yes" {
			iface := &model.Interface{}
			err = db.Model(&iface).Where("device = ? and type = 'gateway'", gwID).Updates(map[string]interface{}{"hyper": int32(hyperID)}).Error
			if err != nil {
				log.Println("Failed to send fdb rules", err)
				return
			}
			err = sendFdbRules(ctx, devIfaces, int32(hyperID), "/opt/cloudland/scripts/backend/add_fwrule.sh")
			if err != nil {
				log.Println("Failed to send fdb rules", err)
				return
			}
		}
	} else if args[3] == "SLAVE" {
		err = db.Model(&gateway).Updates(map[string]interface{}{"peer": int32(hyperID)}).Error
		if err != nil {
			log.Println("Update peer ID failed", err)
			return
		}
	}
	return
}
