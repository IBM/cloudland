/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"log"

	"github.com/IBM/cloudland/web/sca/dbs"
)

func init() {
	dbs.AutoMigrate(&SecurityGroup{}, &SecurityRule{})
}

type SecurityGroup struct {
	Model
	Name      string `gorm:"type:varchar(32)"`
	IsDefault bool   `gorm:"default:false"`
}

type SecurityRule struct {
	Model
	Secgroup    int64
	RemoteIp    string `gorm:"type:varchar(32)"`
	RemoteGroup string `gorm:"type:varchar(36)"`
	Direction   string `gorm:"type:varchar(16)"`
	IpVersion   string `gorm:"type:varchar(12)"`
	Protocol    string `gorm:"type:varchar(20)"`
	PortMin     int32  `gorm:"default:-1"`
	PortMax     int32  `gorm:"default:-1"`
}

func init() {
	dbs.AutoMigrate(&SecurityGroup{}, &SecurityRule{})
}

func GetSecurityRules(secGroups []int64) (securityRules []*SecurityRule, err error) {
	db := dbs.DB()
	securityRules = []*SecurityRule{}
	for _, sgID := range secGroups {
		secrules := []*SecurityRule{}
		err = db.Model(&SecurityRule{}).Where("secgroup = ?", sgID).Find(&secrules).Error
		if err != nil {
			log.Println("DB failed to query security rules", err)
			return
		}
		log.Println("$$$$$$$$$$$$$$$$$$$", secrules)
		securityRules = append(securityRules, secrules...)
	}
	return
}
