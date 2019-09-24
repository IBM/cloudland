package grpcs

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	Add("replace_vnc_passwd", ReplaceVncPasswd)
}

func ReplaceVncPasswd(ctx context.Context, job *model.Job, args []string) (status string, err error) {
	//|:-COMMAND-:| enable_vm_vnc.sh 6 5909 password 192.168.10.100
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
	portN, err := strconv.Atoi(args[2])
	if err != nil {
		log.Println("Invalid port number", err)
		return
	}
	passwd := args[3]
	hyperip := args[4]
	vnc := &model.Vnc{
		InstanceID:   int64(instID),
		LocalAddress: hyperip,
		LocalPort:    int32(portN),
		Passwd:       passwd,
	}
	err = db.Where("instance_id = ?", int64(instID)).Assign(vnc).FirstOrCreate(&model.Vnc{}).Error
	if err != nil {
		log.Println("Failed to update vnc", err)
		return
	}
	return
}
