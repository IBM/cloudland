/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

func init() {
	dbs.AutoMigrate(&SecurityGroup{}, &SecurityRule{})
}

type SecurityGroup struct {
	Model
	Owner      int64        `gorm:"unique_index:idx_account_secgroup;efault:1"` /* The organization ID of the resource */
	Name       string       `gorm:"unique_index:idx_account_secgroup;type:varchar(64)"`
	IsDefault  bool         `gorm:"default:false"`
	Interfaces []*Interface `gorm:"many2many:secgroup_ifaces;"`
	RouterID   int64
	Router     *Router `gorm:"foreignkey:RouterID"`
}

type SecurityRule struct {
	Model
	Owner         int64 `gorm:"default:1"` /* The organization ID of the resource */
	Secgroup      int64
	RemoteIp      string `gorm:"type:varchar(32)"`
	RemoteGroupID int64
	RemoteGroup   *SecurityGroup `gorm:"foreignkey:RemoteGroupID"`
	Direction     string         `gorm:"type:varchar(16)"`
	IpVersion     string         `gorm:"type:varchar(12)"`
	Protocol      string         `gorm:"type:varchar(20)"`
	PortMin       int32          `gorm:"default:-1"`
	PortMax       int32          `gorm:"default:-1"`
}

func init() {
	dbs.AutoMigrate(&SecurityGroup{}, &SecurityRule{})
}
