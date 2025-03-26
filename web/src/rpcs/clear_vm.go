/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	. "web/src/common"
	"web/src/model"
)

func init() {
	Add("clear_vm", ClearVM)
}

func deleteInterfaces(ctx context.Context, instance *model.Instance) (err error) {
	ctx, db := GetContextDB(ctx)
	hyperSet := make(map[int32]struct{})
	instances := []*model.Instance{}
	hyperNode := instance.Hyper
	hyper := &model.Hyper{}
	err = db.Where("hostid = ?", hyperNode).Take(hyper).Error
	if err != nil || hyper.Hostid < 0 {
		logger.Error("Failed to query hypervisor")
		return
	}
	if instance.RouterID > 0 {
		err = db.Where("router_id = ?", instance.RouterID).Find(&instances).Error
		if err != nil {
			logger.Error("Failed to query all instances", err)
			return
		}
		for _, inst := range instances {
			hyperSet[inst.Hyper] = struct{}{}
		}
	}
	hyperList := fmt.Sprintf("group-fdb-%d", hyperNode)
	i := 0
	for key := range hyperSet {
		if i == 0 {
			hyperList = fmt.Sprintf("%s:%d", hyperList, key)
		} else {
			hyperList = fmt.Sprintf("%s,%d", hyperList, key)
		}
		i++
	}
	for _, iface := range instance.Interfaces {
		err = db.Model(&model.Address{}).Where("interface = ?", iface.ID).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error
		if err != nil {
			logger.Error("Failed to Update addresses, %v", err)
			return
		}
		err = db.Delete(iface).Error
		if err != nil {
			logger.Error("Failed to delete interface", err)
			return
		}
		spreadRules := []*FdbRule{{Instance: iface.Name, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper.HostIP, Gateway: iface.Address.Subnet.Gateway, Router: iface.Address.Subnet.RouterID}}
		fdbJson, _ := json.Marshal(spreadRules)
		control := "toall=" + hyperList
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/del_fwrule.sh <<EOF\n%s\nEOF", fdbJson)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Execute floating ip failed", err)
			return
		}
	}
	return
}

func ClearVM(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| clear_vm.sh '127'
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}
	reason := ""
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	err = db.Take(instance).Error
	if err != nil {
		logger.Error("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	err = db.Preload("Address").Preload("Address.Subnet").Preload("Address.Subnet").Where("instance = ?", instID).Find(&instance.Interfaces).Error
	if err != nil {
		logger.Error("Failed to get interfaces", err)
		reason = err.Error()
		return
	}
	err = deleteInterfaces(ctx, instance)
	if err != nil {
		logger.Error("Failed to delete interfaces", err)
		return
	}
	instance.Hostname = fmt.Sprintf("%s-%d", instance.Hostname, instance.CreatedAt.Unix())
	instance.Status = "deleted"
	instance.Reason = reason
	instance.Interfaces = nil
	err = db.Save(instance).Error
	if err != nil {
		return
	}
	if err = db.Delete(instance).Error; err != nil {
		logger.Error("Failed to delete instance, %v", err)
		return
	}
	return
}
