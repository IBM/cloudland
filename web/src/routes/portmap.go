/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

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
	memberShip := GetMemberShip(ctx)
	db := DB()
	instance := &model.Instance{Model: model.Model{ID: instID}}
	err = db.Set("gorm:auto_preload", true).Preload("Interfaces", "primary_if = ?", true).Model(instance).Take(instance).Error
	if err != nil {
		logger.Debug("DB failed to query instance, %v", err)
		return
	}
	iface := instance.Interfaces[0]
	if iface.Address.Subnet.RouterID == 0 {
		err = fmt.Errorf("Portmap can not be created without a router")
		logger.Debug("Portmap can not be created without a router")
		return
	}
	router := &model.Router{Model: model.Model{ID: iface.Address.Subnet.RouterID}}
	err = db.Model(router).Set("gorm:auto_preload", true).Take(router).Error
	if err != nil {
		logger.Debug("DB failed to query router", err)
		return
	}
	count := 1
	rport := 0
	for count > 0 {
		rport = rand.Intn(remoteMax-remoteMin) + remoteMin
		if err = db.Model(&model.Portmap{}).Where("remote_port = ?", rport).Count(&count).Error; err != nil {
			logger.Debug("Failed to query existing remote port", err)
			return
		}
	}
	/*
		control := fmt.Sprintf("toall=router-%d:%d,%d", router.ID, router.Hyper, router.Peer)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/create_portmap.sh '%d' '%s' '%d' '%d'", router.ID, iface.Address.Address, port, rport)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			logger.Debug("Create portmap failed", err)
			return
		}*/
	name := fmt.Sprintf("%s-%d-%d", instance.Hostname, instance.ID, port)
	portmap = &model.Portmap{Model: model.Model{Creater: memberShip.UserID}, Owner: memberShip.OrgID, RouterID: router.ID, InstanceID: instance.ID, Name: name, Status: "pending", LocalAddress: iface.Address.Address, LocalPort: int32(port), RemotePort: int32(rport)}
	err = db.Create(portmap).Error
	if err != nil {
		logger.Debug("DB failed to create port map", err)
		return
	}
	return
}

func (a *PortmapAdmin) Delete(ctx context.Context, id int64) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	portmap := &model.Portmap{Model: model.Model{ID: id}}
	if err = db.Set("gorm:auto_preload", true).Find(portmap).Error; err != nil {
		logger.Debug("Failed to query port map", err)
		return
	}
	if portmap.Router != nil {
		control := fmt.Sprintf("toall=router-%d:%d,%d", portmap.Router.ID, portmap.Router.Hyper, portmap.Router.Peer)
		command := fmt.Sprintf("/opt/cloudland/scripts/backend/clear_portmap.sh '%d' '%s' '%d' '%d'", portmap.Router.ID, portmap.LocalAddress, portmap.LocalPort, portmap.RemotePort)
		err = hyperExecute(ctx, control, command)
		if err != nil {
			logger.Debug("Delete portmap failed", err)
			return
		}
	}
	err = db.Delete(portmap).Error
	if err != nil {
		logger.Debug("DB failed to delete port map", err)
		return
	}
	return
}

func (a *PortmapAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, portmaps []*model.Portmap, err error) {
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
	where := memberShip.GetWhere()
	portmaps = []*model.Portmap{}
	if err = db.Model(&model.Portmap{}).Where(where).Where(query).Count(&total).Error; err != nil {
		logger.Debug("DB failed to count portmap(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Where(query).Find(&portmaps).Error; err != nil {
		logger.Debug("DB failed to query portmap(s), %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, pmap := range portmaps {
			pmap.OwnerInfo = &model.Organization{Model: model.Model{ID: pmap.Owner}}
			if err = db.Take(pmap.OwnerInfo).Error; err != nil {
				logger.Debug("Failed to query owner info", err)
				return
			}
		}
	}

	return
}

func (v *PortmapView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		logger.Debug("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, portmaps, err := portmapAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		logger.Debug("Failed to list portmap(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["Portmaps"] = portmaps
	c.Data["Total"] = total
	c.Data["Pages"] = GetPages(total, limit)
	c.Data["Query"] = query
	c.HTML(200, "portmaps")
}

func (v *PortmapView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	portmapID, err := strconv.Atoi(id)
	if err != nil {
		logger.Debug("Invalid portmap ID", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "portmaps", int64(portmapID))
	if !permit {
		logger.Debug("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = portmapAdmin.Delete(c.Req.Context(), int64(portmapID))
	if err != nil {
		logger.Debug("Failed to delete portmap", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "portmaps",
	})
	return
}

func (v *PortmapView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Debug("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	instances := []*model.Instance{}
	if err := db.Preload("Interfaces", "primary_if = ?", true).Preload("Interfaces.Address").Find(&instances).Error; err != nil {
		return
	}
	c.Data["Instances"] = instances
	c.HTML(200, "portmaps_new")
}

func (v *PortmapView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Debug("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "../portmaps"
	instance := c.QueryTrim("instance")
	port := c.QueryTrim("port")
	instID, err := strconv.Atoi(instance)
	if err != nil {
		logger.Debug("Invalid interface ID", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err = memberShip.CheckOwner(model.Writer, "instances", int64(instID))
	if !permit {
		logger.Debug("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	portNo, err := strconv.Atoi(port)
	if err != nil {
		logger.Debug("Invalid port number", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	_, err = portmapAdmin.Create(c.Req.Context(), int64(instID), portNo)
	if err != nil {
		logger.Debug("Failed to create port map", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
