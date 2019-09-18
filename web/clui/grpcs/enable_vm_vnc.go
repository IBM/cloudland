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
	Add("enable_vm_vnc", EnableVMVNC)
}

func EnableVMVNC(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| enable_vm_vnc.sh 6 192.168.71.110 18000 password
	db := dbs.DB()
	argn := len(args)
	if argn < 5 {
		err = fmt.Errorf("Wrong params")
		log.Println("Invalid args", err)
		return
	}
	instID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}
	/*vnc := &model.Vnc{InstanceID: int64(instID)}
	err = db.Where(vnc).Take(vnc).Error
	if err != nil {
		log.Println("Invalid instance ID", err)
		return
	}*/
	portN, err := strconv.Atoi(args[3])
	if err != nil {
		log.Println("Invalid port number", err)
		return
	}

	vnc := &model.Vnc{}
	expireAt := time.Now().Add(time.Minute * 30)
	err = db.Where(model.Vnc{InstanceID: int64(instID)}).Assign(model.Vnc{Address: args[2], Port: int32(portN), Passwd: args[4], ExpiredAt: &expireAt}).FirstOrCreate(vnc).Error
	if err != nil {
		log.Println("Failed to update vnc", err)
		return
	}
	return
}
