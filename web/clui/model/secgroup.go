/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	dbs.AutoMigrate(&SecurityGroup{}, &SecurityRule{})
}

type SecurityGroup struct {
	Model
	Name      string `gorm:"type:varchar(32)"`
	IsDefault bool   `gorm:"default:false"`
}

type SecurityRule struct {
	Model
	Secgroup    int64
	RemoteIp    string `gorm:"type:varchar(32)"`
	RemoteGroup string `gorm:"type:varchar(36)"`
	Direction   string `gorm:"type:varchar(16)"`
	IpVersion   string `gorm:"type:varchar(12)"`
	Protocol    string `gorm:"type:varchar(20)"`
	PortMin     int32  `gorm:"default:-1"`
	PortMax     int32  `gorm:"default:-1"`
}

func init() {
	dbs.AutoMigrate(&SecurityGroup{}, &SecurityRule{})
}
