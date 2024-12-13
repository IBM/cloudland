/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"fmt"
	"log"
	"strconv"

	. "web/src/common"
	"web/src/model"
)

func init() {
	Add("attach_volume_local", AttachVolume)
	Add("attach_volume_wds_vhost", AttachVolume)
}

func AttachVolume(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| attach_volume.sh 5 7 vdb
	db := DB()
	argn := len(args)
	if argn < 4 {
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
	target := args[3]
	volume := &model.Volume{Model: model.Model{ID: int64(volID)}}
	err = db.Where(volume).Take(volume).Error
	if err != nil {
		log.Println("Failed to query volume", err)
		return
	}
	err = db.Model(&volume).Updates(map[string]interface{}{"target": target, "status": "attached"}).Error
	if err != nil {
		log.Println("Update volume status failed", err)
		return
	}
	return
}
