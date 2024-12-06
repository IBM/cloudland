/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package routes

import (
	"context"
	"log"
	"time"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

func adminPassword() (password string) {
	time.Sleep(time.Second * 5)
	password = viper.GetString("admin.password")
	if password == "" {
		password = "passw0rd"
	}
	return
}

func adminInit(ctx context.Context) {
	username := "admin"
	dbs.AutoUpgrade("01-admin-upgrade", func(db *gorm.DB) (err error) {
		if err = db.Take(&model.User{}, "username = ?", username).Error; err != nil {
			//replace DB function to avoid AutoUpgrade loop
			dbFunc := DB
			defer func() { DB = dbFunc }()
			DB = func() *gorm.DB { return db }
			password := adminPassword()
			var user *model.User
			user, err = userAdmin.Create(ctx, username, password)
			if err != nil {
				return
			}
			var org *model.Organization
			org, err = orgAdmin.Create(ctx, username, username)
			if err != nil {
				return
			}
			var memberShip *MemberShip
			memberShip, err = GetDBMemberShip(user.ID, org.ID)
			if err != nil {
				return
			}
			_, err := secgroupAdmin.GetSecgroupByName(ctx, SystemDefaultName)
			if err != nil {
				ctx1 := memberShip.SetContext(ctx)
				_, err = secgroupAdmin.Create(ctx1, SystemDefaultName, true, nil)
				if err != nil {
					log.Println("Failed to create system default security group", err)
				}
			}
		}
		return
	})
	return
}
