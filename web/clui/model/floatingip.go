/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type FloatingIp struct {
	Model
	FipAddress string `gorm:"type:varchar(64)"`
	IntAddress string `gorm:"type:varchar(64)"`
	Type       string `gorm:"type:varchar(20)"`
	InstanceID int64
	Instance   *Instance  `gorm:"foreignkey:InstanceID"`
	Interface  *Interface `gorm:"foreignkey:FloatingIp"`
	GatewayID  int64
	Gateway    *Gateway `gorm:"foreignkey:GatewayID"`
}

func init() {
	dbs.AutoMigrate(&FloatingIp{})
}

func AllocateFloatingIp(floatingipID int64, gateway *Gateway, ftype string) (fipIface *Interface, err error) {
	db := dbs.DB()
	var subnet *Subnet
	for _, iface := range gateway.Interfaces {
		if strings.Contains(iface.Type, ftype) {
			subnet = iface.Address.Subnet
			break
		}
	}
	if subnet == nil {
		err = fmt.Errorf("Invalid gateway subnet")
		return
	}
	name := ftype + "fip"
	fipIface, err = CreateInterface(subnet.ID, floatingipID, name, "floating", nil)
	if err != nil {
		subnets := []*Subnet{}
		err = db.Model(&Subnet{}).Where("vlan = ? and id <> ?", subnet.Vlan, subnet.ID).Find(subnets).Error
		if err == nil && len(subnets) > 0 {
			for _, s := range subnets {
				fipIface, err = CreateInterface(s.ID, floatingipID, name, "floating", nil)
				if err == nil {
					break
				}
			}
		}
	}
	return
}

func DeallocateFloatingIp(floatingipID int64) (err error) {
	db := dbs.DB()
	DeleteInterfaces(floatingipID, "floating")
	floatingip := &FloatingIp{Model: Model{ID: floatingipID}}
	err = db.Delete(floatingip).Error
	if err != nil {
		log.Println("Failed to delete floating ip, %v", err)
		return
	}
	return
}
