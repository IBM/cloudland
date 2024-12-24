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
	"net/http"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	interfaceAdmin = &InterfaceAdmin{}
	interfaceView  = &InterfaceView{}
)

type InterfaceInfo struct {
	Subnet         *model.Subnet
	MacAddress     string
	IpAddress      string
	SecurityGroups []*model.SecurityGroup
}

type InterfaceAdmin struct{}

type InterfaceView struct{}

func (a *InterfaceAdmin) Update(ctx context.Context, id int64, name, pairs string, sgIDs []int64) (iface *model.Interface, err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	iface = &model.Interface{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Take(iface).Error; err != nil {
		log.Println("Failed to query interface ", err)
		return
	}
	if iface.Name != name {
		iface.Name = name
		if err = db.Model(iface).Updates(iface).Error; err != nil {
			log.Println("Failed to save interface", err)
			return
		}
	}
	if iface.AddrPairs != pairs {
		iface.AddrPairs = pairs
		if err = db.Model(iface).Updates(iface).Error; err != nil {
			log.Println("Failed to save interface", err)
			return
		}
	}
	control := fmt.Sprintf("inter=%d", iface.Hyper)
	if iface.Hyper < 0 {
		instance := &model.Instance{Model: model.Model{ID: iface.Instance}}
		if err = db.Take(instance).Error; err != nil {
			log.Println("Failed to query instance ", err)
			return
		}
		control = fmt.Sprintf("inter=%d", instance.Hyper)
	}
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/allow_as_addr.sh '%s' '%s' <<EOF\n%s\nEOF", iface.Address.Address, iface.MacAddr, pairs)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Launch vm command execution failed", err)
		return
	}
	sgChanged := false
	for _, esg := range iface.SecurityGroups {
		found := false
		for _, sgID := range sgIDs {
			if sgID == esg.ID {
				found = true
				break
			}
		}
		if found == false {
			sgChanged = true
			break
		}
	}
	if !sgChanged {
		for _, sgID := range sgIDs {
			found := false
			for _, esg := range iface.SecurityGroups {
				if sgID == esg.ID {
					found = true
					break
				}
			}
			if found == false {
				sgChanged = true
				break
			}
		}
	}
	log.Println("$$$$ start to sgChanged = ", sgChanged, name)
	if sgChanged == true {
		log.Println("$$$$ start to change security group")
		secgroups := []*model.SecurityGroup{}
		if err = db.Where(sgIDs).Find(&secgroups).Error; err != nil {
			log.Println("Security group query failed", err)
			return
		}
		db.Model(iface).Association("SecurityGroups").Clear()
		iface.SecurityGroups = secgroups
		if err = db.Model(iface).Updates(iface).Error; err != nil {
			log.Println("Failed to save security groups", err)
			return
		}
		var secRules []*model.SecurityRule
		secRules, err = model.GetSecurityRules(secgroups)
		if err != nil {
			log.Println("Failed to get security rules", err)
			return
		}
		securityData := []*SecurityData{}
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
		var jsonData []byte
		jsonData, err = json.Marshal(securityData)
		if err != nil {
			log.Println("Failed to marshal security json data, %v", err)
			return
		}
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/reapply_secgroup.sh '%s' '%s' <<EOF\n%s\nEOF", iface.Address.Address, iface.MacAddr, jsonData)
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
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	ifaceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", int64(ifaceID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	iface := &model.Interface{Model: model.Model{ID: int64(ifaceID)}}
	if err = db.Set("gorm:auto_preload", true).Take(iface).Error; err != nil {
		log.Println("Image query failed", err)
		return
	}
	_, secgroups, err := secgroupAdmin.List(c.Req.Context(), 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Interface"] = iface
	c.Data["Secgroups"] = secgroups
	c.HTML(200, "interfaces_patch")
}

func (v *InterfaceView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	subnetID := c.QueryInt64("subnet")
	subnet, err := subnetAdmin.Get(ctx, subnetID)
	if err != nil {
		log.Println("Get subnet failed", err)
		c.HTML(http.StatusBadRequest, err.Error())
		return
	}
	instID := c.QueryInt64("instance")
	if instID > 0 {
		permit, _ := memberShip.CheckOwner(model.Writer, "instances", int64(instID))
		if !permit {
			log.Println("Not authorized to access instance")
			c.Data["ErrorMsg"] = "Not authorized to access instance"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	address := c.QueryTrim("address")
	mac := c.QueryTrim("mac")
	ifname := c.QueryTrim("ifname")
	sgList := c.QueryTrim("secgroups")
	var sgIDs []int64
	if sgList != "" {
		sg := strings.Split(sgList, ",")
		for i := 0; i < len(sg); i++ {
			sgID, err := strconv.Atoi(sg[i])
			if err != nil {
				log.Println("Invalid security group ID", err)
				continue
			}
			permit, _ := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
			if !permit {
				log.Println("Not authorized to access security group")
				c.Data["ErrorMsg"] = "Not authorized to access security group"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			sgIDs = append(sgIDs, int64(sgID))
		}
	} else {
		sgID := store.Get("defsg").(int64)
		permit, _ := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
		if !permit {
			log.Println("Not authorized to access security group")
			c.Data["ErrorMsg"] = "Not authorized to access security group"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		sgIDs = append(sgIDs, sgID)
	}
	secgroups := []*model.SecurityGroup{}
	if err = DB().Where(sgIDs).Find(&secgroups).Error; err != nil {
		log.Println("Security group query failed", err)
		return
	}
	iface, err := CreateInterface(ctx, subnet, instID, memberShip.OrgID, -1, address, mac, ifname, "instance", secgroups)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})
	}
	c.JSON(200, iface)
}

func (v *InterfaceView) Delete(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	iface := &model.Interface{Model: model.Model{ID: id}}
	err = DB().Take(iface).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = DeleteInterface(ctx, iface)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, "ok")
}

func (v *InterfaceView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../instances"
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	ifaceID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "interfaces", int64(ifaceID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	name := c.QueryTrim("name")
	secgroups := c.QueryStrings("secgroups")
	pairs := c.QueryTrim("pairs")
	var sgIDs []int64
	log.Println("$$$$$$ len = ", len(secgroups))
	if len(secgroups) > 0 {
		for _, s := range secgroups {
			sID, err := strconv.Atoi(s)
			if err != nil {
				log.Println("Invalid security group ID", err)
				continue
			}
			permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sID))
			if !permit {
				log.Println("Not authorized for this operation")
				c.Data["ErrorMsg"] = "Not authorized for this operation"
				c.HTML(http.StatusBadRequest, "error")
				return
			}
			sgIDs = append(sgIDs, int64(sID))
		}
	} else {
		sID := store.Get("defsg").(int64)
		permit, err = memberShip.CheckOwner(model.Writer, "security_groups", int64(sID))
		if !permit {
			log.Println("Not authorized for this operation")
			c.Data["ErrorMsg"] = "Not authorized for this operation"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		sgIDs = append(sgIDs, sID)
	}
	iface, err := interfaceAdmin.Update(c.Req.Context(), int64(ifaceID), name, pairs, sgIDs)
	if err != nil {
		log.Println("Failed to update interface", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, iface)
		return
	}
	c.Redirect(redirectTo)
}
