/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Registry struct {
	Model
	Label           string `gorm:"type:varchar(128)"`
	OcpVersion      string `gorm:"type:varchar(128)"`
	VirtType        string `gorm:"type:varchar(128)"`
	RegistryContent string `gorm:"type:varchar(20480)"`
	Initramfs       string `gorm:"type:varchar(1280)"`
	Kernel          string `gorm:"type:varchar(1280)"`
	Image           string `gorm:"type:varchar(1280)"`
	Installer       string `gorm:"type:varchar(1280)"`
	Cli             string `gorm:"type:varchar(1280)"`
}

func init() {
	dbs.AutoMigrate(&Registry{})
}
