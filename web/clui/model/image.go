package model

import (
	"github.com/IBM/cloudland/web/sca/dbs"
)

type Image struct {
	Model
	Name         string `gorm:"type:varchar(128)"`
	OSCode       string `gorm:"type:varchar(128)"`
	Format       string `gorm:"type:varchar(128)"`
	Architecture string `gorm:"type:varchar(256)"`
	Status       string `gorm:"type:varchar(128)"`
	Href         string `gorm:"type:varchar(256)"`
	Checksum     string `gorm:"type:varchar(36)"`
}

func init() {
	dbs.AutoMigrate(&Image{})
}
