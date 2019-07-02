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
	Add("detach_volume", DetachVolume)
}

func DetachVolume(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| detach_volume.sh 5 7
	db := dbs.DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	_, err = strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}
	volID, err := strconv.Atoi(args[2])
	if err != nil {
		log.Println("Invalid volume ID", err)
		return
	}
	volume := &model.Volume{Model: model.Model{ID: int64(volID)}}
	err = db.Where(volume).Take(volume).Error
	if err != nil {
		log.Println("Failed to query volume", err)
		return
	}
	volume.InstanceID = 0
	volume.Target = ""
	volume.Status = "available"
	err = db.Save(volume).Error
	if err != nil {
		log.Println("Update volume status failed", err)
		return
	}
	return
}
