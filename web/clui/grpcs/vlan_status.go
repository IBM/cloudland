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
	Add("vlan_status", VlanStatus)
}

func VlanStatus(ctx context.Context, job *model.Job, args []string) (status string, err error) {
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
		vlanStatus := strings.Split(statusList[i], ":") // for each vlanStatus: [0: vlanNo, 1: STATUS, 2: FIRST, 3:SECOND]
		if len(vlanStatus) != 4 {                       // four elements as above now
			continue
		}
		vlan, err := strconv.Atoi(vlanStatus[0])
		if err != nil {
			log.Println("Invalid vlan ID", err)
			continue
		}
		netlink := &model.Network{}
		err = db.Where("vlan = ?", vlan).Take(netlink).Error
		if (err != nil && gorm.IsRecordNotFoundError(err)) ||
			(err == nil && netlink.Hyper > 0 && netlink.Hyper != int32(hyperID) && netlink.Peer > 0 && netlink.Peer != int32(hyperID)) {
			log.Println("Invalid vlan", err)
		}
		if err == nil {
			if netlink.Hyper == -1 {
				netlink.Hyper = int32(hyperID)
			} else if netlink.Peer == -1 {
				netlink.Peer = int32(hyperID)
			}
			if vlanStatus[1] == "FIRST" {
				netlink.Hyper = int32(hyperID)
			}
			if vlanStatus[2] == "SECOND" {
				netlink.Peer = int32(hyperID)
			}
			err = db.Save(netlink).Error
			if err != nil {
				log.Println("Failed to update dhcp hyper", err)
			}
		}
	}
	return
}
