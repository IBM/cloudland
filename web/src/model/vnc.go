/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"time"

	"web/src/dbs"
)

type Vnc struct {
	Model
	Owner         int64  `gorm:"default:1"` /* The organization ID of the resource */
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
