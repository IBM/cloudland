/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package grpcs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("launch_vm", LaunchVM)
	Add("oc_vm", LaunchVM)
}

type FdbRule struct {
	Instance int64  `json:"instance"`
	Vni      int64  `json:"vni"`
	InnerIP  string `json:"inner_ip"`
	InnerMac string `json:"inner_mac"`
	OuterIP  string `json:"outer_ip"`
}

func sendFdbRules(ctx context.Context, instance *model.Instance) (err error) {
	db := dbs.DB()
	allSubnets := []*model.Subnet{}
	instanceRules := []*FdbRule{}
	spreadRules := []*FdbRule{}
	for _, iface := range instance.Interfaces {
		allSubnets = append(allSubnets, iface.Address.Subnet)
		hyper := &model.Hyper{}
		err = db.Where("hostid = ?", instance.Hyper).Take(hyper).Error
		if err != nil || hyper.Hostid < 0 {
			log.Println("Failed to query hypervisor")
			continue
		}
		spreadRules = append(spreadRules, &FdbRule{Instance: instance.ID, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper.HostIP})
	}
	allIfaces := []*model.Interface{}
	hyperSet := make(map[int32]struct{})
	for _, subnet := range allSubnets {
		err := db.Preload("Addresses").Preload("Addresses.Subnet").Where("subnet = ?", subnet.ID).Find(allIfaces).Error
		if err != nil {
			log.Println("Failed to query all interfaces")
			continue
		}
		for _, iface := range allIfaces {
			if iface.Hyper >= 0 {
				hyper := &model.Hyper{}
				err = db.Where("hostid = ? and hostid != ?", iface.Hyper, instance.Hyper).Take(hyper).Error
				if err != nil {
					log.Println("Failed to query hypervisor")
					continue
				}
				hyperSet[iface.Hyper] = struct{}{}
				instanceRules = append(instanceRules, &FdbRule{Instance: iface.Instance, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper.HostIP})
			} else if iface.Device > 0 {
				gateway := &model.Gateway{Model: model.Model{ID: iface.Device}}
				if err != nil {
					log.Println("Failed to query gateway")
					continue
				}
				hyperSet[gateway.Hyper] = struct{}{}
				hyperSet[gateway.Peer] = struct{}{}
				hyper1 := &model.Hyper{}
				err = db.Where("hostid = ? and hostid != ?", gateway.Hyper, instance.Hyper).Take(hyper1).Error
				if err != nil {
					log.Println("Failed to query hypervisor")
					continue
				}
				instanceRules = append(instanceRules, &FdbRule{Instance: iface.Instance, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper1.HostIP})
				hyper2 := &model.Hyper{}
				err = db.Where("hostid = ? and hostid != ?", gateway.Hyper, instance.Hyper).Take(hyper2).Error
				if err != nil {
					log.Println("Failed to query hypervisor")
					continue
				}
				instanceRules = append(instanceRules, &FdbRule{Instance: iface.Instance, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper2.HostIP})
			}
		}
	}
	if len(hyperSet) > 0 {
		hyperList := fmt.Sprintf("group-fdb-%d", instance.ID)
		i := 0
		for key := range hyperSet {
			if i == 0 {
				hyperList = fmt.Sprintf("%s:%d", hyperList, key)
			} else {
				hyperList = fmt.Sprintf("%s,%d", hyperList, key)
			}
			i++
		}
		fdbJson, _ := json.Marshal(instanceRules)
		control := "toall=" + hyperList
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/add_fwrule.sh <<EOF\n%s\nEOF", fdbJson)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Add_fwrule execution failed", err)
			return
		}
	}
	if len(instanceRules) > 0 {
		fdbJson, _ := json.Marshal(instanceRules)
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/add_fwrule.sh <<EOF\n%s\nEOF", fdbJson)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Add_fwrule execution failed", err)
			return
		}
	}
	return
}

func LaunchVM(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| launch_vm.sh '127' 'running' '3' 'reason'
	db := dbs.DB()
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	reason := ""
	errHndl := ctx.Value("error")
	if errHndl != nil {
		reason = "Resource is not enough"
		err = db.Model(instance).Updates(map[string]interface{}{
			"status": "error",
			"reason": reason}).Error
		if err != nil {
			log.Println("Failed to update instance", err)
		}
		return
	}
	err = db.Preload("Interfaces").Where(instance).Take(instance).Error
	if err != nil {
		log.Println("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	serverStatus := args[2]
	hyperID := -1
	if serverStatus == "running" {
		hyperID, err = strconv.Atoi(args[3])
		if err != nil {
			log.Println("Invalid hyper ID", err)
			reason = err.Error()
			return
		}
	} else if argn >= 4 {
		reason = args[4]
	}
	instance.Hyper = int32(hyperID)
	err = db.Model(&instance).Updates(map[string]interface{}{
		"status": serverStatus,
		"hyper":  int32(hyperID),
		"reason": reason}).Error
	if err != nil {
		log.Println("Failed to update instance", err)
		return
	}
	err = db.Model(&model.Interface{}).Where("instance = ?", instance.ID).Update(map[string]interface{}{"hyper": int32(hyperID)}).Error
	if err != nil {
		log.Println("Failed to update interface", err)
		return
	}
	err = sendFdbRules(ctx, instance)
	if err != nil {
		log.Println("Failed to update interface", err)
		return
	}
	return
}
