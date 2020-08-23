/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type FloatingIp struct {
	Model
	FipAddress string `gorm:"type:varchar(64)"`
	IntAddress string `gorm:"type:varchar(64)"`
	Type       string `gorm:"type:varchar(20)"`
	InstanceID int64
	Instance   *Instance  `gorm:"foreignkey:InstanceID"`
	Interface  *Interface `gorm:"foreignkey:FloatingIp"`
	GatewayID  int64
	Gateway    *Gateway `gorm:"foreignkey:GatewayID"`
	IPAddress  string
}

func init() {
	dbs.AutoMigrate(&FloatingIp{})
}
