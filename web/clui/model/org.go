/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package model

import (
	"encoding/gob"
	"fmt"

	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	dbs.AutoMigrate(&Member{}, &Organization{})
	var role Role
	gob.Register(role)
	gob.Register(Member{})
	gob.Register([]*Member{})
}

type Role int

const (
	None   Role = iota /* No permissions  */
	Reader             /* Get List permissions */
	Writer             /* Create Edit Patch permission */
	Owner              /* Invite or Remove user to from org */
	Admin              /* Create user and org */
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
	case Admin:
		return "Admin"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

type Organization struct {
	Model
	Name      string `gorm:"size:255;unique_index" json:"name,omitempty"`
	DefaultSG int64
	Members   []*Member `gorm:"foreignkey:OrgID"`
	OwnerUser *User     `gorm:"foreignkey:ID";AssociationForeignKey:Owner`
}

func (Organization) TableName() string {
	return "organizations"
}

type Member struct {
	Model
	UserID   int64
	UserName string
	OrgID    int64
	OrgName  string
	Role     Role
}
