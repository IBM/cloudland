/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package common

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"web/src/model"
	"web/src/utils/log"

	"github.com/jinzhu/gorm"
)

var logger = log.MustGetLogger("common")

type SecurityData struct {
	Secgroup    int64
	RemoteIp    string `json:"remote_ip"`
	RemoteGroup int64  `json:"remote_group"`
	Direction   string `json:"direction"`
	IpVersion   string `json:"ip_version"`
	Protocol    string `json:"protocol"`
	PortMin     int32  `json:"port_min"`
	PortMax     int32  `json:"port_max"`
}

type NetworkRoute struct {
        Network string `json:"network"`
        Netmask string `json:"netmask"`
        Gateway string `json:"gateway"`
}

type InstanceNetwork struct {
        Type    string          `json:"type,omitempty"`
        Address string          `json:"ip_address"`
        Netmask string          `json:"netmask"`
        Link    string          `json:"link"`
        ID      string          `json:"id"`
        Routes  []*NetworkRoute `json:"routes,omitempty"`
}

type AddressInfo struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
}

type SiteIpSubnetInfo struct {
	SiteVlan  int64         `json:"vlan"`
	Gateway   string        `json:"gateway"`
	Addresses []*AddressInfo `json:"addresses"`
}

type VlanInfo struct {
	Device        string              `json:"device"`
	Vlan          int64               `json:"vlan"`
	Gateway       string              `json:"gateway"`
	Router        int64               `json:"router"`
	PublicLink    int64               `json:"public_link"`
	Inbound       int32               `json:"inbound"`
	Outbound      int32               `json:"outbound"`
	AllowSpoofing bool                `json:"allow_spoofing"`
	IpAddr        string              `json:"ip_address"`
	MacAddr       string              `json:"mac_address"`
	SecRules      []*SecurityData     `json:"security"`
	SitesIpInfo    []*SiteIpSubnetInfo `json:"sites_ip_info"`
}

func ApplyInterface(ctx context.Context, instance *model.Instance, iface *model.Interface) (err error) {
	var securityData []*SecurityData
	securityData, err = GetSecurityData(ctx, iface.SecurityGroups)
	if err != nil {
		logger.Debug("DB failed to get security data, %v", err)
		return
	}
	var jsonData []byte
	jsonData, err = json.Marshal(securityData)
	if err != nil {
		logger.Error("Failed to marshal security json data, %v", err)
		return
	}
	subnet := iface.Address.Subnet
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/apply_vm_nic.sh '%d' '%d' '%s' '%s' '%s' '%d' '%d' '%d' '%t'<<EOF\n%s\nEOF", iface.Instance, subnet.Vlan, iface.Address.Address, iface.MacAddr, subnet.Gateway, subnet.RouterID, iface.Inbound, iface.Outbound, iface.AllowSpoofing, jsonData)
	err = HyperExecute(ctx, control, command)
	if err != nil {
		logger.Error("Update vm nic command execution failed", err)
		return
	}
	return
}

func AllocateAddress(ctx context.Context, subnet *model.Subnet, ifaceID int64, ipaddr, addrType string) (address *model.Address, err error) {
	var db *gorm.DB
	ctx, db = GetContextDB(ctx)
	address = &model.Address{}
	if ipaddr == "" {
		err = db.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ? and address != ?", subnet.ID, false, subnet.Gateway).Take(address).Error
	} else {
		if !strings.Contains(ipaddr, "/") {
			preSize, _ := net.IPMask(net.ParseIP(subnet.Netmask).To4()).Size()
			ipaddr = fmt.Sprintf("%s/%d", ipaddr, preSize)
		}
		err = db.Set("gorm:query_option", "FOR UPDATE").Where("subnet_id = ? and allocated = ? and address = ?", subnet.ID, false, ipaddr).Take(address).Error
	}
	if err != nil {
		logger.Error("Failed to query address, %v", err)
		return nil, err
	}
	address.Allocated = true
	address.Type = addrType
	address.Interface = ifaceID
	if err = db.Model(address).Update(address).Error; err != nil {
		logger.Error("Failed to Update address, %v", err)
		return nil, err
	}
	address.Subnet = subnet
	return address, nil
}

func DeallocateAddress(ctx context.Context, ifaces []*model.Interface) (err error) {
	ctx, db := GetContextDB(ctx)
	where := ""
	for i, iface := range ifaces {
		if i == 0 {
			where = fmt.Sprintf("interface='%d'", iface.ID)
		} else {
			where = fmt.Sprintf("%s or interface='%d'", where, iface.ID)
		}
	}
	if err = db.Model(&model.Address{}).Where(where).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error; err != nil {
		logger.Error("Failed to Update addresses, %v", err)
		return
	}
	return
}

func genMacaddr() (mac string, err error) {
	buf := make([]byte, 4)
	_, err = rand.Read(buf)
	if err != nil {
		logger.Error("Failed to generate random numbers, %v", err)
		return
	}
	mac = fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3])
	return mac, nil
}

func CreateInterface(ctx context.Context, subnet *model.Subnet, ID, owner int64, hyper int32, inbound, outbound int32, address, mac, ifaceName, ifType string, secgroups []*model.SecurityGroup) (iface *model.Interface, err error) {
	ctx, db := GetContextDB(ctx)
	primary := false
	if ifaceName == "eth0" {
		primary = true
	}
	if mac == "" {
		mac, err = genMacaddr()
		if err != nil {
			logger.Error("Failed to generate random Mac address, %v", err)
			return
		}
	}
	iface = &model.Interface{
		Owner:          owner,
		Name:           ifaceName,
		MacAddr:        mac,
		PrimaryIf:      primary,
		Inbound:        inbound,
		Outbound:       outbound,
		Subnet:         subnet.ID,
		Hyper:          hyper,
		Type:           ifType,
		Mtu:            1450,
		RouterID:       subnet.RouterID,
		SecurityGroups: secgroups,
	}
	logger.Debugf("Interface: %v", iface)
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
		logger.Error("Failed to create interface, ", err)
		return
	}
	iface.Address, err = AllocateAddress(ctx, subnet, iface.ID, address, "native")
	if err != nil {
		logger.Error("Failed to allocate address", err)
		err2 := db.Delete(iface).Error
		if err2 != nil {
			logger.Error("Failed to delete interface, ", err)
		}
		return
	}
	return
}

func DeleteInterfaces(ctx context.Context, masterID, subnetID int64, ifType string) (err error) {
	ctx, db := GetContextDB(ctx)
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
		logger.Error("Failed to query interfaces, %v", err)
		return
	}
	if len(ifaces) > 0 {
		err = DeallocateAddress(ctx, ifaces)
		if err != nil {
			logger.Error("Failed to deallocate address, %v", err)
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
			logger.Error("Failed to delete interface, %v", err)
			return
		}
	}
	return
}

func DeleteInterface(ctx context.Context, iface *model.Interface) (err error) {
	var db *gorm.DB
	ctx, db = GetContextDB(ctx)
	if err = db.Model(&model.Address{}).Where("interface = ?", iface.ID).Update(map[string]interface{}{"allocated": false, "interface": 0}).Error; err != nil {
		logger.Error("Failed to Update addresses, %v", err)
		return
	}
	err = db.Delete(iface).Error
	if err != nil {
		logger.Error("Failed to delete interface", err)
		return
	}
	return
}

func GetSecurityRules(ctx context.Context, secGroups []*model.SecurityGroup) (securityRules []*model.SecurityRule, err error) {
	ctx, db := GetContextDB(ctx)
	securityRules = []*model.SecurityRule{}
	for _, sg := range secGroups {
		secrules := []*model.SecurityRule{}
		err = db.Model(&model.SecurityRule{}).Where("secgroup = ?", sg.ID).Find(&secrules).Error
		if err != nil {
			logger.Error("DB failed to query security rules", err)
			return
		}
		logger.Debug("Security rule: %v", secrules)
		securityRules = append(securityRules, secrules...)
	}
	return
}

func GetSecurityData(ctx context.Context, secgroups []*model.SecurityGroup) (securityData []*SecurityData, err error) {
	secRules, err := GetSecurityRules(ctx, secgroups)
	if err != nil {
		logger.Error("Failed to get security rules", err)
		return
	}
	for _, rule := range secRules {
		sgr := &SecurityData{
			Secgroup:    rule.Secgroup,
			RemoteIp:    rule.RemoteIp,
			RemoteGroup: rule.RemoteGroupID,
			Direction:   rule.Direction,
			IpVersion:   rule.IpVersion,
			Protocol:    rule.Protocol,
			PortMin:     rule.PortMin,
			PortMax:     rule.PortMax,
		}
		securityData = append(securityData, sgr)
	}
	return
}

func GetInstanceNetworks(ctx context.Context, iface *model.Interface, siteSubnets []*model.Subnet, netID int) (instNetworks []*InstanceNetwork, sitesInfo []*SiteIpSubnetInfo, err error) {
        ctx, db := GetContextDB(ctx)
        subnet := iface.Address.Subnet
        address := strings.Split(iface.Address.Address, "/")[0]
        instNetwork := &InstanceNetwork{
                Address: address,
                Netmask: subnet.Netmask,
                Type:    "ipv4",
                Link:    iface.Name,
                ID:      fmt.Sprintf("network%d", netID),
        }
        if iface.PrimaryIf {
                gateway := strings.Split(subnet.Gateway, "/")[0]
                instRoute := &NetworkRoute{Network: "0.0.0.0", Netmask: "0.0.0.0", Gateway: gateway}
                instNetwork.Routes = append(instNetwork.Routes, instRoute)
        }
        instNetworks = append(instNetworks, instNetwork)
        toUpdate := true
        if len(siteSubnets) == 0 {
                err = db.Where("interface = ?", iface.ID).Find(&iface.SiteSubnets).Error
                if err != nil {
                        logger.Errorf("Failed to query site subnet(s), %v", err)
                        return
                }
                siteSubnets = iface.SiteSubnets
                toUpdate = false
        } else {
                iface.SiteSubnets = siteSubnets
        }
        for _, site := range siteSubnets {
		siteInfo := &SiteIpSubnetInfo{
			SiteVlan: site.Vlan,
			Gateway: site.Gateway,
		}
                siteAddrs := []*model.Address{}
                err = db.Where("subnet_id = ? and address != ?", site.ID, site.Gateway).Find(&siteAddrs).Error
                if err != nil {
                        logger.Errorf("Failed to query site ip(s), %v", err)
                        return
                }
                for _, addr := range siteAddrs {
                        address := fmt.Sprintf("%s/32", strings.Split(addr.Address, "/")[0])
                        instNetworks = append(instNetworks, &InstanceNetwork{
                                Address: address,
                                Netmask: site.Netmask,
                                Type:    "ipv4",
                                Link:    iface.Name,
                                ID:      fmt.Sprintf("network%d", netID),
                        })
			siteInfo.Addresses = append(siteInfo.Addresses, &AddressInfo{
				ID: addr.ID,
				Address: addr.Address,
			})
                }
		sitesInfo = append(sitesInfo, siteInfo)
                if toUpdate {
                        site.Interface = iface.ID
                        err = db.Model(site).Updates(site).Error
                        if err != nil {
                                logger.Errorf("Failed to set site interface", err)
                                return
                        }
                }
        }
        return
}
