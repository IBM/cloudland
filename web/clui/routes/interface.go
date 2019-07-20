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

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/go-macaron/session"
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
	secgroups := []*model.SecurityGroup{}
	where := ""
	for i, sg := range iface.Secgroups {
		if i == 0 {
			where = fmt.Sprintf("id != %d", sg.ID)
		} else {
			where = fmt.Sprintf("%s and id != %d", where, sg.ID)
		}
	}
	if err := db.Where(where).Find(&secgroups).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Interface"] = iface
	c.Data["Secgroups"] = secgroups
	c.HTML(200, "interfaces_patch")
}

func (v *InterfaceView) Patch(c *macaron.Context, store session.Store) {
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
