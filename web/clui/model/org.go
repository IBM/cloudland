/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"fmt"

	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	dbs.AutoMigrate(&Member{}, &Organization{})
}

type Role int

const (
	None Role = iota
	Reader
	Writer
	Owner
)

func (r Role) String() string {
	switch r {
	case None:
		return "None"
	case Reader:
		return "Reader"
	case Writer:
		return "Writer"
	case Owner:
		return "Owner"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

type Organization struct {
	Model
	Name      string `gorm:"size:255;unique_index" json:"name,omitempty"`
	Owner     int64
	DefaultSG int64
}

type Member struct {
	Model
	UserID   int64
	UserName string
	OrgID    int64
	OrgName  string
	Role     Role
}
