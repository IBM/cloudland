/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package common

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"strings"

	"web/src/model"

	"github.com/jinzhu/gorm"
)

func AllocateAddress(ctx context.Context, subnetID, ifaceID int64, ipaddr, addrType string) (address *model.Address, err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	subnet := &model.Subnet{Model: model.Model{ID: subnetID}}
	err = db.Take(subnet).Error
	if err != nil {
		log.Println("Failed to query subnet", err)
		return
	}
	address = &model.Address{Subnet: subnet}
	if ipaddr == "" {
		err = db.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ? and address != ?", subnetID, false, subnet.Gateway).Take(address).Error
	} else {
		if !strings.Contains(ipaddr, "/") {
			preSize, _ := net.IPMask(net.ParseIP(subnet.Netmask).To4()).Size()
			ipaddr = fmt.Sprintf("%s/%d", ipaddr, preSize)
		}
		err = db.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ? and address = ?", subnetID, false, ipaddr).Take(address).Error
	}
	if err != nil {
		log.Println("Failed to query address, %v", err)
		return nil, err
	}
	address.Allocated = true
	address.Type = addrType
	address.Interface = ifaceID
	if err = db.Model(address).Update(address).Error; err != nil {
		log.Println("Failed to Update address, %v", err)
		return nil, err
	}
	return address, nil
}

func DeallocateAddress(ctx context.Context, ifaces []*model.Interface) (err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	where := ""
	for i, iface := range ifaces {
		if i == 0 {
			where = fmt.Sprintf("interface='%d'", iface.ID)
		} else {
			where = fmt.Sprintf("%s or interface='%d'", where, iface.ID)
		}
	}
	if err = db.Model(&model.Address{}).Where(where).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error; err != nil {
		log.Println("Failed to Update addresses, %v", err)
		return
	}
	return
}

func genMacaddr() (mac string, err error) {
	buf := make([]byte, 4)
	_, err = rand.Read(buf)
	if err != nil {
		log.Println("Failed to generate random numbers, %v", err)
		return
	}
	mac = fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3])
	return mac, nil
}

func CreateInterface(ctx context.Context, subnetID, ID, owner int64, hyper int32, address, mac, ifaceName, ifType string, secGroups []*model.SecurityGroup) (iface *model.Interface, err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	subnet := &model.Subnet{Model: model.Model{ID: subnetID}}
	err = db.Take(subnet).Error
	if err != nil {
		log.Println("DB failed to query subnet, %v", err)
		return
	}
	primary := false
	if ifaceName == "eth0" {
		primary = true
	}
	if mac == "" {
		mac, err = genMacaddr()
		if err != nil {
			log.Println("Failed to generate random Mac address, %v", err)
			return
		}
	}
	iface = &model.Interface{
		Owner:     owner,
		Name:      ifaceName,
		MacAddr:   mac,
		PrimaryIf: primary,
		Subnet:    subnetID,
		Hyper:     hyper,
		Type:      ifType,
		Mtu:       1450,
		RouterID:  subnet.RouterID,
		Secgroups: secGroups,
	}
	if ifType == "instance" {
		iface.Instance = ID
	} else if ifType == "floating" {
		iface.FloatingIp = ID
	} else if ifType == "dhcp" {
		iface.Dhcp = ID
	} else if strings.Contains(ifType, "gateway") {
		iface.Device = ID
	}
	err = db.Create(iface).Error
	if err != nil {
		log.Println("Failed to create interface, ", err)
		return
	}
	iface.Address, err = AllocateAddress(ctx, subnetID, iface.ID, address, "native")
	if err != nil {
		log.Println("Failed to allocate address", err)
		err2 := db.Delete(iface).Error
		if err2 != nil {
			log.Println("Failed to delete interface, ", err)
		}
		return
	}
	return
}

func DeleteInterfaces(ctx context.Context, masterID, subnetID int64, ifType string) (err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	ifaces := []*model.Interface{}
	where := ""
	if subnetID > 0 {
		where = fmt.Sprintf("subnet = %d", subnetID)
	}
	if ifType == "instance" {
		err = db.Where("instance = ? and type = ?", masterID, "instance").Where(where).Find(&ifaces).Error
	} else if ifType == "floating" {
		err = db.Where("floating_ip = ? and type = ?", masterID, "floating").Where(where).Find(&ifaces).Error
	} else if ifType == "dhcp" {
		err = db.Where("dhcp = ? and type = ?", masterID, "dhcp").Where(where).Find(&ifaces).Error
	} else {
		err = db.Where("device = ? and type like ?", masterID, "%gateway%").Where(where).Find(&ifaces).Error
	}
	if err != nil {
		log.Println("Failed to query interfaces, %v", err)
		return
	}
	if len(ifaces) > 0 {
		err = DeallocateAddress(ctx, ifaces)
		if err != nil {
			log.Println("Failed to deallocate address, %v", err)
			return
		}
		if ifType == "instance" {
			err = db.Where("instance = ? and type = ?", masterID, "instance").Where(where).Delete(&model.Interface{}).Error
		} else if ifType == "floating" {
			err = db.Where("floating_ip = ? and type = ?", masterID, "floating").Where(where).Delete(&model.Interface{}).Error
		} else if ifType == "gateway" {
			err = db.Where("device = ? and type like ?", masterID, "%gateway%").Where(where).Delete(&model.Interface{}).Error
		} else if ifType == "dhcp" {
			err = db.Where("dhcp = ? and type = ?", masterID, "dhcp").Where(where).Delete(&model.Interface{}).Error
		}
		if err != nil {
			log.Println("Failed to delete interface, %v", err)
			return
		}
	}
	return
}

func DeleteInterface(ctx context.Context, iface *model.Interface) (err error) {
	var db *gorm.DB
	ctx, db = GetCtxDB(ctx)
	if err = db.Model(&model.Address{}).Where("interface = ?", iface.ID).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error; err != nil {
		log.Println("Failed to Update addresses, %v", err)
		return
	}
	err = db.Delete(iface).Error
	if err != nil {
		log.Println("Failed to delete interface", err)
		return
	}
	return
}
