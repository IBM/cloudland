/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

type MemberShip struct {
	UserID   int64
	UserName string
	OrgID    int64
	OrgName  string
	Role     model.Role
}

func (m *MemberShip) GetWhere() (where string) {
	if m.OrgName == "admin" && m.Role == model.Admin {
		where = ""
	} else {
		where = fmt.Sprintf("owner = %d", m.OrgID)
	}
	return
}

func (m *MemberShip) CheckPermission(reqRole model.Role) (permit bool) {
	permit = false
	if m.Role >= reqRole || (m.OrgName == "admin" && m.Role == model.Admin) {
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
	if m.UserID == result.Creater || (m.OrgName == "admin" && m.Role == model.Admin) {
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
	if m.OrgID == result.Owner || m.Role == model.Admin {
		isOwner = true
	}
	return
}

func (m *MemberShip) CheckAdmin(reqRole model.Role, table string, id int64) (admin bool, err error) {
	admin = false
	if m.Role == model.Admin {
		admin = true
		return
	}
	admin, err = m.CheckOwner(reqRole, table, id)
	return
}

func (m *MemberShip) SetContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "membership", m)
}

func GetMemberShip(ctx context.Context) *MemberShip {
	m := ctx.Value("membership")
	if m != nil {
		return m.(*MemberShip)
	}
	return &MemberShip{}
}

func GetDBMemberShip(userID, orgID int64) (ms *MemberShip, err error) {
	db := dbs.DB()
	m := &MemberShip{
		UserID: userID,
		OrgID:  orgID,
	}
	member := &model.Member{}
	err = db.Where("user_id = ? and org_id = ?", userID, orgID).Take(member).Error
	if err != nil || member.ID == 0 {
		log.Println("Failed to query member", err)
		return
	}
	m.UserName = member.UserName
	m.OrgName = member.OrgName
	m.Role = member.Role
	return
}
