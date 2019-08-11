/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/go-macaron/session"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	interfaceAdmin = &InterfaceAdmin{}
	interfaceView  = &InterfaceView{}
)

type InterfaceAdmin struct{}

type InterfaceView struct{}

func (a *InterfaceAdmin) Update(ctx context.Context, id int64, name string, sgIDs []int64) (iface *model.Interface, err error) {
	db := DB()
	iface = &model.Interface{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(iface).Error; err != nil {
		log.Println("Failed to query interface ", err)
		return
	}
	if iface.Name != name {
		iface.Name = name
		if err = db.Save(iface).Error; err != nil {
			log.Println("Failed to save interface", err)
			return
		}
	}
	sgChanged := false
	if len(iface.Secgroups) != len(sgIDs) {
		sgChanged = true
	} else {
		for _, esg := range iface.Secgroups {
			found := false
			for _, sgID := range sgIDs {
				if sgID == esg.ID {
					found = true
				}
			}
			if found == false {
				sgChanged = true
				break
			}
		}
	}
	if sgChanged == true {
		secGroups := []*model.SecurityGroup{}
		if err = db.Where(sgIDs).Find(&secGroups).Error; err != nil {
			log.Println("Security group query failed", err)
			return
		}
		var secRules []*model.SecurityRule
		secRules, err = model.GetSecurityRules(secGroups)
		if err != nil {
			log.Println("Failed to get security rules", err)
			return
		}
		securityData := []*SecurityData{}
		for _, rule := range secRules {
			sgr := &SecurityData{
				Secgroup:    rule.Secgroup,
				RemoteIp:    rule.RemoteIp,
				RemoteGroup: rule.RemoteGroup,
				Direction:   rule.Direction,
				IpVersion:   rule.IpVersion,
				Protocol:    rule.Protocol,
				PortMin:     rule.PortMin,
				PortMax:     rule.PortMax,
			}
			securityData = append(securityData, sgr)
		}
		var jsonData []byte
		jsonData, err = json.Marshal(securityData)
		if err != nil {
			log.Println("Failed to marshal security json data, %v", err)
			return
		}
		control := fmt.Sprintf("inter=%d", iface.Hyper)
		if iface.Hyper < 0 {
			instance := &model.Interface{Model: model.Model{ID: iface.Instance}}
			if err = db.Take(instance).Error; err != nil {
				log.Println("Failed to query instance ", err)
				return
			}
			control = fmt.Sprintf("inter=%d", instance.Hyper)
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/reapply_secgroup.sh %s %s <<EOF\n%s\nEOF", iface.Address.Address, iface.MacAddr, jsonData)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Launch vm command execution failed", err)
			return
		}
	}
	return
}

func (v *InterfaceView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	ifaceID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", int64(ifaceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	iface := &model.Interface{Model: model.Model{ID: int64(ifaceID)}}
	if err = db.Set("gorm:auto_preload", true).Take(iface).Error; err != nil {
		log.Println("Image query failed", err)
		return
	}
	_, secgroups, err := secgroupAdmin.List(c.Req.Context(), 0, 0, "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Interface"] = iface
	c.Data["Secgroups"] = secgroups
	c.HTML(200, "interfaces_patch")
}

func (v *InterfaceView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../instances"
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	ifaceID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", int64(ifaceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	name := c.Params("name")
	secgroups := c.QueryStrings("secgroups")
	var sgIDs []int64
	if len(secgroups) > 0 {
		for _, s := range secgroups {
			sID, err := strconv.Atoi(s)
			if err != nil {
				log.Println("Invalid secondary subnet ID, %v", err)
				continue
			}
			permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sID))
			if !permit {
				log.Println("Not authorized for this operation")
				code := http.StatusUnauthorized
				c.Error(code, http.StatusText(code))
				return
			}
			sgIDs = append(sgIDs, int64(sID))
		}
	} else {
		sID := store.Get("defsg").(int64)
		permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sID))
		if !permit {
			log.Println("Not authorized for this operation")
			code := http.StatusUnauthorized
			c.Error(code, http.StatusText(code))
			return
		}
		sgIDs = append(sgIDs, sID)
	}
	_, err = interfaceAdmin.Update(c.Req.Context(), int64(ifaceID), name, sgIDs)
	if err != nil {
		log.Println("Failed to update interface", err)
		c.HTML(http.StatusBadRequest, err.Error())
	}
	c.Redirect(redirectTo)
}

func AllocateAddress(ctx context.Context, subnetID, ifaceID int64, ipaddr, addrType string) (address *model.Address, err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
	subnet := &model.Subnet{Model: model.Model{ID: subnetID}}
	err = db.Take(subnet).Error
	if err != nil {
		log.Println("Failed to query subnet", err)
		return
	}
	address = &model.Address{Subnet: subnet}
	if ipaddr == "" {
		err = db.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ?", subnetID, false).Take(address).Error
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
	ctx, db = getCtxDB(ctx)
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

func SetGateway(ctx context.Context, subnetID, routerID int64) (subnet *model.Subnet, err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
	subnet = &model.Subnet{
		Model: model.Model{ID: subnetID},
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

func UnsetGateway(ctx context.Context, subnet *model.Subnet) (err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
	subnet.Router = 0
	err = db.Save(subnet).Error
	if err != nil {
		log.Println("Failed to unset gateway, %v", err)
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

func CreateInterface(ctx context.Context, subnetID, ID, owner int64, address, mac, ifaceName, ifType string, secGroups []*model.SecurityGroup) (iface *model.Interface, err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
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
		Model:     model.Model{Owner: owner},
		Name:      ifaceName,
		MacAddr:   mac,
		PrimaryIf: primary,
		Subnet:    subnetID,
		Type:      ifType,
		Mtu:       1450,
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
		log.Println("Failed to create interface, %v", err)
		return
	}
	iface.Address, err = AllocateAddress(ctx, subnetID, iface.ID, address, "native")
	if err != nil {
		log.Println("Failed to allocate address", err)
		return
	}
	return
}

func DeleteInterfaces(ctx context.Context, masterID, subnetID int64, ifType string) (err error) {
	var db *gorm.DB
	ctx, db = getCtxDB(ctx)
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
	ctx, db = getCtxDB(ctx)
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
