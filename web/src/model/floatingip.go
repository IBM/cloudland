/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type FloatingIp struct {
	Model
	Owner      int64  `gorm:"default:1"` /* The organization ID of the resource */
	FipAddress string `gorm:"type:varchar(64)"`
	IntAddress string `gorm:"type:varchar(64)"`
	Type       string `gorm:"type:varchar(20)"`
	InstanceID int64
	Instance   *Instance  `gorm:"foreignkey:InstanceID",gorm:"PRELOAD:false"`
	Interface  *Interface `gorm:"foreignkey:FloatingIp"`
	RouterID   int64
	IPAddress  string
}

func init() {
	dbs.AutoMigrate(&FloatingIp{})
}
