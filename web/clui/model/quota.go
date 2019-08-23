/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Quota struct {
	Model
	Name      string `gorm:"type:varchar(128)"`
	Type      string `gorm:"type:varchar(32)"`
	Cpu       int32
	Memory    int32
	Disk      int32
	Subnet    int32
	PublicIp  int32
	PrivateIp int32
	Gateway   int32
	Volume    int32
	Secgroup  int32
	Secrule   int32
	Instance  int32
	Openshift int32
}

func init() {
	dbs.AutoMigrate(&Quota{})
}
