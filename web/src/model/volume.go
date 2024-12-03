/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

type Volume struct {
	Model
	Owner int64  `gorm:"default:1","index"` /* The organization ID of the resource */
	Name  string `gorm:"type:varchar(128)"`
	/*
		The path of the volume, format is:
		<volume driver>://<volume-path>
		for example:
		Local storage: local:///var/lib/cloudland/volumes/volume-1.qcow2
		WDS Vhost: wds_vhost://<pool-id>/<volume-id>
		WDS ISCSI: wds_iscsi://<pool-id>/<volume-id>
		The volume driver is the name of the driver that is used to create the volume.
	*/
	Path       string `gorm:"type:varchar(256)"`
	Size       int32
	Format     string `gorm:"type:varchar(32)"`
	Status     string `gorm:"type:varchar(32)"`
	Target     string `gorm:"type:varchar(32)"`
	Href       string `gorm:"type:varchar(256)"`
	InstanceID int64
	Instance   *Instance `gorm:"foreignkey:InstanceID"`
	IopsLimit  int32
	IopsBurst  int32
	BpsLimit   int32
	BpsBurst   int32
}

func init() {
	dbs.AutoMigrate(&Volume{})
}
