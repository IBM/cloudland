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
	Add("hyper_status", HyperStatus)
}

func HyperStatus(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| hyper_status.sh '127' 'hyper-0' '0' '64' '26684376' '263662552' '1561731870272' '3086445260864' '1'
	db := DB()
	argn := len(args)
	if argn < 11 {
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
	availCpu, err := strconv.Atoi(args[3])
	if err != nil {
		logger.Debug("Invalid available cpu", err)
		availCpu = 0
	}
	totalCpu, err := strconv.Atoi(args[4])
	if err != nil {
		logger.Debug("Invalid total cpu", err)
		totalCpu = 0
	}
	availMem, err := strconv.Atoi(args[5])
	if err != nil {
		logger.Debug("Invalid available memory", err)
		availMem = 0
	}
	totalMem, err := strconv.Atoi(args[6])
	if err != nil {
		logger.Debug("Invalid total memory", err)
		totalMem = 0
	}
	availDisk, err := strconv.Atoi(args[7])
	if err != nil {
		logger.Debug("Invalid available disk", err)
		availDisk = 0
	}
	totalDisk, err := strconv.Atoi(args[8])
	if err != nil {
		logger.Debug("Invalid total disk", err)
		totalDisk = 0
	}
	hyperStatus, err := strconv.Atoi(args[9])
	if err != nil {
		logger.Debug("Invalid hypervisor status", err)
		hyperStatus = 1
	}
	hostIP := args[10]
	zoneName := args[11]
	zone := &model.Zone{Name: zoneName}
	if zoneName != "" {
		err = db.Where("name = ?", zoneName).FirstOrCreate(zone).Error
		if err != nil {
			logger.Debug("Failed to create zone", err)
			return
		}
	}
	hyper := &model.Hyper{Hostid: int32(hyperID)}
	err = db.Where("hostid = ?", hyperID).Take(hyper).Error
	if err != nil {
		logger.Debug("Failed to take hyper", err)
		err = db.Create(hyper).Error
		if err != nil {
			logger.Debug("Failed to create hyper", err)
			return
		}
	}
	if hyper.RouteIP == "" {
		_, err = SystemRouter(ctx, []string{args[1], args[2]})
		if err != nil {
			logger.Debug("Failed to create system router", err)
		}
	}
	hyper.Hostname = hyperName
	hyper.Status = int32(hyperStatus)
	hyper.VirtType = "kvm-x86_64"
	hyper.Zone = zone
	hyper.HostIP = hostIP
	err = db.Save(hyper).Error
	if err != nil {
		logger.Debug("Failed to save hypervisor", err)
		return
	}
	resource := &model.Resource{
		Hostid:      int32(hyperID),
		Cpu:         int64(availCpu),
		CpuTotal:    int64(totalCpu),
		Memory:      int64(availMem),
		MemoryTotal: int64(totalMem),
		Disk:        int64(availDisk),
		DiskTotal:   int64(totalDisk),
	}
	err = db.Where("hostid = ?", hyperID).Assign(resource).FirstOrCreate(&model.Resource{}).Error
	if err != nil {
		logger.Debug("Failed to create or update hyper resource", err)
		return
	}
	return
}
