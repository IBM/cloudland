/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	subnetAdmin = &SubnetAdmin{}
	subnetView  = &SubnetView{}
	vniMax      = 16777215
	vniMin      = 4096
)

type SubnetAdmin struct{}
type SubnetView struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
	return
}

func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		return val, 32
	} else if len(ip) == net.IPv6len {
		return val, 128
	} else {
		panic(fmt.Errorf("Unsupported address length %d", len(ip)))
	}
}

func getValidVni() (vni int, err error) {
	db := DB()
	count := 1
	for count > 0 {
		vni = rand.Intn(vniMax-vniMin) + vniMin
		if err = db.Model(&model.Subnet{}).Where("vlan = ?", vni).Count(&count).Error; err != nil {
			log.Println("Failed to query existing vlan, %v", err)
			return
		}
	}
	return
}

func (a *SubnetAdmin) Create(name, vlan, network, netmask, gateway, start, end, rtype string) (subnet *model.Subnet, err error) {
	db := DB()
	vlanNo := 0
	if vlan == "" {
		vlanNo, err = getValidVni()
	} else {
		vlanNo, err = strconv.Atoi(vlan)
	}
	if err != nil {
		log.Println("Failed to get valid vlan %s, %v", vlan, err)
		return
	}
	inNet := &net.IPNet{
		IP:   net.ParseIP(network),
		Mask: net.IPMask(net.ParseIP(netmask).To4()),
	}
	_, ipNet, err := net.ParseCIDR(inNet.String())
	if err != nil {
		log.Println("CIDR parsing failed, %v", err)
		return
	}
	addrCount := cidr.AddressCount(ipNet)
	if addrCount < 5 || addrCount > 1000 {
		err = fmt.Errorf("Network/mask must have more than 5 but less than 1000 addresses")
		log.Println("Invalid address count", err)
		return
	}
	if rtype == "" {
		rtype = "internal"
	}
	first, last := cidr.AddressRange(ipNet)
	preSize, _ := inNet.Mask.Size()
	if gateway == "" {
		gateway = cidr.Inc(first).String()
	}
	if start == "" {
		start = cidr.Inc(first).String()
	}
	if start == gateway {
		start = cidr.Inc(net.ParseIP(start)).String()
	}
	if end == "" {
		end = cidr.Dec(last).String()
	}
	if end == gateway {
		end = cidr.Dec(net.ParseIP(end)).String()
	}
	if rtype == "" {
		rtype = "internal"
	}
	gateway = fmt.Sprintf("%s/%d", gateway, preSize)
	subnet = &model.Subnet{Name: name, Network: first.String(), Netmask: netmask, Gateway: gateway, Start: start, End: end, Vlan: int64(vlanNo), Type: rtype}
	err = db.Create(subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		return
	}
	ip := net.ParseIP(start)
	for {
		ipstr := fmt.Sprintf("%s/%d", ip.String(), preSize)
		address := &model.Address{Address: ipstr, Netmask: netmask, Type: "ipv4", SubnetID: subnet.ID}
		err = db.Create(address).Error
		if err != nil {
			log.Println("Database create address failed, %v", err)
		}
		if ip.String() == end {
			break
		}
		ip = cidr.Inc(ip)
		if ipstr == gateway {
			ip = cidr.Inc(ip)
		}
	}
	return
}

func (a *SubnetAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	err = db.Where("subnet_id = ?", id).Delete(&model.Address{}).Error
	if err != nil {
		log.Println("Database delete addresses failed, %v", err)
		return
	}
	err = db.Delete(&model.Subnet{Model: model.Model{ID: id}}).Error
	if err != nil {
		log.Println("Database delete subnet failed, %v", err)
		return
	}
	return
}

func (a *SubnetAdmin) List(offset, limit int64, order string) (total int64, subnets []*model.Subnet, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	subnets = []*model.Subnet{}
	if err = db.Model(&model.Subnet{}).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&subnets).Error; err != nil {
		return
	}

	return
}

func (v *SubnetView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, subnets, err := subnetAdmin.List(offset, limit, order)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Subnets"] = subnets
	c.Data["Total"] = total
	c.HTML(200, "subnets")
}

func (v *SubnetView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	subnetID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = subnetAdmin.Delete(int64(subnetID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "subnets",
	})
	return
}

func (v *SubnetView) New(c *macaron.Context, store session.Store) {
	c.HTML(200, "subnets_new")
}

func (v *SubnetView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../subnets"
	name := c.Query("name")
	vlan := c.Query("vlan")
	rtype := c.Query("rtype")
	network := c.Query("network")
	netmask := c.Query("netmask")
	gateway := c.Query("gateway")
	start := c.Query("start")
	end := c.Query("end")
	_, err := subnetAdmin.Create(name, vlan, network, netmask, gateway, start, end, rtype)
	if err != nil {
		log.Println("Create subnet failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
