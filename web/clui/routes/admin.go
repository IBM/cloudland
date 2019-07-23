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
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
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
			_, err = userAdmin.Create(ctx, username, password)
			if err != nil {
				return
			}
			_, err = orgAdmin.Create(ctx, username, username)
			if err != nil {
				return
			}
		}
		return
	})
}
