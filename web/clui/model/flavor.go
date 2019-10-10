/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Flavor struct {
	Model
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
