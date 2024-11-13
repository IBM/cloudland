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
	Name       string `gorm:"type:varchar(32)"`
	Status     string `gorm:"type:varchar(32)"`
	Type       string `gorm:"type:varchar(32)"`
	Hyper      int32  `gorm:"default:-1"`
	Peer       int32  `gorm:"default:-1"`
	DefaultSG  int64
	Interfaces []*Interface `gorm:"foreignkey:Device"`
	Subnets    []*Subnet    `gorm:"foreignkey:RouterID"`
	PublicID   int64
}

func init() {
	dbs.AutoMigrate(&Router{})
}
