/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"fmt"
	"log"

	"github.com/IBM/cloudland/web/clui/model"
)

var (
	memberShip = &MemberShip{}
)

type MemberShip struct {
	UserID   int64
	UserName string
	OrgID    int64
	OrgName  string
	Role     model.Role
}

func (m *MemberShip) GetWhere() (where string) {
	if m.UserName != "admin" {
		where = fmt.Sprintf("owner = %d", m.OrgID)
	}
	return
}

func (m *MemberShip) CheckPermission(reqRole model.Role) (permit bool) {
	permit = false
	if m.Role >= reqRole {
		permit = true
	}
	if m.UserName == "admin" {
		permit = true
	}
	return
}

func (m *MemberShip) CheckOwner(reqRole model.Role, table string, id int64) (isOwner bool, err error) {
	isOwner = false
	if id == 0 {
		return
	}
	if m.Role < reqRole {
		return
	}
	var owner int64
	db := DB()
	err = db.Table(table).Select("owner").Where("id = ?", id).Scan(&owner).Error
	if err != nil {
		log.Println("Failed to query resource owner", err)
		return
	}
	if id == owner || m.UserName == "admin" {
		isOwner = true
	}
	return
}
