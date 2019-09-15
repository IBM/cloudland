/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"math/rand"
	"time"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type Interface struct {
	Model
	Name       string `gorm:"type:varchar(32)"`
	MacAddr    string `gorm:"type:varchar(32)"`
	Instance   int64
	Device     int64
	Dhcp       int64
	FloatingIp int64
	Subnet     int64
	Address    *Address `gorm:"foreignkey:Interface"`
	Hyper      int32    `gorm:"default:-1"`
	PrimaryIf  bool     `gorm:"default:false"`
	Type       string   `gorm:"type:varchar(20)"`
	Mtu        int32
	Secgroups  []*SecurityGroup `gorm:"many2many:secgroup_ifaces;"`
	AddrPairs  string           `gorm:"type:varchar(256)"`
}

func init() {
	dbs.AutoMigrate(&Interface{})
	rand.Seed(time.Now().UnixNano())
}
