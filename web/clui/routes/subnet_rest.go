/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	cidrFun "github.com/apparentlymart/go-cidr/cidr"
	"github.com/go-openapi/strfmt"
	"github.com/jinzhu/gorm"
	macaron "gopkg.in/macaron.v1"
)

var (
	subnetInstance = &SubnetRest{}
)

type SubnetRest struct{}
type subnetRestAdmin struct{ *model.Subnet }

//ListSubnets : list subnets
func (v *SubnetRest) ListSubnets(c *macaron.Context) {
	// TODO: list oid, need to update subnetadmin function
	_, _, err := ChecKPermissionWithErrorResp(model.Reader, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
	offset := c.QueryInt64("marker")
	limit := c.QueryInt64("limit")
	reverse := c.QueryBool("page_reverse")
	order := "created_at"
	if reverse {
		order = "-created_at"
	}
	_, subnets, err := subnetAdmin.List(c.Req.Context(), offset, limit, order, "", "")
	if err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("List subnets fail", err.Error(), code))
		return
	}
	subnetItems := restModels.Subnets{}
	for _, subnet := range subnets {
		// process empty network
		if subnet.Network == "" {
			continue
		}
		creatAt, _ := strfmt.ParseDateTime(subnet.CreatedAt.Format(time.RFC3339))
		updateAt, _ := strfmt.ParseDateTime(subnet.UpdatedAt.Format(time.RFC3339))
		gateSub := strings.Split(subnet.Gateway, `/`)
		network := fmt.Sprintf(`%s/%s`, subnet.Network, gateSub[1])
		subnetItem := &restModels.Subnet{
			Cidr:           network,
			CreatedAt:      creatAt,
			EnableDhcp:     true,
			GatewayIP:      strfmt.IPv4(gateSub[0]),
			ID:             subnet.UUID,
			IPVersion:      4,
			Name:           subnet.Name,
			NetworkID:      subnet.Netlink.UUID,
			ProjectID:      "default",
			RevisionNumber: 0,
			TenantID:       "default",
			UpdatedAt:      updateAt,
		}
		subnetItems = append(subnetItems, subnetItem)
	}
	subnetsResponse := &restModels.ListSubnetsOKBody{
		Subnets: subnetItems,
	}
	c.JSON(200, subnetsResponse)
}

//CreateSubnet : create subnet in db with network id
func (v *SubnetRest) CreateSubnet(c *macaron.Context) {
	uid, oid, err := ChecKPermissionWithErrorResp(model.Writer, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
	db := DB()
	body, _ := c.Req.Body().Bytes()
	if err := JsonSchemeCheck(`subnet.json`, body); err != nil {
		c.JSON(err.Code, ResponseError{ErrorMsg: *err})
		return
	}
	requestData := &restModels.CreateSubnetParamsBody{}
	if err := json.Unmarshal(body, requestData); err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("Unmarshal fail", err.Error(), code))
		return
	}
	_, cidr, err := net.ParseCIDR(requestData.Subnet.Cidr)
	network, end := cidrFun.AddressRange(cidr)
	if err != nil {
		code := http.StatusBadRequest
		c.JSON(code, NewResponseError("parse Cidr fail", err.Error(), code))
		return
	}
	var gateway net.IP
	if requestData.Subnet.GatewayIP != "" {
		gateway = net.ParseIP(requestData.Subnet.GatewayIP)
		if !cidr.Contains(gateway) {
			code := http.StatusBadRequest
			c.JSON(code, NewResponseError("gatway not in subnet range", err.Error(), code))
			return
		}
	} else {
		gateway = cidrFun.Dec(end)
	}
	subnetName := requestData.Subnet.Name
	networkUUID := requestData.Subnet.NetworkID
	first := cidrFun.Inc(network)
	last := cidrFun.Dec(gateway)
	netmask := net.IP(cidr.Mask).String()
	netmaskSize, _ := cidr.Mask.Size()
	gatewayStr := fmt.Sprintf("%s/%d", gateway.String(), netmaskSize)
	networkInstance := &model.Network{Model: model.Model{UUID: networkUUID, Owner: oid}}
	if err = db.Where(networkInstance).Find(networkInstance).Error; err != nil {
		code := http.StatusInternalServerError
		if gorm.IsRecordNotFoundError(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, NewResponseError(fmt.Sprintf("Invalid network id: %s", networkUUID), err.Error(), code))
		return
	}
	//check the subnet is unique
	subnets := []*model.Subnet{}
	db.Where("vlan = ?", networkInstance.Vlan).Find(&subnets)
	for _, sub := range subnets {
		//check subnet  is unique
		_, subNet, _ := net.ParseCIDR(sub.Gateway)
		if cidr.Contains(net.ParseIP(sub.Start)) ||
			cidr.Contains(net.ParseIP(sub.End)) ||
			subNet.Contains(first) ||
			subNet.Contains(last) {
			code := http.StatusConflict
			errMsg := fmt.Sprintf("ip address overlap with subnet: %s", sub.UUID)
			c.JSON(code, NewResponseError("create subnet fail", errMsg, code))
			return
		}
	}
	subnetInstance := subnetRestAdmin{
		Subnet: &model.Subnet{
			Model:   model.Model{Creater: uid, Owner: oid},
			Name:    subnetName,
			Network: network.String(),
			Netmask: netmask,
			Gateway: gatewayStr,
			Start:   first.String(),
			End:     last.String(),
			Vlan:    networkInstance.Vlan,
			Type:    "ipv4",
		},
	}
	err = subnetInstance.createSubnet(c.Req.Context(), networkInstance, netmaskSize)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to create subnet with error: %v", err))
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("create subnet fail", err.Error(), code))
		return
	}
	creatAt, _ := strfmt.ParseDateTime(subnetInstance.CreatedAt.Format(time.RFC3339))
	updateAt, _ := strfmt.ParseDateTime(subnetInstance.UpdatedAt.Format(time.RFC3339))
	subnetRespons := &restModels.CreateSubnetOKBody{
		Subnet: &restModels.Subnet{
			Cidr:           subnetInstance.Network,
			CreatedAt:      creatAt,
			EnableDhcp:     true,
			GatewayIP:      strfmt.IPv4(subnetInstance.Gateway),
			ID:             subnetInstance.UUID,
			IPVersion:      4,
			Name:           subnetInstance.Name,
			NetworkID:      networkInstance.UUID,
			ProjectID:      "default",
			RevisionNumber: 0,
			TenantID:       "default",
			UpdatedAt:      updateAt,
		},
	}
	c.JSON(200, subnetRespons)
	return
}

//DeleteSubnet : delete subNet
func (v *SubnetRest) DeleteSubnet(c *macaron.Context) {
	var err error
	var oid int64
	_, oid, err = ChecKPermissionWithErrorResp(model.Writer, c)
	if err != nil {
		log.Print(err.Error())
		return
	}
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	subnetUUID := c.Params("id")
	if subnetUUID == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	//check subnet is existing
	subnet := &model.Subnet{Model: model.Model{UUID: subnetUUID, Owner: oid}}
	if err = db.Where(subnet).Take(subnet).Error; err != nil {
		code := http.StatusInternalServerError
		if gorm.IsRecordNotFoundError(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, NewResponseError("Delete subnet fail", err.Error(), code))
		return
	}
	// check no  gw-router was attached
	if subnet.Router > 0 {
		err = fmt.Errorf("Subnet belongs to a gateway")
		log.Println("Subnet belongs to a gateway", err)
		code := http.StatusConflict
		c.JSON(code, NewResponseError("Delete subnet fail", err.Error(), code))
		return
	}
	// check all of ipaddress in this subnet is idle status
	if isUsed, err := v.checkIPaddresIsUsed(subnet.ID); err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("Delete subnet fail", err.Error(), code))
		return
	} else if isUsed {
		code := http.StatusConflict
		errMsg := fmt.Sprintf("Failed to delete subnet: %s, ipaddress in subnet is used", subnetUUID)
		log.Println(errMsg)
		c.JSON(code, NewResponseError("Delete subnet fail", errMsg, code))
		return
	}
	network := subnet.Netlink
	control := ""
	var skipTag bool
	if network.Hyper >= 0 {
		control = fmt.Sprintf("toall=vlan-%d:%d", subnet.Vlan, network.Hyper)
		if network.Peer >= 0 {
			control = fmt.Sprintf("%s,%d", control, network.Peer)
		}
	} else if network.Peer >= 0 {
		control = fmt.Sprintf("inter=%d", network.Peer)
	} else {
		log.Println("Network has no valid hypers")
		skipTag = true
	}
	if skipTag {
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_net.sh '%d' '%s' '%d'", network.Vlan, subnet.Network, subnet.ID)
		err = hyperExecute(c.Req.Context(), control, command)
		if err != nil {
			log.Println("Delete interface failed")
			code := http.StatusInternalServerError
			c.JSON(code, NewResponseError("Delete subnet fail", err.Error(), code))
			return
		}
	}
	if err = db.Delete(subnet).Error; err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("Delete subnet fail", err.Error(), code))
		return
	}
	// start delete ip address
	if err = db.Where("subnet_id = ?", subnet.ID).Delete(model.Address{}).Error; err != nil {
		log.Println("Database delete ip address failed, %v", err)
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("Delete subnet fail", err.Error(), code))
		return
	}
	// delete dhcp interface
	if err = db.Preload("Addresses", "subnet_id = ?", subnet.ID).Where("type = ?", "dhcp").Delete(model.Interface{}).Error; err != nil {
		code := http.StatusInternalServerError
		c.JSON(code, NewResponseError("Delete internal interface fail", err.Error(), code))
		return
	}
	c.Status(http.StatusNoContent)
	return
}

func (v *SubnetRest) checkIPaddresIsUsed(id int64) (isUsed bool, err error) {
	count := 0
	db := DB()
	err = db.Model(&model.Address{}).Where("subnet_id = ? and allocated = ?", id, true).Count(&count).Error
	if err != nil {
		log.Println("Database delete addresses failed, %v", err)
		return
	}
	if count == 0 {
		return false, nil
	}
	iface := &model.Interface{}
	//check interface whether reserved by DHCP
	err = db.Model(iface).Preload("Addresses", "subnet_id = ?", id).Where("type <> ?", "dhcp").Count(&count).Error
	if err != nil {
		log.Println(err.Error())
		return
	}
	if count != 0 {
		return true, nil
	}
	return
}

func (v *subnetRestAdmin) createSubnet(ctx context.Context, network *model.Network, preSize int) (err error) {
	ctx, db := getCtxDB(ctx)
	tx := db.Begin()
	ctx = saveTXtoCtx(ctx, tx)
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			log.Print("calling back")
			tx.Rollback()
		}
	}()
	if v.Subnet == nil {
		return fmt.Errorf("empty subnet instance")
	}
	err = tx.Create(v.Subnet).Error
	if err != nil {
		log.Println("Database create subnet failed, %v", err)
		return
	}
	ip := net.ParseIP(v.Subnet.Start)
	for {
		ipstr := fmt.Sprintf("%s/%d", ip.String(), preSize)
		if ipstr == v.Subnet.Gateway {
			ip = cidrFun.Inc(ip)
		}
		address := &model.Address{
			Model: model.Model{
				Creater: v.Subnet.Creater,
				Owner:   v.Subnet.Owner,
			},
			Address:  ipstr,
			Netmask:  v.Subnet.Netmask,
			Type:     "ipv4",
			SubnetID: v.Subnet.ID,
		}
		err = tx.Create(address).Error
		if err != nil {
			log.Println("Database create address failed, %v", err)
		}
		if ip.String() == v.Subnet.End {
			break
		}
		ip = cidrFun.Inc(ip)
	}
	err = execNetwork(ctx, network, v.Subnet, v.Subnet.Owner)
	if err != nil {
		log.Println("Failed remote execute network creation", err)
		log.Print("Start calling back")
		return
	}
	return
}
