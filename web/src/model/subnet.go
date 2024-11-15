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
	Owner        int64  `gorm:"default:1"` /* The organization ID of the resource */
	Name         string `gorm:"unique_index:idx_router_subnet;type:varchar(32)"`
	Network      string `gorm:"type:varchar(64)"`
	Netmask      string `gorm:"type:varchar(64)"`
	Gateway      string `gorm:"type:varchar(64)"`
	Start        string `gorm:"type:varchar(64)"`
	End          string `gorm:"type:varchar(64)"`
	NameServer   string `gorm:"type:varchar(64)"`
	DomainSearch string `gorm:"type:varchar(256)"`
	Dhcp         bool   `gorm:"default:false"`
	Vlan         int64
	Type         string  `gorm:"type:varchar(20);default:'internal'"`
	RouterID     int64   `gorm:"unique_index:idx_router_subnet"`
	Router       *Router `gorm:"foreignkey:RouterID"`
	Routes       string  `gorm:"type:varchar(256)"`
}

type Address struct {
	Model
	Owner        int64  `gorm:"default:1"` /* The organization ID of the resource */
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
