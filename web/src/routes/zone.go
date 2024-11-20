/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	zoneAdmin = &ZoneAdmin{}
	zoneView  = &ZoneView{}
)

type ZoneAdmin struct{}
type ZoneView struct{}

func (a *ZoneAdmin) List(offset, limit int64, order, query string) (total int64, zones []*model.Zone, err error) {
	db := dbs.DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "hostid"
	}
	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}

	zones = []*model.Zone{}
	if err = db.Model(&model.Zone{}).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Zone").Where("hostid >= 0").Where(query).Find(&zones).Error; err != nil {
		return
	}
	db = db.Offset(0).Limit(-1)
	return
}

func (a *ZoneAdmin) Get(ctx context.Context, id int64) (zone *model.Zone, err error) {
	db := dbs.DB()
	zone = &model.Zone{ID: id}
	if err = db.Take(zone).Error; err != nil {
		log.Println("Failed to query zone", err)
		return
	}
	return
}

func (a *ZoneAdmin) GetZoneByName(ctx context.Context, name string) (zone *model.Zone, err error) {
	db := dbs.DB()
	zone = &model.Zone{}
	err = db.Where("name = ?", name).Take(zone).Error
	if err != nil {
		log.Println("Failed to query zone, %v", err)
		return
	}
	return
}

func (v *ZoneView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.Query("order")
	if order == "" {
		order = "hostid"
	}
	query := c.QueryTrim("q")
	total, zones, err := zoneAdmin.List(offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["Zones"] = zones
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"zones": zones,
			"total": total,
			"pages": pages,
			"query": query,
		})
		return
	}
	c.HTML(200, "zones")
}
