/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Network struct {
	Model
	Name     string `gorm:"type:varchar(100)"`
	Hyper    int32  `gorm:"default:-1"`
	Peer     int32  `gorm:"default:-1"`
	Vlan     int64
	Type     string
	External bool      // external router
	Subnets  []*Subnet `gorm:"foreignkey:Vlan;AssociationForeignKey:Vlan;PRELOAD:false"`
}

func init() {
	dbs.AutoMigrate(&Network{})
}
