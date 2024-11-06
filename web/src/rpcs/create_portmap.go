/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcs

import (
	"context"
	"fmt"
	"log"

	"web/src/model"
	"web/src/dbs"
)

func init() {
	Add("create_portmap", CreatePortmap)
}

func CreatePortmap(ctx context.Context, args []string) (status string, err error) {
	//|:-COMMAND-:| create_portmap.sh 1.2.3.4 18010
	db := dbs.DB()
	argn := len(args)
	if argn < 3 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	err = db.Model(&model.Portmap{}).Where("remote_port = ?", args[2]).Updates(map[string]interface{}{"remote_address": args[1], "status": "ready"}).Error
	if err != nil {
		log.Println("Update hyper/Peer ID failed", err)
		return
	}
	return
}
