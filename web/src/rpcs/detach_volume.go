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
	Add("detach_volume_local", DetachVolume)
	Add("detach_volume_wds_vhost", DetachVolume)
}

func DetachVolume(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| detach_volume.sh_local 5 7
	//|:-COMMAND-:| detach_volume.sh_wds_vhost 5 7
	db := DB()
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
