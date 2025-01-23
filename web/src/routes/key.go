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
	"net/http"
	"strconv"

	"golang.org/x/crypto/ssh"

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	keyAdmin = &KeyAdmin{}
	keyView  = &KeyView{}
)

type KeyAdmin struct{}
type KeyView struct{}

func (a *KeyAdmin) CreateKeyPair(ctx context.Context) (publicKey, fingerPrint, privateKey string, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized to create keys")
		err = fmt.Errorf("Not authorized")
		return
	}
	// generate key
	private, er := rsa.GenerateKey(rand.Reader, 1024)
	if er != nil {
		logger.Error("failed to create privateKey ")
		err = er
		return
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)}
	privateKey = string(pem.EncodeToMemory(privateKeyPEM))
	pub, er := ssh.NewPublicKey(&private.PublicKey)
	if er != nil {
		logger.Error("failed to create publicKey")
		err = er
		return
	}
	temp := ssh.MarshalAuthorizedKey(pub)
	publicKey = string(temp)
	fingerPrint = ssh.FingerprintLegacyMD5(pub)
	return
}

func (a *KeyAdmin) Create(ctx context.Context, name, publicKey string) (key *model.Key, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized to create keys")
		err = fmt.Errorf("Not authorized")
		return
	}
	pub, _, _, _, puberr := ssh.ParseAuthorizedKey([]byte(publicKey))
	if puberr != nil {
		logger.Error("Invalid public key")
		err = puberr
		return
	}
	fingerPrint := ssh.FingerprintLegacyMD5(pub)
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	key = &model.Key{Model: model.Model{Creater: memberShip.UserID}, Owner: memberShip.OrgID, Name: name, PublicKey: publicKey, FingerPrint: fingerPrint}
	err = db.Create(key).Error
	if err != nil {
		logger.Error("DB failed to create key, %v", err)
		return
	}
	return
}

func (a *KeyAdmin) Delete(ctx context.Context, key *model.Key) (err error) {
	ctx, db, newTransaction := StartTransaction(ctx)
	defer func() {
		if newTransaction {
			EndTransaction(ctx, err)
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, key.Owner)
	if !permit {
		logger.Error("Not authorized to delete the key")
		err = fmt.Errorf("Not authorized")
		return
	}
	key.Name = fmt.Sprintf("%s-%d", key.Name, key.CreatedAt.Unix())
	err = db.Model(key).Update("name", key.Name).Error
	if err != nil {
		logger.Error("DB failed to update key name", err)
		return
	}
	if err = db.Delete(key).Error; err != nil {
		logger.Error("DB failed to delete key ", err)
		return
	}
	return
}

func (a *KeyAdmin) Get(ctx context.Context, id int64) (key *model.Key, err error) {
	if id <= 0 {
		err = fmt.Errorf("Invalid key ID: %d", id)
		logger.Error(err)
		return
	}
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	key = &model.Key{Model: model.Model{ID: id}}
	err = db.Where(where).Take(key).Error
	if err != nil {
		logger.Error("Failed to query key, %v", err)
		return
	}
	return
}

func (a *KeyAdmin) GetKeyByUUID(ctx context.Context, uuID string) (key *model.Key, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	key = &model.Key{}
	err = db.Where(where).Where("uuid = ?", uuID).Take(key).Error
	if err != nil {
		logger.Error("Failed to query key, %v", err)
		return
	}
	return
}

func (a *KeyAdmin) GetKeyByName(ctx context.Context, name string) (key *model.Key, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	key = &model.Key{}
	err = db.Where(where).Where("name = ?", name).Take(key).Error
	if err != nil {
		logger.Error("Failed to query key, %v", err)
		return
	}
	return
}

func (a *KeyAdmin) GetKey(ctx context.Context, reference *BaseReference) (key *model.Key, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Key base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		key, err = a.GetKeyByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		key, err = a.GetKeyByName(ctx, reference.Name)
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
		logger.Error("DB failed to count keys, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Where(query).Find(&keys).Error; err != nil {
		logger.Error("DB failed to query keys, %v", err)
		return
	}
	permit := memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, key := range keys {
			key.OwnerInfo = &model.Organization{Model: model.Model{ID: key.Owner}}
			if err = db.Take(key.OwnerInfo).Error; err != nil {
				logger.Error("Failed to query owner info", err)
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
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.Error(http.StatusBadRequest)
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
		logger.Error("Failed to list keys, %v", err)
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
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.Error(http.StatusBadRequest)
		return
	}
	keyID, err := strconv.Atoi(id)
	if err != nil {
		logger.Error("Invalid key id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	key, err := keyAdmin.Get(ctx, int64(keyID))
	if err != nil {
		logger.Error("Failed to get key, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = keyAdmin.Delete(ctx, key)
	if err != nil {
		logger.Error("Failed to delete key, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
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
		logger.Error("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.Error(http.StatusBadRequest)
		return
	}
	c.HTML(200, "keys_new")
}

func (v *KeyView) Confirm(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	name := c.QueryTrim("name")
	publicKey := c.QueryTrim("pubkey")
	_, err := keyAdmin.Create(ctx, name, publicKey)
	if err != nil {
		logger.Error("Failed to create key ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	if c.QueryTrim("from_instance") != "" {
		_, _, err := keyAdmin.List(c.Req.Context(), 0, -1, "", "")
		if err != nil {
			logger.Error("Failed to list keys ", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(500, "500")
			return
		}
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
			logger.Error("Public key is wrong")
			c.Data["ErrorMsg"] = "Public key is wrong"
			c.Error(http.StatusBadRequest)
			return
		}
	}
	return
}

/*
func (v *KeyView) SolvePublicKeyDbError(c *macaron.Context, store session.Store, name, publicKey, fingerPrint string) {
	key, err := keyAdmin.Create(c.Req.Context(), name, publicKey, fingerPrint)
	if err != nil {
		logger.Error("Failed, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
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
			c.Error(http.StatusBadRequest)
			return
		}
	} else {
		keyView.SolvePublicKeyDbError(c, store, name, publicKey, fingerPrint)
	}
}
*/

func (v *KeyView) SolveListKeyError(c *macaron.Context, store session.Store) {
	if c.QueryTrim("from_instance") != "" {
		_, _, err := keyAdmin.List(c.Req.Context(), 0, -1, "", "")
		if err != nil {
			logger.Error("Failed to list keys, %v", err)
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(500, "500")
			return
		}
	} else {
		redirectTo := "../keys"
		c.Redirect(redirectTo)
	}
	return
}

func (v *KeyView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	name := c.QueryTrim("name")
	if c.QueryTrim("pubkey") != "" {
		publicKey := c.QueryTrim("pubkey")
		_, err := keyAdmin.Create(ctx, name, publicKey)
		if err != nil {
			logger.Error("failed to create key")
			c.Data["ErrorMsg"] = err.Error()
			c.Error(http.StatusBadRequest)
		}
		redirectTo := "../keys"
		c.Redirect(redirectTo)
	} else {
		publicKey, fingerPrint, privateKey, err := keyAdmin.CreateKeyPair(ctx)
		if err != nil {
			logger.Error("failed")
			c.Data["ErrorMsg"] = err.Error()
			c.Error(http.StatusBadRequest)
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
