/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"fmt"
	"strconv"

	. "web/src/common"
	"web/src/model"
)

func init() {
	Add("system_router", SystemRouter)
}

func SystemRouter(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_router.sh '7' '2' 'MASTER' 'yes'
	db := DB()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		logger.Debug("Invalid args", err)
		return
	}
	hyperID, err := strconv.Atoi(args[1])
	if err != nil || hyperID < 0 {
		logger.Debug("Invalid hypervisor ID", err)
		return
	}
	hyperName := args[2]
	hyper := &model.Hyper{}
	err = db.Where("hostid = ?", hyperID).Take(hyper).Error
	if err != nil {
		logger.Debug("Failed to query hypervisor", err)
		return
	}
	if hyper.Hostname != hyperName {
		logger.Debug("Hypervisor hostname mismatch", err)
		return
	}
	subnets := []*model.Subnet{}
	err = db.Where("type = 'public'").Find(&subnets).Error
	if err != nil {
		logger.Debug("Failed to get public subnet", err)
		return
	}
	var sysIface *model.Interface
	if hyper.RouteIP == "" {
		for _, subnet := range subnets {
			sysIface, err = CreateInterface(ctx, subnet, 0, 0, int32(hyperID), "", "", hyperName, "system", nil)
			if err == nil {
				break
			}
			logger.Debugf("Failed to create system router interface for hypervisor %d from subnet %d, %v", hyperID, subnet.ID, err)
		}
		hyper.RouteIP = sysIface.Address.Address
		err = db.Save(hyper).Error
		if err != nil {
			logger.Debug("Failed to save hyper address", err)
			return
		}
	} else {
		address := &model.Address{}
		err = db.Preload("Subnet").Where("address = ?", hyper.RouteIP).Take(address).Error
		if err != nil {
			logger.Debug("Failed to get hyper address", err)
			return
		}
		if address.Allocated {
			sysIface = &model.Interface{Address: address}
		} else {
			sysIface, err = CreateInterface(ctx, address.Subnet, 0, 0, int32(hyperID), hyper.RouteIP, "", hyperName, "system", nil)
			if err != nil {
				logger.Debugf("Failed to create interface with address %s, %v", hyper.RouteIP, err)
				return
			}
		}
	}
	if sysIface == nil {
		logger.Debugf("Failed to allocate public ip for system router of hypervisor %d", hyperID)
		return
	}
	subnet := sysIface.Address.Subnet
	control := fmt.Sprintf("inter=%d", hyperID)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/system_router.sh '%d' '%s' '%s'", subnet.Vlan, sysIface.Address.Address, subnet.Gateway)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Debug("Add_fwrule execution failed", err)
		return
	}
	return
}
