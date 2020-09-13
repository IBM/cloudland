/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Console struct {
	Model
	Instance   int64
	HashSecret string `gorm:"type:varchar(256)"`
	Type       string
}

func init() {
	dbs.AutoMigrate(&Console{})
}
