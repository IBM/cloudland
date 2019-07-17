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

type Network struct {
	Model
	Name    string `gorm:"type:varchar(100)"`
	Hyper   int32  `gorm:"default:-1"`
	Peer    int32  `gorm:"default:-1"`
	Vlan    int64
	Type    string
	Subnets []*Subnet `gorm:"foreignkey:Vlan;AssociationForeignKey:Vlan;PRELOAD:false"`
}

type Subnet struct {
	Model
	Name    string `gorm:"type:varchar(32)"`
	Network string `gorm:"type:varchar(64)"`
	Netmask string `gorm:"type:varchar(64)"`
	Gateway string `gorm:"type:varchar(64)"`
	Start   string `gorm:"type:varchar(64)"`
	End     string `gorm:"type:varchar(64)"`
	Vlan    int64
	Netlink *Network `gorm:"foreignkey:Vlan;AssociationForeignKey:Vlan"`
	Type    string   `gorm:"type:varchar(20);default:'internal'"`
	Router  int64
}

type Address struct {
	Model
	Address   string `gorm:"type:varchar(64)"`
	Netmask   string `gorm:"type:varchar(64)"`
	Type      string `gorm:"type:varchar(20);default:'native'"`
	Allocated bool   `gorm:"default:false"`
	Reserved  bool   `gorm:"default:false"`
	SubnetID  int64
	Subnet    *Subnet `gorm:"foreignkey:SubnetID"`
	Interface int64
}

func init() {
	dbs.AutoMigrate(&Network{})
	dbs.AutoMigrate(&Subnet{})
	dbs.AutoMigrate(&Address{})
}

func AllocateAddress(subnetID, ifaceID int64, addrType string) (address *Address, err error) {
	db := dbs.DB()
	subnet := &Subnet{Model: Model{ID: subnetID}}
	err = db.Take(subnet).Error
	if err != nil {
		log.Println("Failed to query subnet", err)
		return
	}
	address = &Address{Subnet: subnet}
	tx := dbs.DB().Begin()
	err = tx.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ?", subnetID, false).Take(address).Error
	if err != nil {
		tx.Rollback()
		log.Println("Failed to query address, %v", err)
		return nil, err
	}
	address.Allocated = true
	address.Type = addrType
	address.Interface = ifaceID
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
		Model: Model{ID: subnetID},
	}
	err = db.Model(subnet).Take(subnet).Error
	if err != nil {
		log.Println("Failed to get subnet, %v", err)
		return nil, err
	}
	subnet.Router = routerID
	err = db.Model(subnet).Save(subnet).Error
	if err != nil {
		log.Println("Failed to set gateway, %v", err)
		return nil, err
	}
	return
}

func UnsetGateway(subnet *Subnet) (err error) {
	db := dbs.DB()
	subnet.Router = 0
	err = db.Save(subnet).Error
	if err != nil {
		log.Println("Failed to unset gateway, %v", err)
		return
	}
	return
}
