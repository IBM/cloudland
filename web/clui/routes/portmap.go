/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	portmapAdmin = &PortmapAdmin{}
	portmapView  = &PortmapView{}
	remoteMin    = 18000
	remoteMax    = 20000
)

type PortmapAdmin struct{}
type PortmapView struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
	return
}

func (a *PortmapAdmin) Create(ctx context.Context, instID int64, port int) (portmap *model.Portmap, err error) {
	db := DB()
	instance := &model.Instance{Model: model.Model{ID: instID}}
	err = db.Set("gorm:auto_preload", true).Preload("Interfaces", "primary_if = ?", true).Model(instance).Take(instance).Error
	if err != nil {
		log.Println("DB failed to query instance, %v", err)
		return
	}
	iface := instance.Interfaces[0]
	if iface.Address.Subnet.Router == 0 {
		err = fmt.Errorf("Portmap can not be created without a gateway")
		log.Println("Portmap can not be created without a gateway")
		return
	}
	gateway := &model.Gateway{Model: model.Model{ID: iface.Address.Subnet.Router}}
	err = db.Model(gateway).Set("gorm:auto_preload", true).Take(gateway).Error
	if err != nil {
		log.Println("DB failed to query gateway", err)
		return
	}
	count := 1
	rport := 0
	for count > 0 {
		rport = rand.Intn(remoteMax-remoteMin) + remoteMin
		if err = db.Model(&model.Portmap{}).Where("remote_port = ?", rport).Count(&count).Error; err != nil {
			log.Println("Failed to query existing remote port", err)
			return
		}
	}
	control := fmt.Sprintf("toall=router-%d:%d,%d", gateway.ID, gateway.Hyper, gateway.Peer)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_portmap.sh %d %s %d %d", gateway.ID, iface.Address.Address, port, rport)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Create portmap failed", err)
		return
	}
	name := fmt.Sprintf("%s-%d-%d", instance.Hostname, instance.ID, port)
	portmap = &model.Portmap{GatewayID: gateway.ID, InstanceID: instance.ID, Name: name, Status: "pending", LocalAddress: iface.Address.Address, LocalPort: int32(port), RemotePort: int32(rport)}
	err = db.Create(portmap).Error
	if err != nil {
		log.Println("DB failed to create port map", err)
		return
	}
	return
}

func (a *PortmapAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	portmap := &model.Portmap{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Find(portmap).Error; err != nil {
		log.Println("Failed to query port map", err)
		return
	}
	if portmap.Gateway != nil {
		control := fmt.Sprintf("toall=router-%d:%d,%d", portmap.Gateway.ID, portmap.Gateway.Hyper, portmap.Gateway.Peer)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_portmap.sh %d %s %d %d", portmap.Gateway.ID, portmap.LocalAddress, portmap.LocalPort, portmap.RemotePort)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			log.Println("Delete portmap failed", err)
			return
		}
	}
	err = db.Delete(portmap).Error
	if err != nil {
		log.Println("DB failed to delete port map", err)
		return
	}
	return
}

func (a *PortmapAdmin) List(offset, limit int64, order string) (total int64, portmaps []*model.Portmap, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	portmaps = []*model.Portmap{}
	if err = db.Model(&model.Portmap{}).Count(&total).Error; err != nil {
		log.Println("DB failed to count portmap(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&portmaps).Error; err != nil {
		log.Println("DB failed to query portmap(s), %v", err)
		return
	}

	return
}

func (v *PortmapView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, portmaps, err := portmapAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list portmap(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["Portmaps"] = portmaps
	c.Data["Total"] = total
	c.HTML(200, "portmaps")
}

func (v *PortmapView) Delete(c *macaron.Context, store session.Store) (err error) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	portmapID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid portmap ID", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = portmapAdmin.Delete(c.Req.Context(), int64(portmapID))
	if err != nil {
		log.Println("Failed to delete portmap", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "portmaps",
	})
	return
}

func (v *PortmapView) New(c *macaron.Context, store session.Store) {
	db := DB()
	instances := []*model.Instance{}
	if err := db.Preload("Interfaces", "primary_if = ?", true).Preload("Interfaces.Address").Find(&instances).Error; err != nil {
		return
	}
	c.Data["Instances"] = instances
	c.HTML(200, "portmaps_new")
}

func (v *PortmapView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../portmaps"
	instance := c.Query("instance")
	port := c.Query("port")
	instID, err := strconv.Atoi(instance)
	if err != nil {
		log.Println("Invalid interface ID", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	portNo, err := strconv.Atoi(port)
	if err != nil {
		log.Println("Invalid port number", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	_, err = portmapAdmin.Create(c.Req.Context(), int64(instID), portNo)
	if err != nil {
		log.Println("Failed to create port map", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
