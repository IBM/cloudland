/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/src/dbs"
)

type Portmap struct {
	Model
	Name          string `gorm:"type:varchar(32)"`
	Status        string `gorm:"type:varchar(32)"`
	LocalPort     int32
	LocalAddress  string `gorm:"type:varchar(64)"`
	RemotePort    int32
	RemoteAddress string `gorm:"type:varchar(64)"`
	RouterID      int64
	Router        *Router `gorm:"foreignkey:RouterID"`
	InstanceID    int64
}

func init() {
	dbs.AutoMigrate(&Portmap{})
}
