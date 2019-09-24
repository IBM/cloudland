/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"time"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type Vnc struct {
	Model
	LocalAddress  string `gorm:"type:varchar(64)"`
	LocalPort     int32
	AccessAddress string `gorm:"type:varchar(64)"`
	AccessPort    int32
	Passwd        string `gorm:"type:varchar(32)"`
	InstanceID    int64
	Router        int64
	ExpiredAt     *time.Time
}

func init() {
	dbs.AutoMigrate(&Vnc{})
}
