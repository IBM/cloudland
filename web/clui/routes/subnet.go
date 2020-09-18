/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
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

func generateIPAddresses(subnet *model.Subnet, start net.IP, end net.IP, preSize int) (err error) {
	db := DB()
	ip := start
	for {
		ipstr := fmt.Sprintf("%s/%d", ip.String(), preSize)
		if ipstr == subnet.Gateway {
			if ip.String() == end.String() {
				break
			} else {
				ip = cidr.Inc(ip)
				ipstr = fmt.Sprintf("%s/%d", ip.String(), preSize)
			}
		}
		address := &model.Address{
			Model: model.Model{
				Creater: subnet.Creater,
				Owner:   subnet.Owner,
			},
			Address:  ipstr,
			Netmask:  subnet.Netmask,
			Type:     "ipv4",
			SubnetID: subnet.ID,
		}
		err = db.Create(address).Error
		if err != nil {
			log.Println("Database create IP address failed, %v", err)
			return err
		}
		if ip.String() == end.String() {
			break
		}
		ip = cidr.Inc(ip)
	}
	return nil
}

func (a *SubnetAdmin) Update(ctx context.Context, id int64, name, gateway, start, end, routes string) (subnet *model.Subnet, err error) {
	db := DB()
	subnet = &model.Subnet{Model: model.Model{ID: id}}
	err = db.Take(subnet).Error
	if err != nil {
		log.Println("DB failed to query subnet ", err)
		return
	}
	if name != "" {
		subnet.Name = name
	}
	preSize, _ := net.IPMask(net.ParseIP(subnet.Netmask).To4()).Size()
	if gateway != "" {
		subnet.Gateway = fmt.Sprintf("%s/%d", gateway, preSize)
	}
	if (start != "" && subnet.Start != start) || (end != "" && subnet.End != end) {
		if start == "" {
			start = subnet.Start
		}
		if end == "" {
			end = subnet.End
		}
		if bytes.Compare(net.ParseIP(start), net.ParseIP(subnet.Start)) > 0 || bytes.Compare(net.ParseIP(end), net.ParseIP(subnet.End)) < 0 {
			log.Println("Subnet Update failed: only allow expansion of IP address range")
			err = fmt.Errorf("Update_subnet_reduce_ip_range")
			return
		}
		if bytes.Compare(net.ParseIP(start), net.ParseIP(subnet.Start)) < 0 {
			err = generateIPAddresses(subnet, net.ParseIP(start), cidr.Dec(net.ParseIP(subnet.Start)), preSize)
			if err != nil {
				return
			}
			subnet.Start = start
		}
		if bytes.Compare(net.ParseIP(end), net.ParseIP(subnet.End)) > 0 {
			err = generateIPAddresses(subnet, cidr.Inc(net.ParseIP(subnet.End)), net.ParseIP(end), preSize)
			if err != nil {
				return
			}
			subnet.End = end
		}
	}
	subnet.Routes = routes
	err = db.Save(subnet).Error
	if err != nil {
		log.Println("DB failed to save subnet ", err)
		return
	}
	if subnet.Router > 0 {
		err = setRouting(ctx, subnet.ID, subnet, false)
		if err != nil {
			log.Println("Failed to set routing for subnet")
			return
		}
	} else if subnet.Type != "internal" {
		var ifaces []*model.Interface
		ifType := fmt.Sprintf("gateway_%s", subnet.Type)
		err = db.Where("type = ? and subnet = ?", ifType, subnet.ID).Find(&ifaces).Error
		if err != nil {
			log.Println("DB failed to query interfaces")
			return
		}
		for _, iface := range ifaces {
			err = setRouting(ctx, iface.Device, subnet, true)
			if err != nil {
				log.Println("Failed to set routing for subnet")
			}
		}
	}
	return
}

func setRouting(ctx context.Context, gatewayID int64, subnet *model.Subnet, routeOnly bool) (err error) {
	db := DB()
	gateway := &model.Gateway{Model: model.Model{ID: gatewayID}}
	err = db.Take(gateway).Error
	if err != nil {
		log.Println("DB failed to query router", err)
		return
	}
	control := fmt.Sprintf("toall=router-%d:%d,%d", gateway.ID, gateway.Hyper, gateway.Peer)
	if routeOnly {
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_route.sh '%d' '%d' '%s'<<EOF\n%s\nEOF", gateway.ID, subnet.Vlan, subnet.Type, subnet.Routes)
		err = hyperExecute(ctx, control, command)
	} else {
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_gw_route.sh '%d' '%s' '%d' soft <<EOF\n%s\nEOF", gateway.ID, subnet.Gateway, subnet.Vlan, subnet.Routes)
		err = hyperExecute(ctx, control, command)
	}
	if err != nil {
		log.Println("Set gateway failed")
		return
	}
	return
}

func (a *SubnetAdmin) Create(ctx context.Context, name, vlan, network, netmask, gateway, start, end, rtype, dns, domain, dhcp string, routes string, cluster, owner int64) (subnet *model.Subnet, err error) {
	memberShip := GetMemberShip(ctx)
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
	err = db.Model(&model.Network{}).Where("vlan = ?", vlanNo).Count(&count).Error
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
		Model:        model.Model{Creater: memberShip.UserID, Owner: owner},
		Name:         name,
		Network:      first.String(),
		Netmask:      netmask,
		Gateway:      gateway,
		Start:        start,
		End:          end,
		NameServer:   dns,
		DomainSearch: domain,
		Dhcp:         dhcp,
		ClusterID:    cluster,
		Vlan:         int64(vlanNo),
		Type:         rtype,
		Routes:       routes,
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
	netlink := &model.Network{Vlan: int64(vlanNo)}
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
	_ = execNetwork(ctx, netlink, subnet, owner)
	return
}

func execNetwork(ctx context.Context, netlink *model.Network, subnet *model.Subnet, owner int64) (err error) {
	if subnet.Dhcp == "no" {
		return
	}
	if netlink.Hyper < 0 {
		var dhcp1 *model.Interface
		dhcp1, err = CreateInterface(ctx, subnet.ID, netlink.ID, owner, -1, "", "", "dhcp-1", "dhcp", nil)
		if err != nil {
			log.Println("Failed to allocate dhcp first address", err)
			return
		}
		control := fmt.Sprintf("inter=")
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_net.sh '%d' '%s' '%s' '%s' '%s' '%d' 'FIRST' '%s' '%s'", netlink.Vlan, subnet.Network, subnet.Netmask, subnet.Gateway, dhcp1.Address.Address, subnet.ID, subnet.NameServer, subnet.DomainSearch)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Failed to create first dhcp", err)
			return
		}
	}
	if netlink.Peer < 0 {
		var dhcp2 *model.Interface
		dhcp2, err = CreateInterface(ctx, subnet.ID, netlink.ID, owner, -1, "", "", "dhcp-2", "dhcp", nil)
		if err != nil {
			log.Println("Failed to allocate dhcp first address", err)
			return
		}
		control := fmt.Sprintf("inter=")
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_net.sh '%d' '%s' '%s' '%s' '%s' '%d' 'SECOND' '%s' '%s'", netlink.Vlan, subnet.Network, subnet.Netmask, subnet.Gateway, dhcp2.Address.Address, subnet.ID, subnet.NameServer, subnet.DomainSearch)
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
	ctx = saveTXtoCtx(ctx, db)
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
	err = db.Model(&model.Interface{}).Where("subnet = ? and type <> ?", subnet.ID, "dhcp").Count(&count).Error
	if err != nil {
		log.Println("Failed to query interfaces", err)
		return
	}
	if count > 0 {
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
	if netlink != nil {
		err = DeleteInterfaces(ctx, netlink.ID, subnet.ID, "dhcp")
		if err != nil {
			log.Println("Failed to delete dhcp interfaces")
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
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_net.sh '%d' '%s' '%d'", netlink.Vlan, subnet.Network, subnet.ID)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete interface failed")
			return
		}
	}
	if count <= 1 && netlink != nil {
		err = db.Delete(netlink).Error
		if err != nil {
			log.Println("Failed to delete network")
			return
		}
	}
	return
}

func (a *SubnetAdmin) List(ctx context.Context, offset, limit int64, order, query, sql string) (total int64, subnets []*model.Subnet, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}
	where := ""
	wm := memberShip.GetWhere()
	if wm != "" {
		where = fmt.Sprintf("type != 'internal' or %s", wm)
	}
	subnets = []*model.Subnet{}
	if err = db.Model(&model.Subnet{}).Where(where).Where(query).Where(sql).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Netlink").Where(where).Where(query).Where(sql).Find(&subnets).Error; err != nil {
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, subnet := range subnets {
			subnet.OwnerInfo = &model.Organization{Model: model.Model{ID: subnet.Owner}}
			if err = db.Take(subnet.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				return
			}
		}
	}

	return
}

func (v *SubnetView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit <= 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, subnets, err := subnetAdmin.List(c.Req.Context(), offset, limit, order, query, "")
	if err != nil {
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["Subnets"] = subnets
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"subnets": subnets,
			"total":   total,
			"pages":   pages,
			"query":   query,
		})
		return
	}
	c.HTML(200, "subnets")
}

func (v *SubnetView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	subnetID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "subnets", int64(subnetID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = subnetAdmin.Delete(c.Req.Context(), int64(subnetID))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "subnets",
	})
	return
}

func (v *SubnetView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "subnets_new")
}

func (v *SubnetView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	id := c.ParamsInt64("id")
	if id <= 0 {
		c.Data["ErrorMsg"] = "Id <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Reader, "subnets", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	subnet := &model.Subnet{Model: model.Model{ID: id}}
	err = DB().Take(subnet).Error
	if err != nil {
		log.Println("Database failed to query subnet", err)
		return
	}
	routes := []*StaticRoute{}
	err = json.Unmarshal([]byte(subnet.Routes), &routes)
	if err != nil || len(routes) == 0 {
		log.Println("Failed to unmarshal routes", err)
		subnet.Routes = ""
	} else {
		for i, route := range routes {
			if i == 0 {
				subnet.Routes = fmt.Sprintf("%s:%s", route.Destination, route.Nexthop)
			} else {
				subnet.Routes = fmt.Sprintf("%s %s:%s", subnet.Routes, route.Destination, route.Nexthop)
			}
		}
	}
	subnet.Gateway = strings.Split(subnet.Gateway, "/")[0]
	c.Data["Subnet"] = subnet
	c.HTML(200, "subnets_patch")
}

func (v *SubnetView) checkRoutes(network, netmask, gateway, start, end, dns, routes string, id int64) (routeJson string, err error) {
	if id > 0 {
		db := DB()
		subnet := &model.Subnet{Model: model.Model{ID: id}}
		err = db.Take(subnet).Error
		if err != nil {
			log.Println("DB failed to query subnet ", err)
			return
		}
		network = subnet.Network
		netmask = subnet.Netmask
	}
	inNet := &net.IPNet{
		IP:   net.ParseIP(network),
		Mask: net.IPMask(net.ParseIP(netmask).To4()),
	}
	if gateway != "" && !inNet.Contains(net.ParseIP(gateway)) {
		log.Println("Gateway not belonging to network/netmask")
		err = fmt.Errorf("Gateway not belonging to network/netmask")
		return
	}
	if start != "" && !inNet.Contains(net.ParseIP(start)) {
		log.Println("Start not belonging to network/netmask")
		err = fmt.Errorf("Start not belonging to network/netmask")
		return
	}
	if end != "" && !inNet.Contains(net.ParseIP(end)) {
		log.Println("End not belonging to network/netmask")
		err = fmt.Errorf("End not belonging to network/netmask")
		return
	}
	if dns != "" && net.ParseIP(dns) == nil {
		log.Println("Name server is not an valid IP address")
		err = fmt.Errorf("Name server is not an valid IP address")
		return
	}
	sRoutes := []*StaticRoute{}
	if routes != "" {
		routeList := strings.Split(routes, " ")
		for _, route := range routeList {
			pair := strings.Split(route, ":")
			if len(pair) != 2 {
				log.Println("No valid pair delimiter")
				err = fmt.Errorf("No valid pair delimiter")
				return
			}
			ipmask := pair[0]
			if !strings.Contains(ipmask, "/") {
				log.Println("IPmask has no slash")
				err = fmt.Errorf("IPmask has no slash")
				return
			}
			_, _, err = net.ParseCIDR(ipmask)
			if err != nil {
				log.Println("Failed to parse cidr")
				err = fmt.Errorf("Failed to parse cidr")
				return
			}
			nexthop := pair[1]
			if !inNet.Contains(net.ParseIP(nexthop)) {
				log.Println("Nexthop not belonging to network/netmask")
				err = fmt.Errorf("Nexthop not belonging to network/netmask")
				return
			}
			netrt := &StaticRoute{
				Destination: ipmask,
				Nexthop:     nexthop,
			}
			sRoutes = append(sRoutes, netrt)
		}
	}
	jsonData, err := json.Marshal(sRoutes)
	if err == nil {
		routeJson = string(jsonData)
	}
	return
}

func (v *SubnetView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../subnets"
	id := c.ParamsInt64("id")
	if id <= 0 {
		c.Data["ErrorMsg"] = "Id <= 0"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "subnets", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	name := c.QueryTrim("name")
	network := c.QueryTrim("network")
	netmask := c.QueryTrim("netmask")
	gateway := c.QueryTrim("gateway")
	start := c.QueryTrim("start")
	end := c.QueryTrim("end")
	routes := c.QueryTrim("routes")
	routeJson, err := v.checkRoutes(network, netmask, gateway, start, end, "", routes, id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	subnet, err := subnetAdmin.Update(c.Req.Context(), id, name, gateway, start, end, routeJson)
	if err != nil {
		log.Println("Create subnet failed", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, subnet)
		return
	}
	c.Redirect(redirectTo)
}

func (v *SubnetView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../subnets"
	name := c.QueryTrim("name")
	vlan := c.QueryTrim("vlan")
	rtype := c.QueryTrim("rtype")
	network := c.QueryTrim("network")
	netmask := c.QueryTrim("netmask")
	gateway := c.QueryTrim("gateway")
	routes := c.QueryTrim("routes")
	start := c.QueryTrim("start")
	end := c.QueryTrim("end")
	dns := c.QueryTrim("dns")
	domain := c.QueryTrim("domain")
	dhcp := c.QueryTrim("dhcp")
	if dhcp != "no" {
		dhcp = "yes"
	}
	routeJson, err := v.checkRoutes(network, netmask, gateway, start, end, dns, routes, 0)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	subnet, err := subnetAdmin.Create(c.Req.Context(), name, vlan, network, netmask, gateway, start, end, rtype, dns, domain, dhcp, routeJson, 0, memberShip.OrgID)
	if err != nil {
		log.Println("Create subnet failed, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, subnet)
		return
	}
	c.Redirect(redirectTo)
}
