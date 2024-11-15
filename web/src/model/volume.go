/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Volume struct {
	Model
	Owner      int64  `gorm:"default:1"` /* The organization ID of the resource */
	Name       string `gorm:"type:varchar(128)"`
	Path       string `gorm:"type:varchar(128)"`
	Size       int32
	Format     string `gorm:"type:varchar(32)"`
	Status     string `gorm:"type:varchar(32)"`
	Target     string `gorm:"type:varchar(32)"`
	Href       string `gorm:"type:varchar(256)"`
	InstanceID int64
	Instance   *Instance `gorm:"foreignkey:InstanceID"`
}

func init() {
	dbs.AutoMigrate(&Volume{})
}
