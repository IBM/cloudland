/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"web/src/dbs"
)

const (
	OS_LINUX   = "linux"
	OS_WINDOWS = "windows"
	OS_OTHER   = "other"
)

// OSCodes is a list of supported operating systems
var OSCodes = []string{OS_LINUX, OS_WINDOWS, OS_OTHER}

type Image struct {
	Model
	Owner                 int64     `gorm:"default:1"` /* The organization ID of the resource */
	Name                  string    `gorm:"type:varchar(128)"`
	OSCode                string    `gorm:"type:varchar(128);default:'linux'"`
	Format                string    `gorm:"type:varchar(128)"`
	Architecture          string    `gorm:"type:varchar(256)"`
	Status                string    `gorm:"type:varchar(128)"`
	Href                  string    `gorm:"type:varchar(256)"`
	Checksum              string    `gorm:"type:varchar(36)"`
	OsHashAlgo            string    `gorm:"type:varchar(36)"`
	OsHashValue           string    `gorm:"type:varchar(36)"`
	Holder                string    `gorm:"type:varchar(36)"`
	Protected             bool      `gorm:"default:false"`
	Visibility            string    `gorm:"type:varchar(36)"`
	MiniDisk              int32     `gorm:"default:0"`
	MiniMem               int32     `gorm:"default:0"`
	Size                  int64     `gorm:"default:0"`
	OsVersion             string    `gorm:"type:varchar(128)"`
	VirtType              string    `gorm:"type:varchar(36)"`
	UserName              string    `gorm:"type:varchar(128)"`
	QAEnabled             bool      `gorm:"default:false"`
	CaptureFromInstanceID int64     `gorm:"default:0"`
	CaptureFromInstance   *Instance `gorm:"foreignkey:InstanceID"`
}

func init() {
	dbs.AutoMigrate(&Image{})
}

func (i *Image) Clone() *Image {
	return &Image{
		Owner:                 i.Owner,
		Name:                  i.Name,
		OSCode:                i.OSCode,
		Format:                i.Format,
		Architecture:          i.Architecture,
		Status:                i.Status,
		Href:                  i.Href,
		Checksum:              i.Checksum,
		OsHashAlgo:            i.OsHashAlgo,
		OsHashValue:           i.OsHashValue,
		Holder:                i.Holder,
		Protected:             i.Protected,
		Visibility:            i.Visibility,
		MiniDisk:              i.MiniDisk,
		MiniMem:               i.MiniMem,
		Size:                  i.Size,
		OsVersion:             i.OsVersion,
		VirtType:              i.VirtType,
		UserName:              i.UserName,
		QAEnabled:             i.QAEnabled,
		CaptureFromInstanceID: i.CaptureFromInstanceID,
		CaptureFromInstance:   i.CaptureFromInstance,
	}
}
