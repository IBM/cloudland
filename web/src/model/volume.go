/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"strings"
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
	Booting    bool
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

func (v *Volume) ParsePath() []string {
	if v.Path != "" {
		parts := strings.SplitN(v.Path, "://", 2)
		if len(parts) == 2 {
			driver := parts[0]
			if driver == "local" {
				return []string{driver, parts[1]}
			} else {
				res := []string{driver}
				res = append(res, strings.Split(parts[1], "/")...)
				return res
			}
		}
	}
	return nil
}

func (v *Volume) GetVolumeDriver() string {
	parts := v.ParsePath()
	if parts != nil {
		return parts[0]
	}
	return ""
}

func (v *Volume) GetVolumePath() string {
	parts := v.ParsePath()
	if (parts != nil) && (parts[0] == "local") {
		return parts[1]
	}
	return v.Path
}

func (v *Volume) GetVolumePoolID() string {
	parts := v.ParsePath()
	if (parts != nil) && (len(parts) == 3) {
		return parts[1]
	}
	return ""
}

func (v *Volume) GetOriginVolumeID() string {
	parts := v.ParsePath()
	if (parts != nil) && (len(parts) == 3) {
		return parts[2]
	}
	return ""
}

func init() {
	dbs.AutoMigrate(&Volume{})
}
