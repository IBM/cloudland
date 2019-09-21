package grpcs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("enable_vnc_portmap", EnableVncPortmap)
}

func EnableVncPortmap(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| enable_vm_vnc.sh 6 192.168.71.110 18000
	db := dbs.DB()
	argn := len(args)
	if argn < 2 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}
	portmap := args[2]

	expireAt := time.Now().Add(time.Minute * 30)
	vnc := &model.Vnc{
		Portmap:   portmap,
		ExpiredAt: &expireAt,
	}
	err = db.Model(vnc).Where("instance_id = ?", int64(instID)).Update(vnc).Error
	if err != nil {
		log.Println("Failed to update vnc", err)
		return
	}
	return
}
