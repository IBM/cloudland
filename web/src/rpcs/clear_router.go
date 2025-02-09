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
	Add("clear_router", ClearRouter)
}

func ClearRouter(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| clear_router.sh 5 277 MASTER
	db := DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		logger.Error("Invalid args", err)
		return
	}
	routerID, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error("Invalid gateway ID", err)
		return
	}
	router := &model.Router{Model: model.Model{ID: int64(routerID)}}
	err = db.Preload("Interfaces").Preload("Interfaces.Address").Preload("Interfaces.Address.Subnet").Where(router).Take(router).Error
	if err != nil {
		logger.Error("Invalid gateway ID", err)
		return
	}
	return
}
