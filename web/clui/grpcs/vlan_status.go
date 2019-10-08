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
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/jinzhu/gorm"
)

func init() {
	Add("vlan_status", VlanStatus)
}

func VlanStatus(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| vlan_status.sh '127' '12345 54321'
	db := dbs.DB()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	hyperID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid hypervisor ID", err)
		return
	}
	statusList := strings.Split(args[2], " ")
	for i := 0; i < len(statusList); i++ {
		vlan, err := strconv.Atoi(statusList[i])
		if err != nil {
			log.Println("Invalid instance ID", err)
			continue
		}
		netlink := &model.Network{}
		err = db.Where("vlan = ?", vlan).Take(netlink).Error
		if (err != nil && gorm.IsRecordNotFoundError(err)) ||
			(err == nil && netlink.Hyper > 0 && netlink.Hyper != int32(hyperID) && netlink.Peer > 0 && netlink.Peer != int32(hyperID)) {
			log.Println("Invalid vlan", err)
			sciClient := RemoteExecClient()
			control := fmt.Sprintf("inter=%d", hyperID)
			command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_nspace.sh 'vlan%d'", vlan)
			sciReq := &scripts.ExecuteRequest{
				Id:      100,
				Extra:   0,
				Control: control,
				Command: command,
			}
			_, err = sciClient.Execute(ctx, sciReq)
			if err != nil {
				log.Println("SCI client execution failed", err)
				continue
			}
		}
	}
	return
}
