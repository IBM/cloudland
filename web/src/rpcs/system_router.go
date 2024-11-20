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

	"web/src/model"
	"web/src/dbs"
	"web/src/common"
)

func init() {
	Add("system_router", SystemRouter)
}

func SystemRouter(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_router.sh '7' '2' 'MASTER' 'yes'
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
	hyperName := args[2]
	hyper := &model.Hyper{}
	err = db.Where("hostid = ?", hyperID).Take(hyper).Error
	if err != nil {
		log.Println("Failed to query hypervisor", err)
		return
	}
	if hyper.Hostname != hyperName {
		log.Println("Hypervisor hostname mismatch", err)
		return
	}
	subnet := &model.Subnet{}
	err = db.Where("type = 'public'").Take(&subnet).Error
	if err != nil {
		log.Println("Failed to get public subnet", err)
		return
	}
	sysIface, err := common.CreateInterface(ctx, subnet.ID, 0, 0, int32(hyperID), "", "", hyperName, "system", nil)
	if err != nil {
		log.Printf("Failed to create system router interface for hypervisor %d, %v", hyperID, err)
		return
	}
	control := fmt.Sprintf("inter=%d", hyperID)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/system_router.sh '%d' '%s' '%s'", subnet.Vlan, sysIface.Address.Address, subnet.Gateway)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Add_fwrule execution failed", err)
		return
	}
	return
}
