/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Subnet struct {
	Model
	Name         string `gorm:"type:varchar(32)"`
	Network      string `gorm:"type:varchar(64)"`
	Netmask      string `gorm:"type:varchar(64)"`
	Gateway      string `gorm:"type:varchar(64)"`
	Start        string `gorm:"type:varchar(64)"`
	End          string `gorm:"type:varchar(64)"`
	NameServer   string `gorm:"type:varchar(64)"`
	DomainSearch string `gorm:"type:varchar(256)"`
	Dhcp         bool   `gorm:"default:false"`
	Vlan         int64
	Type         string `gorm:"type:varchar(20);default:'internal'"`
	RouterID     int64
	Router       *Router `gorm:"foreignkey:RouterID"`
	Routes       string  `gorm:"type:varchar(256)"`
}

type Address struct {
	Model
	Address   string `gorm:"type:varchar(64)"`
	Netmask   string `gorm:"type:varchar(64)"`
	Type      string `gorm:"type:varchar(20);default:'native'"`
	Allocated bool   `gorm:"default:false"`
	Reserved  bool   `gorm:"default:false"`
	SubnetID  int64
	Subnet    *Subnet `gorm:"foreignkey:SubnetID"`
	Interface int64
}

func init() {
	dbs.AutoMigrate(&Subnet{})
	dbs.AutoMigrate(&Address{})
}
