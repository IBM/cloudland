/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
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

func (a *KeyAdmin) Create(name, pubkey string) (key *model.Key, err error) {
	db := DB()
	key = &model.Key{Name: name, PublicKey: pubkey}
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

func (a *KeyAdmin) List(offset, limit int64, order string) (total int64, keys []*model.Key, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	keys = []*model.Key{}
	if err = db.Model(&model.Key{}).Count(&total).Error; err != nil {
		log.Println("DB failed to count keys, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&keys).Error; err != nil {
		log.Println("DB failed to query keys, %v", err)
		return
	}

	return
}

func (v *KeyView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, keys, err := keyAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list keys, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Keys"] = keys
	c.Data["Total"] = total
	c.HTML(200, "keys")
}

func (v *KeyView) Delete(c *macaron.Context, store session.Store) (err error) {
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
	c.HTML(200, "keys_new")
}

func (v *KeyView) Create(c *macaron.Context, store session.Store) {
	redirectTo := "../keys"
	name := c.Query("name")
	pubkey := c.Query("pubkey")
	_, err := keyAdmin.Create(name, pubkey)
	if err != nil {
		log.Println("Failed to create key, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
