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
	Add("pre_migrate", PreMigrateVM)
}

type FdbRule struct {
	Instance string `json:"instance"`
	Vni      int64  `json:"vni"`
	InnerIP  string `json:"inner_ip"`
	InnerMac string `json:"inner_mac"`
	OuterIP  string `json:"outer_ip"`
	Gateway  string `json:"gateway"`
	Router   int64  `json:"router"`
}

func PreMigrateVM(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| pre_migrate_vm.sh '127' 'migrate_prepared' '3'
	db := DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid instance ID", err)
		return
	}
	instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
	err = db.Take(instance).Error
	if err != nil {
		logger.Error("Invalid instance ID", err)
		reason = err.Error()
		return
	}
	err = db.Preload("Address").Preload("Address.Subnet").Preload("Address.Subnet.Router").Where("instance = ?", instID).Find(&instance.Interfaces).Error
	if err != nil {
		logger.Error("Failed to get interfaces", err)
		reason = err.Error()
		return
	}
	hyperID, err := strconv.Atoi(args[3])
	if err != nil {
		logger.Error("Invalid hyper ID", err)
		reason = err.Error()
		return
	}
	instance.Hyper = int32(hyperID)
	return
}

func syncNicInfo(ctx context.Context, instance *model.Instance) (err error) {
	vlans := []*VlanInfo{}
	var securityData []*SecurityData
	db := DB()
	for _, iface := range instance.Interfaces {
		err = db.Model(iface).Related(&iface.SecurityGroups, "SecurityGroups").Error
		if err != nil {
			logger.Error("Get security groups for interface failed", err)
			return
		}
		securityData, err = GetSecurityData(ctx, iface.SecurityGroups)
		if err != nil {
			logger.Error("Get security data for interface failed", err)
			return
		}
		subnet := iface.Address.Subnet
		vlans = append(vlans, &VlanInfo{Device: iface.Name, Vlan: subnet.Vlan, Inbound: iface.Inbound, Outbound: iface.Outbound, AllowSpoofing: iface.AllowSpoofing, Gateway: subnet.Gateway, Router: subnet.RouterID, IpAddr: iface.Address.Address, MacAddr: iface.MacAddr, SecRules: securityData})
	}
	jsonData, err := json.Marshal(vlans)
	if err != nil {
		logger.Error("Failed to marshal instance json data", err)
		return
	}
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/sync_nic_info.sh '%d' '%s' <<EOF\n%s\nEOF", instance.ID, instance.Hostname, jsonData)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Execute floating ip failed", err)
		return
	}
	return
}

func syncFloatingIp(ctx context.Context, instance *model.Instance) (err error) {
	db := DB()
	var primaryIface *model.Interface
	for i, iface := range instance.Interfaces {
		if iface.PrimaryIf {
			primaryIface = instance.Interfaces[i]
			break
		}
	}
	if primaryIface != nil {
		floatingIp := &model.FloatingIp{}
		err = db.Preload("Interface").Preload("Interface.Address").Preload("Interface.Address.Subnet").Where("instance_id = ?", instance.ID).Take(floatingIp).Error
		if err != nil {
			logger.Error("Failed to get floating ip", err)
			return
		}
		pubSubnet := floatingIp.Interface.Address.Subnet
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_floating.sh '%d' '%s' '%s' '%d' '%s' '%d'", floatingIp.RouterID, floatingIp.FipAddress, pubSubnet.Gateway, pubSubnet.Vlan, primaryIface.Address.Address, primaryIface.Address.Subnet.Vlan)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			logger.Error("Execute floating ip failed", err)
			return
		}
	}
	return
}
