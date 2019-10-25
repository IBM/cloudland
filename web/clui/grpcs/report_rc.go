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
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("report_rc", ReportRC)
}

func ReportRC(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| report_rc.sh 'cpu=12/16' 'memory=13395304/16016744' 'disk=58969763392/108580577280'
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	id := ctx.Value("hostid").(int32)
	var cpu int64
	var cpuTotal int64
	var memory int64
	var memoryTotal int64
	var disk int64
	var diskTotal int64
	for _, arg := range args[1:] {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			log.Println("Invalid key value pair", arg)
			return
		}
		key := kv[0]
		value := kv[1]
		vp := strings.Split(value, "/")
		if len(vp) != 2 {
			log.Println("Invalid format of value pair", value)
			return
		}
		if key == "cpu" {
			cpu, err = strconv.ParseInt(vp[0], 10, 64)
			cpuTotal, err = strconv.ParseInt(vp[1], 10, 64)
		} else if key == "memory" {
			memory, err = strconv.ParseInt(vp[0], 10, 64)
			memoryTotal, err = strconv.ParseInt(vp[1], 10, 64)
		} else if key == "disk" {
			disk, err = strconv.ParseInt(vp[0], 10, 64)
			diskTotal, err = strconv.ParseInt(vp[1], 10, 64)
		} else {
			log.Println("Undefined resource type")
		}
		if err != nil {
			log.Println("Failed to get value", err)
		}
	}
	db := dbs.DB()
	resource := &model.Resource{
		Hostid:      id,
		Cpu:         cpu,
		CpuTotal:    cpuTotal,
		Memory:      memory,
		MemoryTotal: memoryTotal,
		Disk:        disk,
		DiskTotal:   diskTotal,
	}
	err = db.Where("hostid = ?", id).Assign(resource).FirstOrCreate(&model.Resource{}).Error
	if err != nil {
		log.Println("Failed to create or update hyper resource", err)
		return
	}
	return
}
