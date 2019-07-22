/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
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
		if err = db.Model(&model.Network{}).Where("vlan = ?", vni).Count(&count).Error; err != nil {
			log.Println("Failed to query existing vlan, %v", err)
			return
		}
	}
	return
}

func checkIfExistVni(vni int64) (result bool, err error) {
	db := DB()
	count := 0
	if err = db.Model(&model.Network{}).Where("vlan = ?", vni).Count(&count).Error; err != nil {
		log.Println("Failed to query existing vlan, %v", err)
		return
	}
	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (a *SubnetAdmin) Create(ctx context.Context, name, vlan, network, netmask, gateway, start, end, rtype, uuid string, owner int64) (subnet *model.Subnet, err error) {
	if owner == 0 {
		owner = memberShip.OrgID
	}
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
	count := 0
	err = db.Model(&model.Subnet{}).Where("vlan = ?", vlanNo).Count(&count).Error
	if err != nil {
		log.Println("Database failed to count network", err)
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
	gateway = fmt.Sprintf("%s/%d", gateway, preSize)
	subnet = &model.Subnet{
		Model:   model.Model{Creater: memberShip.UserID, Owner: owner, UUID: uuid},
		Name:    name,
		Network: first.String(),
		Netmask: netmask,
		Gateway: gateway,
		Start:   start,
		End:     end,
		Vlan:    int64(vlanNo),
		Type:    rtype,
	}
	err = db.Create(subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		return
	}
	ip := net.ParseIP(start)
	for {
		ipstr := fmt.Sprintf("%s/%d", ip.String(), preSize)
		address := &model.Address{Model: model.Model{Creater: memberShip.UserID, Owner: owner}, Address: ipstr, Netmask: netmask, Type: "ipv4", SubnetID: subnet.ID}
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
	netlink := &model.Network{Vlan: int64(vlanNo), Type: "vxlan"}
	if vlanNo < 4096 {
		netlink.Type = "vlan"
	}
	if count < 1 {
		netlink.Creater = memberShip.UserID
		netlink.Owner = owner
		err = db.Create(netlink).Error
		if err != nil {
			log.Println("Database failed to create network", err)
			return
		}
	} else {
		err = db.Where(netlink).Take(netlink).Error
		if err != nil {
			log.Println("Database failed to query network", err)
			return
		}
	}
	err = execNetwork(ctx, netlink, subnet, owner)
	if err != nil {
		log.Println("Failed remote execute network creation", err)
		return
	}
	return
}

func execNetwork(ctx context.Context, netlink *model.Network, subnet *model.Subnet, owner int64) (err error) {
	if netlink.Hyper < 0 {
		var dhcp1 *model.Interface
		dhcp1, err = CreateInterface(ctx, subnet.ID, netlink.ID, owner, "", "dhcp-1", "dhcp", nil)
		if err != nil {
			log.Println("Failed to allocate dhcp first address", err)
			return
		}
		control := fmt.Sprintf("inter=")
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_net.sh %d %s %s %s %s %d FIRST", netlink.Vlan, subnet.Network, subnet.Netmask, subnet.Gateway, dhcp1.Address.Address, subnet.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Failed to create first dhcp", err)
			return
		}
	}
	if netlink.Peer < 0 {
		var dhcp2 *model.Interface
		dhcp2, err = CreateInterface(ctx, subnet.ID, netlink.ID, owner, "", "dhcp-2", "dhcp", nil)
		if err != nil {
			log.Println("Failed to allocate dhcp first address", err)
			return
		}
		control := fmt.Sprintf("inter=")
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_net.sh %d %s %s %s %s %d SECOND", netlink.Vlan, subnet.Network, subnet.Netmask, subnet.Gateway, dhcp2.Address.Address, subnet.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Failed to create second dhcp", err)
			return
		}
	}
	return
}

func (a *SubnetAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	subnet := &model.Subnet{Model: model.Model{ID: id}}
	err = db.Preload("Netlink").Take(subnet).Error
	if err != nil {
		log.Println("Database failed to query subnet", err)
		return
	}
	if subnet.Router > 0 {
		err = fmt.Errorf("Subnet belongs to a gateway")
		log.Println("Subnet belongs to a gateway", err)
		return
	}
	count := 0
	err = db.Model(&model.Address{}).Where("subnet_id = ? and allocated = ?", id, true).Count(&count).Error
	if err != nil {
		log.Println("Database delete addresses failed, %v", err)
		return
	}
	if count > 2 {
		err = fmt.Errorf("Some addresses of this subnet in use")
		log.Println("There are addresses of this subnet still in use")
		return
	}
	err = db.Model(&model.Subnet{}).Where("vlan = ?", subnet.Vlan).Count(&count).Error
	if err != nil {
		log.Println("Database failed to count subnet", err)
		return
	}
	err = db.Delete(subnet).Error
	if err != nil {
		log.Println("Database delete subnet failed, %v", err)
		return
	}
	//delete ip address
	err = db.Where("subnet_id = ?", id).Delete(model.Address{}).Error
	if err != nil {
		log.Println("Database delete ip address failed, %v", err)
		return
	}
	netlink := subnet.Netlink
	if count <= 1 && netlink != nil {
		err = db.Where("dhcp = ?", netlink.ID).Delete(&model.Interface{}).Error
		if err != nil {
			log.Println("Failed to delete dhcp interfaces")
			return
		}
		err = db.Delete(netlink).Error
		if err != nil {
			log.Println("Failed to delete network")
			return
		}
		control := ""
		if netlink.Hyper >= 0 {
			control = fmt.Sprintf("toall=vlan-%d:%d", subnet.Vlan, netlink.Hyper)
			if netlink.Peer >= 0 {
				control = fmt.Sprintf("%s,%d", control, netlink.Peer)
			}
		} else if netlink.Peer >= 0 {
			control = fmt.Sprintf("inter=%d", netlink.Peer)
		} else {
			log.Println("Network has no valid hypers")
			return
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_net.sh %d %s %d", netlink.Vlan, subnet.Network, subnet.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete interface failed")
			return
		}
	}
	return
}

func (a *SubnetAdmin) DeleteByUUID(uuid string) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	count := 0
	subnets := []*model.Subnet{}
	db.Where("uuid = ?", uuid).Find(&subnets)
	for _, subnet := range subnets {
		err = db.Model(&model.Address{}).Where("subnet_id = ? and allocated = ?", subnet.ID, true).Count(&count).Error
		if err != nil {
			log.Println("Database delete addresses failed, %v", err)
			return
		}
		if count > 0 {
			err = fmt.Errorf("Some addresses of this network in use")
			log.Println("There are addresses of this network still in use")
			return
		}
		err = db.Delete(&model.Subnet{Model: model.Model{ID: subnet.ID}}).Error
		if err != nil {
			log.Println("Database delete network failed, %v", err)
			return
		}
		//delete ip address
		err = db.Where("subnet_id = ?", subnet.ID).Delete(model.Address{}).Error
		if err != nil {
			log.Println("Database delete ip address failed, %v", err)
			return
		}
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

	where := ""
	wm := memberShip.GetWhere()
	if wm != "" {
		where = fmt.Sprintf("type != 'internal' or %s", wm)
	}
	subnets = []*model.Subnet{}
	if err = db.Model(&model.Subnet{}).Where(where).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Find(&subnets).Error; err != nil {
		return
	}

	return
}

func (v *SubnetView) List(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
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
	permit, err := memberShip.CheckOwner(model.Writer, "subnets", int64(subnetID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = subnetAdmin.Delete(c.Req.Context(), int64(subnetID))
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
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "subnets_new")
}

func (v *SubnetView) Create(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../subnets"
	name := c.Query("name")
	vlan := c.Query("vlan")
	rtype := c.Query("rtype")
	network := c.Query("network")
	netmask := c.Query("netmask")
	gateway := c.Query("gateway")
	start := c.Query("start")
	end := c.Query("end")
	_, err := subnetAdmin.Create(c.Req.Context(), name, vlan, network, netmask, gateway, start, end, rtype, "", memberShip.OrgID)
	if err != nil {
		log.Println("Create subnet failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
