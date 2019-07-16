/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"fmt"

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
