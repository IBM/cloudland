/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Glusterfs struct {
	Model
	Name      string `gorm:"type:varchar(128)"`
	Endpoint  string `gorm:"type:varchar(256)"`
	Status    string `gorm:"type:varchar(32)"`
	WorkerNum int32
	Flavor    int64
	Key       int64
	HeketiKey int64
	SubnetID  int64
	Subnet    *Subnet `gorm:"foreignkey:SubnetID"`
	ClusterID int64
}

func init() {
	dbs.AutoMigrate(&Glusterfs{})
}
