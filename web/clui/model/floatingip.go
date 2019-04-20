package model

import (
	"fmt"
	"strings"

	"github.com/IBM/cloudland/web/sca/dbs"
)

type FloatingIp struct {
	Model
	FipAddress  *Address `gorm:"foreignkey:FloatingIp"`
	InstanceID  int64
	Instance    *Instance `gorm:"foreignkey:InstanceID"`
	InterfaceID int64
	Interface   *Interface `gorm:"foreignkey:InterfaceID"`
	Gateway     int64
}

func init() {
	dbs.AutoMigrate(&FloatingIp{})
}

func AllocateFloatingIp(floatingipID int64, gateway *Gateway, ftype string) (address *Address, err error) {
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
	}
	address, err = AllocateAddress(subnet.ID, floatingipID, "floating")
	if err != nil {
		subnets := []*Subnet{}
		err = db.Model(&Subnet{}).Where("vlan = ? and id <> ?", subnet.Vlan, subnet.ID).Find(subnets).Error
		if err == nil && len(subnets) > 0 {
			for _, s := range subnets {
				address, err = AllocateAddress(s.ID, floatingipID, "floating")
				if err == nil {
					break
				}
			}
		}
	}
	return
}
