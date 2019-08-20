/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"time"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type Resource struct {
	ID          int64 `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Hostid      int32 `gorm:"unique_index"`
	Cpu         int64
	CpuTotal    int64
	Memory      int64
	MemoryTotal int64
	Disk        int64
	DiskTotal   int64
}

func init() {
	dbs.AutoMigrate(&Resource{})
}
