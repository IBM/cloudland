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
	//|:-COMMAND-:| hyper_status.sh '127' 'hyper-0' '0/64' '26684376/263662552' '1561731870272/3086445260864'
	db := dbs.DB()
	argn := len(args)
	if argn < 6 {
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
	hyper.Cpu = args[3]
	hyper.Memory = args[4]
	hyper.Disk = args[5]
	err = db.Save(hyper).Error
	if err != nil {
		log.Println("Failed to save hypervisor", err)
		return
	}
	return
}
