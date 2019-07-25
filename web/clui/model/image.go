/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Image struct {
	Model
	Name         string `gorm:"type:varchar(128)"`
	OSCode       string `gorm:"type:varchar(128)"`
	Format       string `gorm:"type:varchar(128)"`
	Architecture string `gorm:"type:varchar(256)"`
	Status       string `gorm:"type:varchar(128)"`
	Href         string `gorm:"type:varchar(256)"`
	Checksum     string `gorm:"type:varchar(36)"`
	OsHashAlgo   string `gorm:"type:varchar(36)"`
	OsHashValue  string `gorm:"type:varchar(36)"`
	Holder       string `gorm:"type:varchar(36)"`
	Protected    bool
	Visibility   string `gorm:"type:varchar(36)"`
	MiniDisk     int32
	MiniMem      int32
	Size         int64
}

func init() {
	dbs.AutoMigrate(&Image{})
}
