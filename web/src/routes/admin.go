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
	var user *model.User
	var org *model.Organization
	username := "admin"
	dbs.AutoUpgrade("01-admin-upgrade", func(db *gorm.DB) (err error) {
		if err = db.Take(&model.User{}, "username = ?", username).Error; err != nil {
			//replace DB function to avoid AutoUpgrade loop
			dbFunc := DB
			defer func() { DB = dbFunc }()
			DB = func() *gorm.DB { return db }
			password := adminPassword()
			user, err = userAdmin.Create(ctx, username, password)
			if err != nil {
				return
			}
			org, err = orgAdmin.Create(ctx, username, username)
			if err != nil {
				return
			}
		}
		return
	})
	_, err := secgroupAdmin.GetSecgroupByName(ctx, SystemDefaultSGName)
	if err != nil {
		if user == nil {
			user, err = userAdmin.GetUserByName(username)
			if err != nil {
				log.Println("Failed to get user", err)
				return
			}
		}
		if org == nil {
			org, err = orgAdmin.GetOrgByName(username)
			if err != nil {
				log.Println("Failed to get org", err)
				return
			}
		}
		memberShip, err := GetDBMemberShip(user.ID, org.ID)
		if err != nil {
			log.Println("Failed to get membership", err)
			return
		}
		memberShip.Role = model.Admin
		ctx1 := memberShip.SetContext(ctx)
		_, _ = secgroupAdmin.Create(ctx1, SystemDefaultSGName, true, nil)
	}
	return
}
