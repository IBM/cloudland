/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Task struct {
	Model
	Mission     int64
	Name        string `gorm:"type:varchar(64)"`
	Summary     string `gorm:"type:varchar(128)`
	Status      string `gorm:"type:varchar(32)"`
}

type Migration struct {
	Model
	Name       string `gorm:"type:varchar(64)"`
	InstanceID int64
	Instance   *Instance `gorm:"foreignkey:InstanceID"`
	Force      bool   `gorm:"default:false"`
	FromHyper  int32
	ToHyper    int32
	Phases     []*Task `gorm:"foreignkey:Mission"`
}

func init() {
	dbs.AutoMigrate(&Migration{})
}
