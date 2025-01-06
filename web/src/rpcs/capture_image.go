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
	"web/src/routes"
)

func init() {
	Add("capture_image", CaptureImage)
}

func CaptureImage(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| capture_image.sh '5' 'available' 'qcow2' 'message'
	argn := len(args)
	if argn < 4 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	imgID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid image ID", err)
		return
	}
	db := DB()
	image := &model.Image{Model: model.Model{ID: int64(imgID)}}
	err = db.Take(image).Error
	if err != nil {
		logger.Error("Invalid image ID", err)
		return
	}
	image.Status = args[2]
	if image.Status == "error" {
		err_msg := args[4]
		// log the error message and continue to save image
		logger.Errorf("Capture image failed: %s", err_msg)
	}
	vol_driver := routes.GetVolumeDriver()
	if vol_driver == "local" {
		image.Format = args[3]
		var imageSize int
		imageSize, err = strconv.Atoi(args[4])
		if err != nil {
			logger.Error("Invalid image size", err)
			return
		}
		image.Size = int64(imageSize)
	}
	err = db.Save(image).Error
	if err != nil {
		logger.Error("Update image failed", err)
		return
	}
	return
}
