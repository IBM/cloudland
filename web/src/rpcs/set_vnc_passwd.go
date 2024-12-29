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
	Add("set_vnc_passwd", SetVncPasswd)
}

func SetVncPasswd(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| enable_vm_vnc.sh 6 5909 password 192.168.10.100
	db := DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		logger.Debug("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Debug("Invalid instance ID", err)
		return
	}
	portN, err := strconv.Atoi(args[2])
	if err != nil {
		logger.Debug("Invalid port number", err)
		return
	}
	hyperip := args[3]
	vnc := &model.Vnc{
		InstanceID:   int64(instID),
		LocalAddress: hyperip,
		LocalPort:    int32(portN),
	}
	err = db.Where("instance_id = ?", int64(instID)).Assign(vnc).FirstOrCreate(&model.Vnc{}).Error
	if err != nil {
		logger.Debug("Failed to update vnc", err)
		return
	}
	return
}
