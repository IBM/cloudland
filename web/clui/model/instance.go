/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Instance struct {
	Model
	Hostname    string        `gorm:"type:varchar(128)"`
	Domain      string        `gorm:"type:varchar(128)"`
	Status      string        `gorm:"type:varchar(32)"`
	Reason      string        `gorm:"type:text"`
	FloatingIps []*FloatingIp `gorm:"PRELOAD:false"`
	Volumes     []*Volume     `gorm:"PRELOAD:false"`
	Interfaces  []*Interface  `gorm:"foreignkey:Instance"`
	Portmaps    []*Portmap    `gorm:"foreignkey:instanceID"`
	FlavorID    int64
	Flavor      *Flavor `gorm:"foreignkey:FlavorID"`
	ImageID     int64
	Image       *Image `gorm:"foreignkey:ImageID"`
	ClusterID   int64
	Cluster     *Openshift `gorm:"PRELOAD:false"`
	Keys        []*Key     `gorm:"many2many:InstanceKeys;"`
	Userdata    string     `gorm:"type:text"`
	Hyper       int32      `gorm:"default:-1"`
}

func init() {
	dbs.AutoMigrate(&Instance{})
}
