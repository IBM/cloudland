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
	if m.Role != model.Admin {
		where = fmt.Sprintf("owner = %d", m.OrgID)
	}
	return
}

func (m *MemberShip) CheckPermission(reqRole model.Role) (permit bool) {
	permit = false
	if m.Role >= reqRole || m.Role == model.Admin {
		permit = true
	}
	return
}

func (m *MemberShip) CheckCreater(table string, id int64) (isCreater bool, err error) {
	isCreater = false
	db := DB()
	type Result struct {
		Creater int64
	}
	var result Result
	err = db.Table(table).Select("creater").Where("id = ?", id).Scan(&result).Error
	if err != nil {
		log.Println("Failed to query resource creater", err)
		return
	}
	if memberShip.UserID == result.Creater || (m.OrgName == "admin" && m.Role == model.Admin) {
		isCreater = true
	}
	return
}

func (m *MemberShip) CheckUser(id int64) (permit bool, err error) {
	permit = false
	db := DB()
	user := &model.User{Model: model.Model{ID: id}}
	err = db.Take(&user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		return
	}
	if user.ID == m.UserID || (m.OrgName == "admin" && m.Role == model.Admin) {
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
	type Result struct {
		Owner int64
	}
	var result Result
	db := DB()
	err = db.Table(table).Select("owner").Where("id = ?", id).Scan(&result).Error
	if err != nil {
		log.Println("Failed to query resource owner", err)
		return
	}
	if memberShip.OrgID == result.Owner || (m.OrgName == "admin" && m.Role == model.Admin) {
		isOwner = true
	}
	return
}
