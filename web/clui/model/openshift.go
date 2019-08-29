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
	ClusterName string      `gorm:"type:varchar(128)"`
	BaseDomain  string      `gorm:"type:varchar(256)"`
	Instances   []*Instance `gorm:"foreignkey:ClusterID"`
	Haflag      bool
	WorkerNum   int32
	Flavor      int64
	Key         int64
	AdminUser   string `gorm:"type:varchar(128)"`
	AdminPasswd string `gorm:"type:varchar(128)"`
	KubeConfig  string `gorm:"type:text"`
}

func init() {
	dbs.AutoMigrate(&Openshift{})
}
