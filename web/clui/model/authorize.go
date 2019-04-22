/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Authorize struct {
	Model
	ResourceType string `gorm:"type:varchar(32)"`
	ResourceID   uint
	User         uint
	Project      uint
}

func init() {
	dbs.AutoMigrate(&Authorize{})
}
