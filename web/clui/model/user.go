/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type User struct {
	Model
	Username string `gorm:"size:255;unique_index" json:"username,omitempty"`
	Password string `gorm:"size:255" json:"password,omitempty"`
}

func init() {
	dbs.AutoMigrate(&User{})
}
