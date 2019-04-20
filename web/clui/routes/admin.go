package routes

import (
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

func adminPassword() (password string) {
	password = viper.GetString("admin.password")
	if password == "" {
		password = "passw0rd"
	}
	return
}

func init() {
	username := "admin"
	password := adminPassword()
	dbs.AutoUpgrade("01-admin-upgrade", func(db *gorm.DB) (err error) {
		if err = db.Take(&model.User{}, "username = ?", username).Error; err != nil {
			dbFunc := DB
			defer func() { DB = dbFunc }()
			DB = func() *gorm.DB { return db }
			_, err = userAdmin.Create(username, password)
			if err != nil {
				return
			}
			_, err = orgAdmin.Create(username, username)
			if err != nil {
				return
			}
		}
		return
	})
}
