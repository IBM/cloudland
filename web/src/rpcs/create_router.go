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

	"github.com/IBM/cloudland/web/src/model"
	"github.com/IBM/cloudland/web/src/dbs"
)

func init() {
	Add("create_router", CreateRouter)
}

func getHyperGroup(zoneID int64, ignoreID int32) (hyperGroup string, err error) {
	db := dbs.DB()
	hypers := []*model.Hyper{}
	where := fmt.Sprintf("zone_id = %d and status = 1", zoneID)
	if err = db.Where(where).Find(&hypers).Error; err != nil {
		log.Println("Hypers query failed", err)
		return
	}
	if len(hypers) == 0 {
		log.Println("No qualified hypervisor")
		return
	}
	for i, h := range hypers {
		if h.Hostid == ignoreID {
			continue
		}
		if i == 0 {
			hyperGroup = fmt.Sprintf("group-zone-%d:%d", zoneID, h.Hostid)
		} else {
			hyperGroup = fmt.Sprintf("%s,%d", hyperGroup, h.Hostid)
		}
	}
	return
}

func CreateRouter(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_router.sh '7' '2' 'MASTER' 'yes'
	db := dbs.DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	routerID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid router ID", err)
		return
	}
	hyperID, err := strconv.Atoi(args[2])
	if err != nil {
		log.Println("Invalid hypervisor ID", err)
		return
	}
	role := args[3]
	router := &model.Router{Model: model.Model{ID: int64(routerID)}}
	err = db.Take(router).Error
	if err != nil {
		log.Println("Failed to get router", err)
		return
	}
	if role == "MASTER" {
		router.Hyper = int32(hyperID)
		hyperGroup, _ := getHyperGroup(router.ZoneID, router.Hyper)
		if hyperGroup != "" {
			pubSubnet := &model.Subnet{Model: model.Model{ID: int64(router.PublicID)}}
			err = db.Take(pubSubnet).Error
			if err != nil {
				log.Println("Failed to get public subnet", err)
				return
			}
			pubIface := &model.Interface{}
			err = db.Where("subnet = ? and Device = ?", pubSubnet.ID, router.ID).Preload("address").Find(&pubIface).Error
			if err != nil {
				log.Println("Failed to get public interface", err)
				return
			}
			control := "select=" + hyperGroup
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_router.sh '%d' '%s' '%s' '%d' '%s' 'SLAVE'", router.ID, pubSubnet.Gateway, pubIface.Address.Address, router.VrrpVni, router.PeerAddr)
			err = HyperExecute(ctx, control, command)
			if err != nil {
				log.Println("Create peer router command execution failed, %v", err)
				return
			}
		}
	} else if role == "SLAVE" {
		router.Peer = int32(hyperID)
	}
	err = db.Save(router).Error
	if err != nil {
		log.Println("Failed to update router", err)
	}
	return
}
