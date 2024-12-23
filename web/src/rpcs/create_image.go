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
	Add("create_image", CreateImage)
}

func CreateImage(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_image.sh '5' 'available' 'qcow2'
	db := DB()
	argn := len(args)
	if argn < 5 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	imgID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid gateway ID", err)
		return
	}
	image := &model.Image{Model: model.Model{ID: int64(imgID)}}
	err = db.Take(image).Error
	if err != nil {
		log.Println("Invalid image ID", err)
		return
	}
	image.Status = args[2]
	image.Format = args[3]
	imageSize, err := strconv.Atoi(args[4])
	if err != nil {
		log.Println("Invalid image size", err)
		return
	}
	image.Size = int64(imageSize)
	err = db.Save(image).Error
	if err != nil {
		log.Println("Update image failed", err)
		return
	}
	return
}
