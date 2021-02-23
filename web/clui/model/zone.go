/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"time"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type Zone struct {
	ID        int64 `gorm:"primary_key"`
	Name      string
	Default   bool
	Subnets   []*Subnet `gorm:"many2many:subnet_zones;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func init() {
	dbs.AutoMigrate(&Zone{})
}
