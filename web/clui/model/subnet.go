/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"fmt"
	"log"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type Subnet struct {
	Model
	Name    string `gorm:"type:varchar(32)"`
	Network string `gorm:"type:varchar(64)"`
	Netmask string `gorm:"type:varchar(64)"`
	Gateway string `gorm:"type:varchar(64)"`
	Start   string `gorm:"type:varchar(64)"`
	End     string `gorm:"type:varchar(64)"`
	Vlan    int64
	Type    string `gorm:"default:'internal'"`
	Router  int64
}

type Address struct {
	Model
	Address    string `gorm:"type:varchar(64)"`
	Netmask    string `gorm:"type:varchar(64)"`
	Type       string `gorm:"default:'native'"`
	Allocated  bool   `gorm:"default:false"`
	Reserved   bool   `gorm:"default:false"`
	SubnetID   int64
	Subnet     *Subnet `gorm:"foreignkey:SubnetID"`
	Interface  int64
	FloatingIp int64
}

func init() {
	dbs.AutoMigrate(&Subnet{})
	dbs.AutoMigrate(&Address{})
}

func AllocateAddress(subnetID, ifaceID int64, addrType string) (address *Address, err error) {
	address = &Address{}
	tx := dbs.DB().Begin()
	err = tx.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ?", subnetID, 0).Take(address).Error
	if err != nil {
		tx.Rollback()
		log.Println("Failed to query address, %v", err)
		return nil, err
	}
	address.Allocated = true
	address.Interface = ifaceID
	address.Type = addrType
	if err = tx.Model(address).Update(address).Error; err != nil {
		tx.Rollback()
		log.Println("Failed to Update address, %v", err)
		return nil, err
	}
	tx.Commit()
	return address, nil
}

func DeallocateAddress(ifaces []*Interface) (err error) {
	db := dbs.DB()
	where := ""
	for i, iface := range ifaces {
		if i == 0 {
			where = fmt.Sprintf("interface='%d'", iface.ID)
		} else {
			where = fmt.Sprintf("%s or interface='%d'", where, iface.ID)
		}
	}
	if err = db.Model(&Address{}).Where(where).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error; err != nil {
		log.Println("Failed to Update addresses, %v", err)
		return
	}
	return
}

func SetGateway(subnetID, routerID int64) (subnet *Subnet, err error) {
	db := dbs.DB()
	subnet = &Subnet{
		Model:  Model{ID: subnetID},
		Router: routerID,
	}
	err = db.Model(subnet).Update(subnet).Error
	if err != nil {
		log.Println("Failed to set gateway, %v", err)
		return nil, err
	}
	return
}

func UnsetGateway(subnetID int64) (err error) {
	db := dbs.DB()
	subnet := &Subnet{
		Model: Model{ID: subnetID},
	}
	err = db.Model(subnet).Update("router = ?", "").Error
	if err != nil {
		log.Println("Failed to unset gateway, %v", err)
		return
	}
	return
}
