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
	Instance string `json:"instance"`
	Vni      int64  `json:"vni"`
	InnerIP  string `json:"inner_ip"`
	InnerMac string `json:"inner_mac"`
	OuterIP  string `json:"outer_ip"`
}

func sendFdbRules(ctx context.Context, devIfaces []*model.Interface, hyperNode int32, fdbScript string) (err error) {
	db := dbs.DB()
	allSubnets := []int64{}
	localRules := []*FdbRule{}
	spreadRules := []*FdbRule{}
	for _, iface := range devIfaces {
		allSubnets = append(allSubnets, iface.Address.SubnetID)
		hyper := &model.Hyper{}
		err = db.Where("hostid = ?", hyperNode).Take(hyper).Error
		if err != nil || hyper.Hostid < 0 {
			log.Println("Failed to query hypervisor")
			continue
		}
		spreadRules = append(spreadRules, &FdbRule{Instance: iface.Name, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper.HostIP})
	}
	allIfaces := []*model.Interface{}
	hyperSet := make(map[int32]struct{})
	for _, subnetID := range allSubnets {
		err := db.Preload("Address").Preload("Address.Subnet").Where("subnet = ?", subnetID).Find(&allIfaces).Error
		if err != nil {
			log.Println("Failed to query all interfaces", err)
			continue
		}
		for _, iface := range allIfaces {
			hyper := &model.Hyper{}
			err = db.Where("hostid = ? and hostid != ?", iface.Hyper, hyperNode).Take(hyper).Error
			if err != nil {
				log.Println("Failed to query hypervisor", err)
				continue
			}
			hyperSet[iface.Hyper] = struct{}{}
			localRules = append(localRules, &FdbRule{Instance: iface.Name, Vni: iface.Address.Subnet.Vlan, InnerIP: iface.Address.Address, InnerMac: iface.MacAddr, OuterIP: hyper.HostIP})
		}
	}
	if len(hyperSet) > 0 && len(spreadRules) > 0 {
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
		fdbJson, _ := json.Marshal(spreadRules)
		control := "toall=" + hyperList
		command := fmt.Sprintf("%s <<EOF\n%s\nEOF", fdbScript, fdbJson)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Add_fwrule execution failed", err)
			return
		}
	}
	if len(localRules) > 0 {
		fdbJson, _ := json.Marshal(localRules)
		control := fmt.Sprintf("inter=%d", hyperNode)
		command := fmt.Sprintf("%s <<EOF\n%s\nEOF", fdbScript, fdbJson)
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
	err = db.Preload("Interfaces").Preload("Interfaces.Address").Preload("Interfaces.Address.Subnet").Take(instance).Error
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
	err = sendFdbRules(ctx, instance.Interfaces, instance.Hyper, "/opt/cloudland/scripts/backend/add_fwrule.sh")
	if err != nil {
		log.Println("Failed to send fdb rules", err)
		return
	}
	return
}
