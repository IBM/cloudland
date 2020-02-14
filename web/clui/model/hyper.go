/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package model

import (
	context "context"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/cloudland/web/clui/hypers"
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Hyper struct {
	ID        int64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Hostid    int32  `gorm:"unique_index"`
	Hostname  string `gorm:"type:varchar(64)"`
	Status    int32
	Parentid  int32
	Children  int32
	Duration  int64
	Resource  *Resource `gorm:"foreignkey:Hostid;AssociationForeignKey:Hostid`
}

func init() {
	dbs.AutoMigrate(&Hyper{})
}

const (
	HYPER_INIT    = ""
	HYPER_CREATED = "created"
	HYPER_ACTIVE  = "active"
)

var (
	HyperStatusValues = map[int32]string{
		0: HYPER_CREATED,
		1: HYPER_ACTIVE,
	}
	HyperStatusNames = map[string]int32{
		HYPER_INIT:    0,
		HYPER_CREATED: 0,
		HYPER_ACTIVE:  1,
	}
)

func (hyper *Hyper) LoadRequest(h *hypers.Hyper) {
	hyper.Hostid = h.GetId()
	hyper.Hostname = h.GetHostname()
	hyper.Parentid = h.GetParentid()
	hyper.Status = HyperStatusNames[h.GetStatus()]
	hyper.Duration = h.GetDuration()
}

func (hyper *Hyper) ToReply() (h *hypers.Hyper) {
	h = &hypers.Hyper{
		Id:       hyper.Hostid,
		Hostname: hyper.Hostname,
		Status:   HyperStatusValues[hyper.Status],
		Parentid: hyper.Parentid,
		Duration: hyper.Duration,
	}
	return
}

func (hyper *Hyper) LoadControl(control string) {
	items := strings.Split(control, " ")
	for _, item := range items {
		kv := strings.Split(item, "=")
		if len(kv) != 2 {
			continue
		}
		k, v := kv[0], kv[1]
		if v == "" {
			continue
		}
		switch k {
		case "id":
			if id, err := strconv.Atoi(v); err == nil {
				hyper.Hostid = int32(id)
			}
		case "hostname":
			hyper.Hostname = v
		case "num":
			if num, err := strconv.Atoi(v); err == nil {
				hyper.Children = int32(num)
			}
		}
	}
}

func (hyper *Hyper) LoadCommand(command string) {
	// 6845,cloudland-136,1
	items := strings.Split(command, ",")
	if len(items) != 3 {
		return
	}
	if items[1] != "" {
		hyper.Hostname = items[1]
	}
	if items[0] != "" {
		if id, err := strconv.Atoi(items[0]); err == nil {
			hyper.Hostid = int32(id)
		}
		if status, err := strconv.Atoi(items[2]); err == nil {
			hyper.Status = int32(status)
		}
	}
}

func (hyper *Hyper) Updates(ctx context.Context, values *Hyper) (err error) {
	logger, ctx := startLogging(ctx, "Updates")
	db := dbs.DB()
	where := map[string]interface{}{
		"hostid": values.Hostid,
	}
	if err = db.FirstOrCreate(hyper, where).Error; err != nil {
		logger.Error(err)
		return
	}

	if err = db.Model(hyper).Updates(values).Error; err != nil {
		logger.Error(err)
		return
	}
	return
}
