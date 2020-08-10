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
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/jinzhu/gorm"
)

func init() {
	Add("inst_status", InstanceStatus)
}

type SecurityData struct {
	Secgroup    int64
	RemoteIp    string `json:"remote_ip"`
	RemoteGroup string `json:"remote_group"`
	Direction   string `json:"direction"`
	IpVersion   string `json:"ip_version"`
	Protocol    string `json:"protocol"`
	PortMin     int32  `json:"port_min"`
	PortMax     int32  `json:"port_max"`
}

func InstanceStatus(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| launch_vm.sh '3' '5 running 7 running 9 shut_off'
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
	for i := 0; i < len(statusList); i += 2 {
		instID, err := strconv.Atoi(statusList[i])
		if err != nil {
			log.Println("Invalid instance ID", err)
			continue
		}
		status := statusList[i+1]
		instance := &model.Instance{Model: model.Model{ID: int64(instID)}}
		err = db.Unscoped().Take(instance).Error
		if err != nil {
			log.Println("Invalid instance ID", err)
			if gorm.IsRecordNotFoundError(err) {
				instance.Hostname = "unknown"
				instance.Status = status
				instance.Hyper = int32(hyperID)
				err = db.Create(instance).Error
				if err != nil {
					log.Println("Failed to create unknown instance", err)
				}
			}
			continue
		}
		if instance.Status == "migrating" && instance.Hyper == int32(hyperID) {
			continue
		}
		if instance.Status != status {
			query := fmt.Sprintf("status = '%s'", "migrating")
			err = db.Unscoped().Model(instance).Where(query).Update(map[string]interface{}{
				"status":     status,
				"deleted_at": nil,
			}).Error
			if err != nil {
				log.Println("Failed to update status", err)
			}
		}
		if instance.Hyper != int32(hyperID) {
			instance.Hyper = int32(hyperID)
			err = db.Unscoped().Model(instance).Update(map[string]interface{}{
				"hyper":      int32(hyperID),
				"deleted_at": nil,
			}).Error
			if err != nil {
				log.Println("Failed to hypervisor", err)
			}
			err = db.Unscoped().Model(&model.Interface{}).Where("instance = ?", instance.ID).Update(map[string]interface{}{
				"hyper":      int32(hyperID),
				"deleted_at": nil,
			}).Error
			if err != nil {
				log.Println("Failed to update interface", err)
				continue
			}
			err = ApplySecgroups(ctx, instance)
			if err != nil {
				log.Println("Failed to apply security groups", err)
				continue
			}
		}
	}
	return
}

func ApplySecgroups(ctx context.Context, instance *model.Instance) (err error) {
	db := dbs.DB()
	var ifaces []*model.Interface
	if err = db.Set("gorm:auto_preload", true).Where("instance = ?", instance.ID).Find(&ifaces).Error; err != nil {
		log.Println("Interfaces query failed", err)
		return
	}
	for _, iface := range ifaces {
		var secRules []*model.SecurityRule
		secRules, err = model.GetSecurityRules(iface.Secgroups)
		if err != nil {
			log.Println("Failed to get security rules", err)
			continue
		}
		securityData := []*SecurityData{}
		for _, rule := range secRules {
			sgr := &SecurityData{
				Secgroup:    rule.Secgroup,
				RemoteIp:    rule.RemoteIp,
				RemoteGroup: rule.RemoteGroup,
				Direction:   rule.Direction,
				IpVersion:   rule.IpVersion,
				Protocol:    rule.Protocol,
				PortMin:     rule.PortMin,
				PortMax:     rule.PortMax,
			}
			securityData = append(securityData, sgr)
		}
		var jsonData []byte
		jsonData, err = json.Marshal(securityData)
		if err != nil {
			log.Println("Failed to marshal security json data, %v", err)
			continue
		}
		control := fmt.Sprintf("inter=%d", instance.Hyper)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/attach_nic.sh '%d' '%d' '%s' '%s' <<EOF\n%s\nEOF", instance.ID, iface.Address.Subnet.Vlan, iface.Address.Address, iface.MacAddr, jsonData)
		err = HyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed", err)
			continue
		}
	}
	return
}
