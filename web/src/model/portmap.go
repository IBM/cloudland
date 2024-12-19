/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Portmap struct {
	Model
	Owner         int64  `gorm:"default:1"` /* The organization ID of the resource */
	Name          string `gorm:"type:varchar(64)"`
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
