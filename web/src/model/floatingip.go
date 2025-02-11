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
	UUID       string `gorm:"type:char(36);uniqueIndex"`
	Owner      int64  `gorm:"default:1"`                     /* The organization ID of the resource */
	Name       string `gorm:"type:varchar(64)",gorm:"index"` /* The name of the resource */
	FipAddress string `gorm:"type:varchar(64)"`
	IntAddress string `gorm:"type:varchar(64)"`
	InstanceID int64
	Instance   *Instance  `gorm:"foreignkey:InstanceID",gorm:"PRELOAD:false"`
	Interface  *Interface `gorm:"foreignkey:FloatingIp"`
	RouterID   int64
	Inbound    int32
	Outbound   int32
	Router     *Router `gorm:"foreignkey:RouterID"`
	IPAddress  string
}

func init() {
	dbs.AutoMigrate(&FloatingIp{})
}
