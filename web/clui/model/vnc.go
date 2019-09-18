/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"time"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type Vnc struct {
	Model
	Address    string `gorm:"type:varchar(15)"`
	Port       int32
	Passwd     string `gorm:"type:varchar(32)"`
	InstanceID int64
	ExpiredAt  *time.Time
}

func init() {
	dbs.AutoMigrate(&Vnc{})
}
