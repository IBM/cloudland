/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
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
	keyTemp = &KeyTemp{}
)

type KeyAdmin struct{}
type KeyView struct{}
type KeyTemp struct{}

func (point *KeyTemp) Create() (publicKey, privateKey string, err error){
	// generate key
	private, er := rsa.GenerateKey(rand.Reader, 1024)
	if er != nil {
		log.Println("failed to create privateKey ")
		err = er
		return
	}
	privateKeyPEM := &pem.Block{Type:"RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)}
	privateKey = string(pem.EncodeToMemory(privateKeyPEM))
	pub, er := ssh.NewPublicKey(&private.PublicKey)
	if er != nil {
		log.Println("failed to create publicKey")
		err = er
		return
	}
	publicKey = string(ssh.MarshalAuthorizedKey(pub))
	return
}

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
		limit = 16
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
				err = nil
				continue
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
	total, keys, err := keyAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list keys, %v", err)
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
	c.Data["Keys"] = keys
	c.Data["Total"] = total
	c.Data["Pages"] = GetPages(total, limit)
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"keys":  keys,
			"total": total,
			"pages": pages,
			"query": query,
		})
		return
	}
	c.HTML(200, "keys")
}

func (v *KeyView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	keyID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid key id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "keys", int64(keyID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = keyAdmin.Delete(int64(keyID))
	if err != nil {
		log.Println("Failed to delete key, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "keys",
	})
	return
}

func (v *KeyView) New(c *macaron.Context, store session.Store)(){
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	hostname := c.QueryTrim("hostname")
	hyper := c.QueryTrim("hyper")
	count := c.QueryTrim("count")
	userData := c.QueryTrim("userData")
	
	if hostname != ""{
		c.Data["InstanceFlag"] = 1
	}
	c.Data["Hostname"] = hostname
	c.Data["hyper"] = hyper
	c.Data["count"] = count
	c.Data["userData"] = userData
	c.HTML(200, "keys_new");
}

func (v *KeyView) Confirm(c *macaron.Context, store session.Store){
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	
	name := c.QueryTrim("name")
	publicKey := c.QueryTrim("PublicKey")
	hostname := c.QueryTrim("host")
	key, err := keyAdmin.Create(c.Req.Context(), name, publicKey)
	if err != nil {
		log.Println("Failed to create key, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, key)
		return
	}
	
	var redirectTo string
	if c.QueryTrim("flags") == ""{
		redirectTo = "../keys"
		c.Redirect(redirectTo)
	}else{
		redirectTo = "../instances?hostname=" + hostname
		c.Redirect(redirectTo)
	}
}



func (v *KeyView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	if c.QueryTrim("flags") != ""{
		c.Data["InstanceFlag"] = 1
	}
	hostname := c.QueryTrim("host")
	name := c.QueryTrim("name")
	publicKey, privateKey, err := keyTemp.Create()
	
	if err != nil{
		log.Println("failed")
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	
	
	c.Data["KeyName"] = name
	c.Data["PublicKey"] = publicKey
	c.Data["HostName"] = hostname
	c.Data["PrivateKey"] = privateKey
	c.HTML(200, "newKey")
}
