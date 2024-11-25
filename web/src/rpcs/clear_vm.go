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

	. "web/src/common"
	"web/src/model"
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
	db := DB()
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
	control := ""
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
	db := DB()
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
	err = db.Preload("Address").Preload("Address.Subnet").Preload("Address.Subnet").Where("instance = ?", instID).Find(&instance.Interfaces).Error
	if err != nil {
		log.Println("Failed to get interfaces", err)
		reason = err.Error()
		return
	}
	err = deleteInterfaces(ctx, instance)
	if err != nil {
		log.Println("Failed to delete interfaces", err)
		return
	}
	instance.Hostname = fmt.Sprintf("%s-%d", instance.Hostname, instance.CreatedAt.Unix())
	instance.Status = "deleted"
	instance.Reason = reason
	err = db.Save(instance).Error
	if err != nil {
		return
	}
	if err = db.Delete(instance).Error; err != nil {
		log.Println("Failed to delete instance, %v", err)
		return
	}
	err = sendFdbRules(ctx, instance, "/opt/cloudland/scripts/backend/del_fwrule.sh")
	if err != nil {
		log.Println("Failed to send clear fdb rules", err)
		return
	}
	return
}
