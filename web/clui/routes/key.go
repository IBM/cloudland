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
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/ssh"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	keyAdmin   = &KeyAdmin{}
	keyView    = &KeyView{}
	keyTemp    = &KeyTemp{}
	apiKeyView = &APIKeyView{}
)

type KeyAdmin struct{}
type KeyView struct{}
type KeyTemp struct{}
type APIKeyView struct {
	Name   string
	Pubkey string
}

func (point *KeyTemp) Create() (publicKey, fingerPrint, privateKey string, err error) {
	// generate key
	private, er := rsa.GenerateKey(rand.Reader, 1024)
	if er != nil {
		log.Println("failed to create privateKey ")
		err = er
		return
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)}
	privateKey = string(pem.EncodeToMemory(privateKeyPEM))
	pub, er := ssh.NewPublicKey(&private.PublicKey)
	if er != nil {
		log.Println("failed to create publicKey")
		err = er
		return
	}
	temp := ssh.MarshalAuthorizedKey(pub)
	publicKey = string(temp)
	fingerPrint = ssh.FingerprintLegacyMD5(pub)
	return
}

func (point *KeyTemp) CreateFingerPrint(publicKey string) (fingerPrint string, err error) {
	pubKeyBytes := []byte(publicKey)
	pub, _, _, _, puberr := ssh.ParseAuthorizedKey(pubKeyBytes)
	if puberr != nil {
		log.Println("Public key is wrong.")
		err = puberr
		return
	}
	fingerPrint = ssh.FingerprintLegacyMD5(pub)
	return
}

func (a *KeyAdmin) Create(ctx context.Context, name, pubkey, fingerprint string) (key *model.Key, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	key = &model.Key{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Name: name, PublicKey: pubkey, FingerPrint: fingerprint}
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

func (v *KeyView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "keys_new")
}

func (v *KeyView) Confirm(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	name := c.QueryTrim("name")
	publicKey := c.QueryTrim("pubkey")
	log.Println("Your Public Key, %v", publicKey)
	pubKeyBytes := []byte(publicKey)
	pub, _, _, _, _ := ssh.ParseAuthorizedKey(pubKeyBytes)
	fingerPrint := ssh.FingerprintLegacyMD5(pub)
	key, err := keyAdmin.Create(c.Req.Context(), name, publicKey, fingerPrint)
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
	if c.QueryTrim("from_instance") != "" {
		_, keys, err := keyAdmin.List(c.Req.Context(), 0, -1, "", "")
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
		c.JSON(200, map[string]interface{}{
			"keys": keys,
		})
	} else {
		var redirectTo string
		redirectTo = "../keys"
		c.Redirect(redirectTo)
	}
}

func (v *KeyView) SolvePrintedPublicKeyError(c *macaron.Context, store session.Store, err error) {
	if err != nil {
		if c.QueryTrim("from_instance") != "" {
			c.JSON(200, map[string]interface{}{
				"error": "Public key is wrong",
			})
			return
		} else {
			log.Println("Public key is wrong")
			c.Data["ErrorMsg"] = "Public key is wrong"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	return
}

func (v *KeyView) SolvePublicKeyDbError(c *macaron.Context, store session.Store, name, publicKey, fingerPrint string) {
	key, err := keyAdmin.Create(c.Req.Context(), name, publicKey, fingerPrint)
	if err != nil {
		log.Println("Failed, %v", err)
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
	return
}

func (v *KeyView) SearchDbFingerPrint(c *macaron.Context, store session.Store, fingerPrint, publicKey, name string) {
	db := DB()
	var keydb []model.Key
	x := db.Where(&model.Key{FingerPrint: fingerPrint}).Find(&keydb)
	length := len(*(x.Value.(*[]model.Key)))
	if length != 0 {
		if c.QueryTrim("from_instance") != "" {
			c.JSON(200, map[string]interface{}{
				"error": "This public key has been used",
			})
			return
		} else {
			c.Data["ErrorMsg"] = "This public key has been used"
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	} else {
		keyView.SolvePublicKeyDbError(c, store, name, publicKey, fingerPrint)
	}
}

func (v *KeyView) SolveListKeyError(c *macaron.Context, store session.Store) {
	if c.QueryTrim("from_instance") != "" {
		_, keys, err := keyAdmin.List(c.Req.Context(), 0, -1, "", "")
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
		c.JSON(200, map[string]interface{}{
			"keys": keys,
		})
	} else {
		redirectTo := "../keys"
		c.Redirect(redirectTo)
	}
	return
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
	name := c.QueryTrim("name")
	if c.QueryTrim("pubkey") != "" {
		publicKey := c.QueryTrim("pubkey")
		fingerPrint, puberr := keyTemp.CreateFingerPrint(publicKey)
		keyView.SolvePrintedPublicKeyError(c, store, puberr)
		keyView.SearchDbFingerPrint(c, store, fingerPrint, publicKey, name)
		keyView.SolveListKeyError(c, store)
	} else {
		publicKey, fingerPrint, privateKey, err := keyTemp.Create()
		if err != nil {
			log.Println("failed")
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		if c.QueryTrim("from_instance") != "" {
			fmt.Println("from_instance:" + c.QueryTrim("from_instance"))
			c.JSON(200, map[string]interface{}{
				"keyName":    name,
				"publicKey":  publicKey,
				"privateKey": privateKey,
			})
			return
		} else {
			c.Data["KeyName"] = name
			c.Data["PublicKey"] = publicKey
			c.Data["PrivateKey"] = privateKey
			c.Data["fingerPrint"] = fingerPrint
			c.HTML(200, "new_key")
		}
	}
}

func (v *APIKeyView) List(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		c.JSON(403, map[string]interface{}{
			"ErrorMsg": "Not authorized for this operation",
		})
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
		c.JSON(500, map[string]interface{}{
			"ErrorMsg": "Failed to list keys." + err.Error(),
		})
		return

	}
	pages := GetPages(total, limit)
	c.JSON(200, map[string]interface{}{
		"keys":  keys,
		"total": total,
		"pages": pages,
		"query": query,
	})

}

func (v *APIKeyView) Create(c *macaron.Context, store session.Store, apiKeyView APIKeyView) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.JSON(403, map[string]interface{}{
			"ErrorMsg": "Not authorized for this operation",
		})
		return
	}
	name := apiKeyView.Name
	pubkey := apiKeyView.Pubkey
	if pubkey != "" {
		publicKey := pubkey
		fingerPrint, err := keyTemp.CreateFingerPrint(publicKey)
		if err != nil {
			c.JSON(400, map[string]interface{}{
				"ErrorMsg": "Public key is wrong." + err.Error(),
			})
			return
		}

		db := DB()
		var keydb []model.Key
		x := db.Where(&model.Key{FingerPrint: fingerPrint}).Find(&keydb)
		length := len(*(x.Value.(*[]model.Key)))
		if length != 0 {
			c.JSON(200, map[string]interface{}{
				"ErrorMsg": "This public key has been used.",
			})
			return
		}
		key, err := keyAdmin.Create(c.Req.Context(), name, publicKey, fingerPrint)
		if err != nil {
			log.Println("Failed, %v", err)
			c.JSON(500, map[string]interface{}{
				"ErrorMsg": "Failed to insert public key to DB." + err.Error(),
			})
			return
		}
		c.JSON(200, map[string]interface{}{
			"keyName":   key.Name,
			"publicKey": key.PublicKey,
		})
		return

	} else {
		publicKey, _, privateKey, err := keyTemp.Create()
		if err != nil {
			c.JSON(500, map[string]interface{}{
				"ErrorMsg": "Failed to create public key and private key." + err.Error(),
			})
			return
		}
		c.JSON(200, map[string]interface{}{
			"keyName":    name,
			"publicKey":  publicKey,
			"privateKey": privateKey,
		})
		return
	}
}

func (v *APIKeyView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Key id is empty.",
		})
		return
	}
	keyID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid key id, %v", err)
		c.JSON(400, map[string]interface{}{
			"ErrorMsg": "Failed to get key id.",
		})
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "keys", int64(keyID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.JSON(403, map[string]interface{}{
			"ErrorMsg": "Not authorized for this operation",
		})
		return
	}
	err = keyAdmin.Delete(int64(keyID))
	if err != nil {
		log.Println("Failed to delete key, %v", err)
		c.JSON(500, map[string]interface{}{
			"ErrorMsg": "Failed to delete key." + err.Error(),
		})
		return
	}
	c.JSON(200, map[string]interface{}{
		"Msg": "Success.",
	})
	return
}
