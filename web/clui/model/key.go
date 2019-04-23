/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Key struct {
	Model
	Name        string `gorm:"type:varchar(100)"`
	PublicKey   string `gorm:"type:varchar(8192)"`
	Length      int32
	FingerPrint string `gorm:"type:varchar(100)"`
}

func init() {
	dbs.AutoMigrate(&Key{})
}
