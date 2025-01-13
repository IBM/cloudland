/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"math/rand"
	"time"

	"web/src/dbs"
)

type Interface struct {
	Model
	Owner          int64  `gorm:"default:1"` /* The organization ID of the resource */
	Name           string `gorm:"type:varchar(32)"`
	MacAddr        string `gorm:"type:varchar(32)"`
	Instance       int64
	Device         int64
	Dhcp           int64
	FloatingIp     int64
	Subnet         int64
	RouterID       int64
	Address        *Address `gorm:"foreignkey:Interface"`
	Hyper          int32    `gorm:"default:-1"`
	PrimaryIf      bool     `gorm:"default:false"`
	Type           string   `gorm:"type:varchar(20)"`
	Mtu            int32
	Inbound        int32
	Outbound       int32
	AllowSpoofing  bool
	SecurityGroups []*SecurityGroup `gorm:"many2many:secgroup_ifaces;"`
}

func init() {
	dbs.AutoMigrate(&Interface{})
	rand.Seed(time.Now().UnixNano())
}
