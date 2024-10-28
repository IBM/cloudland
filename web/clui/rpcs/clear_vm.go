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

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("clear_vm", ClearVM)
}

func deleteInterfaces(ctx context.Context, instance *model.Instance) (err error) {
	for _, iface := range instance.Interfaces {
		err = deleteInterface(ctx, iface)
		if err != nil {
			log.Println("Failed to delete interface", err)
			continue
		}
	}
	return
}

func deleteInterface(ctx context.Context, iface *model.Interface) (err error) {
	db := dbs.DB()
        if err = db.Model(&model.Address{}).Where("interface = ?", iface.ID).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error; err != nil {
                log.Println("Failed to Update addresses, %v", err)
                return
        }
        err = db.Delete(iface).Error
        if err != nil {
                log.Println("Failed to delete interface", err)
                return
        }
	vlan := iface.Address.Subnet.Vlan
	netlink := iface.Address.Subnet.Netlink
	if netlink == nil {
		log.Println("Subnet doesn't have network")
		return
	}
	control := ""
	if netlink.Hyper >= 0 {
		control = fmt.Sprintf("inter=%d", netlink.Hyper)
		if netlink.Peer >= 0 && netlink.Hyper != netlink.Peer {
			control = fmt.Sprintf("toall=vlan-%d:%d,%d", iface.Address.Subnet.Vlan, netlink.Hyper, netlink.Peer)
		}
	} else if netlink.Peer >= 0 {
		control = fmt.Sprintf("inter=%d", netlink.Peer)
	} else {
		log.Println("Network has no valid hypers")
		return
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/del_host.sh '%d' '%s' '%s'", vlan, iface.MacAddr, iface.Address.Address)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Execute deleting interface failed")
		return
	}
	return
}

func ClearVM(ctx context.Context, args []string) (status string, err error) {
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
	err = db.Take(instance).Error
	if err != nil {
		log.Println("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	err = db.Preload("Address").Preload("Address.Subnet").Where("instance = ?", instID).Find(&instance.Interfaces).Error
	if err != nil {
		log.Println("Failed to get interfaces", err)
		reason = err.Error()
		return
	}
	err = db.Model(&instance).Updates(map[string]interface{}{
		"status": "deleted",
		"reason": reason}).Error
	if err != nil {
		return
	}
	err = sendFdbRules(ctx, instance, "/opt/cloudland/scripts/backend/del_fwrule.sh")
	if err != nil {
		log.Println("Failed to send clear fdb rules", err)
		return
	}
	err = deleteInterfaces(ctx, instance)
	if err != nil {
		log.Println("Failed to delete interfaces", err)
		return
	}
	if err = db.Delete(instance).Error; err != nil {
		log.Println("Failed to delete instance, %v", err)
		return
	}
	return
}
