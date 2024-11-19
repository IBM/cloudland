/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Flavor struct {
	Model
	Owner     int64  `gorm:"default:1"` /* The organization ID of the resource */
	Name      string `gorm:"type:varchar(128)"`
	Cpu       int32
	Memory    int32
	Disk      int32
	Swap      int32
	Ephemeral int32
}

func init() {
	dbs.AutoMigrate(&Flavor{})
}
