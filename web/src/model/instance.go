/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Instance struct {
	Model
	UUID        string        `gorm:"type:char(36);uniqueIndex"`
	Owner       int64         `gorm:"default:1"` /* The organization ID of the resource */
	Hostname    string        `gorm:"unique_index:idx_router_instance;type:varchar(128)"`
	Domain      string        `gorm:"type:varchar(128)"`
	Status      string        `gorm:"type:varchar(32)"`
	Reason      string        `gorm:"type:text"`
	FloatingIps []*FloatingIp `gorm:"foreignkey:InstanceID",gorm:"PRELOAD:false`
	Volumes     []*Volume     `gorm:"foreignkey:InstanceID",gorm:"PRELOAD:false"`
	Interfaces  []*Interface  `gorm:"foreignkey:Instance`
	Portmaps    []*Portmap    `gorm:"foreignkey:instanceID"`
	FlavorID    int64
	Flavor      *Flavor `gorm:"foreignkey:FlavorID"`
	ImageID     int64
	Image       *Image `gorm:"foreignkey:ImageID"`
	Snapshot    int64
	Keys        []*Key `gorm:"many2many:instance_keys;"`
	PasswdLogin bool   `gorm:"default:false"`
	Userdata    string `gorm:"type:text"`
	Hyper       int32  `gorm:"default:-1"`
	ZoneID      int64
	Zone        *Zone `gorm:"foreignkey:ZoneID"`
	RouterID    int64 `gorm:"unique_index:idx_router_instance"`
	Router      *Router
}

func init() {
	dbs.AutoMigrate(&Instance{})
}
