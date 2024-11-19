/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Authorize struct {
	Model
	Owner        int64  `gorm:"default:1"` /* The organization ID of the resource */
	ResourceType string `gorm:"type:varchar(32)"`
	ResourceID   uint
	User         uint
	Project      uint
}

func init() {
	dbs.AutoMigrate(&Authorize{})
}
