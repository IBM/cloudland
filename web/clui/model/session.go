/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Session struct {
	Key    string `gorm:"type:varchar(16);primary_key;not null"`
	Data   []byte
	Expiry int32 `gorm:"not null"`
}

func (Session) TableName() string {
	return "session"
}

func init() {
	dbs.AutoMigrate(&Session{})
}
