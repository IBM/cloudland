/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type TaskStage struct {
	Model
}

type Migration struct {
	Model
	Name       string `gorm:"type:varchar(128)"`
	Force      bool   `gorm:"default:false"`
	FromHyper  int32
	ToHyper    int32
	Stages     []*TaskStage
	InstanceID int64
	Instance   *Instance `gorm:"foreignkey:InstanceID",gorm:"PRELOAD:false"`
}

func init() {
	dbs.AutoMigrate(&Migration{})
}
