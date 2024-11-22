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

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

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

func checkIfExistVni(vni int64) (result bool, err error) {
	db := DB()
	count := 0
	if err = db.Model(&model.Subnet{}).Where("vlan = ?", vni).Count(&count).Error; err != nil {
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
			Model:    model.Model{Creater: subnet.Creater},
			Owner:    subnet.Owner,
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

func (a *SubnetAdmin) Get(ctx context.Context, id int64) (subnet *model.Subnet, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid subnet ID: %d", id)
		log.Println(err)
		return
	}
	memberShip := GetMemberShip(ctx)
	db := DB()
	where := memberShip.GetWhere()
	subnet = &model.Subnet{Model: model.Model{ID: id}}
	err = db.Where(where).Take(subnet).Error
	if err != nil {
		log.Println("DB failed to query subnet ", err)
		return
	}
	if subnet.RouterID > 0 {
		subnet.Router = &model.Router{Model: model.Model{ID: subnet.RouterID}}
		err = db.Take(subnet.Router).Error
		if err != nil {
			log.Println("Failed to query router ", err)
			return
		}
	}
	if subnet.Type != "public" {
		permit := memberShip.ValidateOwner(model.Reader, subnet.Owner)
		if !permit {
			log.Println("Not authorized to read the subnet")
			err = fmt.Errorf("Not authorized")
			return
		}
	}
	return
}

func (a *SubnetAdmin) GetSubnetByUUID(ctx context.Context, uuID string) (subnet *model.Subnet, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	subnet = &model.Subnet{}
	err = db.Where(where).Where("uuid = ?", uuID).Take(subnet).Error
	if err != nil {
		log.Println("Failed to query subnet, %v", err)
		return
	}
	if subnet.RouterID > 0 {
		subnet.Router = &model.Router{Model: model.Model{ID: subnet.RouterID}}
		err = db.Take(subnet.Router).Error
		if err != nil {
			log.Println("Failed to query router ", err)
			return
		}
	}
	if subnet.Type != "public" {
		permit := memberShip.ValidateOwner(model.Reader, subnet.Owner)
		if !permit {
			log.Println("Not authorized to read the subnet")
			err = fmt.Errorf("Not authorized")
			return
		}
	}
	return
}

func (a *SubnetAdmin) GetSubnetByName(ctx context.Context, name string) (subnet *model.Subnet, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	subnet = &model.Subnet{}
	err = db.Where(where).Where("name = ?", name).Take(subnet).Error
	if err != nil {
		log.Println("Failed to query subnet ", err)
		return
	}
	if subnet.RouterID > 0 {
		subnet.Router = &model.Router{Model: model.Model{ID: subnet.RouterID}}
		err = db.Take(subnet.Router).Error
		if err != nil {
			log.Println("Failed to query router ", err)
			return
		}
	}
	if subnet.Type != "public" {
		permit := memberShip.ValidateOwner(model.Reader, subnet.Owner)
		if !permit {
			log.Println("Not authorized to read the subnet")
			err = fmt.Errorf("Not authorized")
			return
		}
	}
	return
}

func (a *SubnetAdmin) GetSubnet(ctx context.Context, reference *BaseReference) (subnet *model.Subnet, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Subnet base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		subnet, err = a.GetSubnetByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		subnet, err = a.GetSubnetByName(ctx, reference.Name)
		return
	}
	return
}

func (a *SubnetAdmin) Update(ctx context.Context, id int64, name, gateway, start, end, dns, routes string) (subnet *model.Subnet, err error) {
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
	if dns != "" {
		subnet.NameServer = dns
	}
	subnet.Routes = routes
	err = db.Save(subnet).Error
	if err != nil {
		log.Println("DB failed to save subnet ", err)
		return
	}
	if subnet.RouterID > 0 {
		err = setRouting(ctx, subnet.RouterID, subnet, false)
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

func clearRouting(ctx context.Context, routerID int64, subnet *model.Subnet) (err error) {
	db := DB()
	router := &model.Router{Model: model.Model{ID: routerID}}
	err = db.Take(router).Error
	if err != nil {
		log.Println("DB failed to query router", err)
		return
	}
	if router.Hyper >= 0 {
		control := fmt.Sprintf("toall=router-%d:%d", router.ID, router.Hyper)
		if router.Peer >= 0 {
			control = fmt.Sprintf("%s,%d", control, router.Peer)
		}
		if router.Hyper == router.Peer {
			control = fmt.Sprintf("inter=%d", router.Hyper)
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_gateway.sh '%d' '%s' '%d'", router.ID, subnet.Gateway, subnet.Vlan)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Set gateway failed")
			return
		}
	}
	return
}

func setRouting(ctx context.Context, routerID int64, subnet *model.Subnet, routeOnly bool) (err error) {
	db := DB()
	router := &model.Router{Model: model.Model{ID: routerID}}
	err = db.Take(router).Error
	if err != nil {
		log.Println("DB failed to query router", err)
		return
	}
	_, err = CreateInterface(ctx, subnet, routerID, router.Owner, router.Hyper, subnet.Gateway, "", "subnet-gw", "gateway", nil)
	if err != nil {
		log.Println("Failed to create gateway subnet interface", err)
		return
	}
	return
}

func (a *SubnetAdmin) Create(ctx context.Context, vlan int, name, network, gateway, start, end, rtype, dns, domain string, dhcp bool, router *model.Router) (subnet *model.Subnet, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		err = fmt.Errorf("Not authorized for this operation")
		return
	}
	owner := memberShip.OrgID
	db := DB()
	if vlan <= 0 {
		vlan, err = getValidVni()
		if err != nil {
			log.Println("Failed to get valid vlan %s, %v", vlan, err)
			return
		}
	}
	count := 0
	err = db.Model(&model.Subnet{}).Where("vlan = ?", vlan).Count(&count).Error
	if err != nil {
		log.Println("Database failed to count network", err)
		return
	}
	var routerID int64
	if router != nil {
		routerID = router.ID
	}
	_, ipNet, err := net.ParseCIDR(network)
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
	preSize, _ := ipNet.Mask.Size()
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
	netmask := net.IP(net.CIDRMask(preSize, 32)).String()
	subnet = &model.Subnet{
		Model:        model.Model{Creater: memberShip.UserID},
		Owner:        owner,
		Name:         name,
		Network:      network,
		Netmask:      netmask,
		Gateway:      gateway,
		Start:        start,
		End:          end,
		NameServer:   dns,
		DomainSearch: domain,
		Dhcp:         dhcp,
		Vlan:         int64(vlan),
		Type:         rtype,
		RouterID:     routerID,
	}
	err = db.Create(subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		return
	}
	ip := net.ParseIP(start)
	for {
		ipstr := fmt.Sprintf("%s/%d", ip.String(), preSize)
		address := &model.Address{Model: model.Model{Creater: memberShip.UserID}, Owner: owner, Address: ipstr, Netmask: netmask, Type: "ipv4", SubnetID: subnet.ID}
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
	// Create record for gateway
	address := &model.Address{Model: model.Model{Creater: memberShip.UserID}, Owner: owner, Address: gateway, Netmask: netmask, Type: "ipv4", SubnetID: subnet.ID}
	err = db.Create(address).Error
	if err != nil {
		log.Println("Database create address for gateway failed, %v", err)
	}
	if subnet.RouterID > 0 {
		err = setRouting(ctx, subnet.RouterID, subnet, false)
		if err != nil {
			log.Println("Failed to set routing for subnet")
			return
		}
	}
	return
}

func (a *SubnetAdmin) Delete(ctx context.Context, subnet *model.Subnet) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	ctx = SaveTXtoCtx(ctx, db)
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, subnet.Owner)
	if !permit {
		log.Println("Not authorized to delete the subnet")
		err = fmt.Errorf("Not authorized")
		return
	}
	count := 0
	err = db.Model(&model.Interface{}).Where("subnet = ? and type <> 'dhcp' and type <> 'gateway'", subnet.ID).Count(&count).Error
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
	err = db.Where("subnet_id = ?", subnet.ID).Delete(model.Address{}).Error
	if err != nil {
		log.Println("Database delete ip address failed, %v", err)
		return
	}
	if subnet.RouterID > 0 {
		err = clearRouting(ctx, subnet.RouterID, subnet)
		if err != nil {
			log.Println("Failed to set routing for subnet")
			return
		}
	}
	return
}

func (a *SubnetAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, subnets []*model.Subnet, err error) {
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
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	if where != "" {
		where = fmt.Sprintf("type = 'public' or %s", where)
	}
	subnets = []*model.Subnet{}
	if err = db.Model(&model.Subnet{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Router").Where(where).Where(query).Find(&subnets).Error; err != nil {
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
	total, subnets, err := subnetAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["Subnets"] = subnets
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.Data["UserID"] = store.Get("uid").(int64)
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
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.Error(http.StatusBadRequest)
		return
	}
	subnetID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	subnet, err := subnetAdmin.Get(ctx, int64(subnetID))
	if err != nil {
		log.Println("Failed to get subnet ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = subnetAdmin.Delete(c.Req.Context(), subnet)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
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
	routers := []*model.Router{}
	err := DB().Find(&routers).Error
	if err != nil {
		log.Println("Database failed to query gateways", err)
		return
	}
	c.Data["Routers"] = routers
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
	dns := c.QueryTrim("dns")
	routes := c.QueryTrim("routes")
	routeJson, err := v.checkRoutes(network, netmask, gateway, start, end, dns, routes, id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	subnet, err := subnetAdmin.Update(c.Req.Context(), id, name, gateway, start, end, dns, routeJson)
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
	ctx := c.Req.Context()
	redirectTo := "../subnets"
	name := c.QueryTrim("name")
	vlan := c.QueryInt("vlan")
	rtype := c.QueryTrim("rtype")
	network := c.QueryTrim("network")
	routerID := c.QueryInt64("router")
	gateway := c.QueryTrim("gateway")
	start := c.QueryTrim("start")
	end := c.QueryTrim("end")
	dns := c.QueryTrim("dns")
	domain := c.QueryTrim("domain")
	dhcpStr := c.QueryTrim("dhcp")
	dhcp := false
	if dhcpStr != "no" {
		dhcp = true
	}
	/*
		routeJson, err := v.checkRoutes(network, netmask, gateway, start, end, dns, routes, 0)
		if err != nil {
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	*/
	var router *model.Router
	var err error
	if routerID > 0 {
		router, err = routerAdmin.Get(ctx, routerID)
		if err != nil {
			log.Println("Get router failed ", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(404, "404")
			return
		}
	}
	_, err = subnetAdmin.Create(ctx, vlan, name, network, gateway, start, end, rtype, dns, domain, dhcp, router)
	if err != nil {
		log.Println("Create subnet failed ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(400, "400")
		return
	}
	c.Redirect(redirectTo)
}
