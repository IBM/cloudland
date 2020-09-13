/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package grpcs

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("hyper_status", HyperStatus)
}

func HyperStatus(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| hyper_status.sh '127' 'hyper-0' '0' '64' '26684376' '263662552' '1561731870272' '3086445260864' '1'
	db := dbs.DB()
	argn := len(args)
	if argn < 10 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	hyperID, err := strconv.Atoi(args[1])
	if err != nil || hyperID < 0 {
		log.Println("Invalid hypervisor ID", err)
		return
	}
	hyper := &model.Hyper{}
	err = db.Where("hostid = ?", hyperID).Take(hyper).Error
	if err != nil {
		log.Println("Failed to query hypervisor", err)
		return
	}
	hyper.Hostname = args[2]
	availCpu, err := strconv.Atoi(args[3])
	if err != nil {
		log.Println("Invalid available cpu", err)
		availCpu = 0
	}
	totalCpu, err := strconv.Atoi(args[4])
	if err != nil {
		log.Println("Invalid total cpu", err)
		totalCpu = 0
	}
	availMem, err := strconv.Atoi(args[5])
	if err != nil {
		log.Println("Invalid available memory", err)
		availMem = 0
	}
	totalMem, err := strconv.Atoi(args[6])
	if err != nil {
		log.Println("Invalid total memory", err)
		totalMem = 0
	}
	availDisk, err := strconv.Atoi(args[7])
	if err != nil {
		log.Println("Invalid available disk", err)
		availDisk = 0
	}
	totalDisk, err := strconv.Atoi(args[8])
	if err != nil {
		log.Println("Invalid total disk", err)
		totalDisk = 0
	}
	hyperStatus, err := strconv.Atoi(args[9])
	if err != nil {
		log.Println("Invalid hypervisor status", err)
		hyperStatus = 1
	}
	err = db.Where("hostid = ?", hyperID).FirstOrCreate(&model.Resource{Hostid: int32(hyperID)}).Error
	if err != nil {
		log.Println("Failed to create resource", err)
		return
	}
	err = db.Model(&model.Resource{}).Where("hostid = ?", hyperID).Update(map[string]interface{}{
		"cpu":          int64(availCpu),
		"cpu_total":    int64(totalCpu),
		"memory":       int64(availMem),
		"memory_total": int64(totalMem),
		"disk":         int64(availDisk),
		"disk_total":   int64(totalDisk),
	}).Error
	if err != nil {
		log.Println("Failed to save resource", err)
		return
	}
	hyper.Status = int32(hyperStatus)
	err = db.Save(hyper).Error
	if err != nil {
		log.Println("Failed to save hypervisor", err)
		return
	}
	return
}
