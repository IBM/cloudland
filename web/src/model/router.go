/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Router struct {
	Model
	Owner      int64  `gorm:"unique_index:idx_account_router;default:1"` /* The organization ID of the resource */
	Name       string `gorm:"unique_index:idx_account_router;type:varchar(64)"`
	Status     string `gorm:"type:varchar(32)"`
	Hyper      int32  `gorm:"default:-1"`
	Peer       int32  `gorm:"default:-1"`
	DefaultSG  int64
	Interfaces []*Interface `gorm:"foreignkey:Device"`
	Subnets    []*Subnet    `gorm:"foreignkey:RouterID"`
}

func init() {
	dbs.AutoMigrate(&Router{})
}
