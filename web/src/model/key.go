/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Key struct {
	Model
	Owner       int64  `gorm:"unique_index:idx_account_key;default:1"` /* The organization ID of the resource */
	Name        string `gorm:"unique_index:idx_account_key;type:varchar(100)"`
	PublicKey   string `gorm:"type:varchar(8192)"`
	Length      int32
	FingerPrint string `gorm:"type:varchar(100)"`
}

func init() {
	dbs.AutoMigrate(&Key{})
}
