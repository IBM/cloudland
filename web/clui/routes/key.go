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
	"strconv"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	keyAdmin = &KeyAdmin{}
	keyView  = &KeyView{}
)

type KeyAdmin struct{}
type KeyView struct{}

func (a *KeyAdmin) Create(ctx context.Context, name, pubkey string) (key *model.Key, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	key = &model.Key{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Name: name, PublicKey: pubkey}
	err = db.Create(key).Error
	if err != nil {
		log.Println("DB failed to create key, %v", err)
		return
	}
	return
}

func (a *KeyAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Key{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("DB failed to delete key, %v", err)
		return
	}
	return
}

func (a *KeyAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, keys []*model.Key, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}
	where := memberShip.GetWhere()
	keys = []*model.Key{}
	if err = db.Model(&model.Key{}).Where(where).Where(query).Count(&total).Error; err != nil {
		log.Println("DB failed to count keys, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Where(query).Find(&keys).Error; err != nil {
		log.Println("DB failed to query keys, %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, key := range keys {
			key.OwnerInfo = &model.Organization{Model: model.Model{ID: key.Owner}}
			if err = db.Take(key.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				return
			}
		}
	}

	return
}

func (v *KeyView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 10
	}
	order := c.QueryTrim("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, keys, err := keyAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list keys, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Keys"] = keys
	c.Data["Total"] = total
	c.Data["Pages"] = GetPages(total, limit)
	c.Data["Query"] = query
	c.HTML(200, "keys")
}

func (v *KeyView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	keyID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid key id, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "keys", int64(keyID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = keyAdmin.Delete(int64(keyID))
	if err != nil {
		log.Println("Failed to delete key, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "keys",
	})
	return
}

func (v *KeyView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "keys_new")
}

func (v *KeyView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../keys"
	name := c.QueryTrim("name")
	pubkey := c.QueryTrim("pubkey")
	_, err := keyAdmin.Create(c.Req.Context(), name, pubkey)
	if err != nil {
		log.Println("Failed to create key, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
