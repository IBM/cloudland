/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type User struct {
	Model
	Username string    `gorm:"size:255;unique_index" json:"username,omitempty"`
	Password string    `gorm:"size:255" json:"password,omitempty"`
	Members  []*Member `gorm:"foreignkey:UserID"`
}

func (User) TableName() string {
	return "users"
}
func init() {
	dbs.AutoMigrate(&User{})
}
