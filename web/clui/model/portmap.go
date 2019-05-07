/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Portmap struct {
	Model
	Name          string `gorm:"type:varchar(32)"`
	Status        string `gorm:"type:varchar(32)"`
	LocalPort     int32
	LocalAddress  string `gorm:"type:varchar(64)"`
	RemotePort    int32
	RemoteAddress string `gorm:"type:varchar(64)"`
	GatewayID     int64
	Gateway       *Gateway `gorm:"foreignkey:GatewayID"`
	InstanceID    int64
}

func init() {
	dbs.AutoMigrate(&Portmap{})
}
