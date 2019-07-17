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
	if m.Role >= reqRole || m.UserName == "admin" {
		permit = true
	}
	return
}

func (m *MemberShip) CheckCreater(table string, id int64) (isCreater bool, err error) {
	isCreater = false
	db := DB()
	var creater []int64
	err = db.Table(table).Select("creater").Where("id = ?", id).Scan(&creater).Error
	if err != nil || len(creater) != 1 {
		log.Println("Failed to query resource creater", err)
		return
	}
	if memberShip.UserID == creater[0] || m.UserName == "admin" {
		isCreater = true
	}
	return
}

func (m *MemberShip) CheckUser(id int64) (permit bool, err error) {
	permit = false
	db := DB()
	user := &model.User{}
	err = db.Take(&user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		return
	}
	if user.ID == m.UserID || m.UserName == "admin" {
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
	var owner []int64
	db := DB()
	err = db.Table(table).Select("owner").Where("id = ?", id).Scan(&owner).Error
	if err != nil || len(owner) != 1 {
		log.Println("Failed to query resource owner", err)
		return
	}
	if memberShip.OrgID == owner[0] || m.UserName == "admin" {
		isOwner = true
	}
	return
}
