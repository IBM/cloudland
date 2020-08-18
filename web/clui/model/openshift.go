/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Openshift struct {
	Model
	ClusterName  string `gorm:"type:varchar(128)"`
	BaseDomain   string `gorm:"type:varchar(256)"`
	Version      string `gorm:"type:varchar(32)"`
	Status       string `gorm:"type:varchar(32)"`
	Haflag       string
	Instances    []*Instance `gorm:"foreignkey:ClusterID"`
	Subnet       *Subnet     `gorm:"foreignkey:ClusterID"`
	WorkerNum    int32
	Flavor       int64
	MasterFlavor int64
	WorkerFlavor int64
	Key          int64
	GlusterID    int64
}

func init() {
	dbs.AutoMigrate(&Openshift{})
}
