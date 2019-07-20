/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	"golang.org/x/crypto/bcrypt"
	macaron "gopkg.in/macaron.v1"
)

var (
	userAdmin = &UserAdmin{}
	userView  = &UserView{}
)

type UserAdmin struct{}

type UserView struct{}

func (a *UserAdmin) Create(username, password string) (user *model.User, err error) {
	db := DB()
	if password, err = a.GenerateFromPassword(password); err != nil {
		return
	}
	user = &model.User{Model: model.Model{Creater: memberShip.UserID}, Username: username, Password: password}
	err = db.Create(user).Error
	if err != nil {
		log.Println("DB failed to create user, %v", err)
	}
	if memberShip.OrgName != "admin" {
		member := &model.Member{UserID: user.ID, UserName: username, OrgID: memberShip.OrgID, OrgName: memberShip.OrgName, Role: model.Reader}
		err = db.Create(member).Error
		if err != nil {
			log.Println("DB failed to create organization member ", err)
			return
		}
	}
	return
}

func (a *UserAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.User{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("DB failed to delete user, %v", err)
		return
	}
	return
}

func (a *UserAdmin) Update(id int64, password string, members []string) (user *model.User, err error) {
	db := DB()
	user = &model.User{Model: model.Model{ID: id}}
	err = db.Set("gorm:auto_preload", true).Take(user).Error
	if err != nil {
		log.Println("DB failed to query user", err)
		return
	}
	password = strings.TrimSpace(password)
	if password != "" {
		if password, err = a.GenerateFromPassword(password); err != nil {
			return
		}
		err = db.Model(user).Update("password", password).Error
		if err != nil {
			log.Println("DB failed to update user password", err)
			return
		}
	}
	for _, em := range user.Members {
		found := false
		for _, name := range members {
			if em.OrgName == name {
				found = true
				break
			}
		}
		if found == false {
			err = db.Where("user_name = ? and org_name = ?", user.Username, em.OrgName).Delete(&model.Member{}).Error
			if err != nil {
				log.Println("DB failed to delete member", err)
				return
			}
		}
	}
	return
}

func (a *UserAdmin) List(offset, limit int64, order string) (total int64, users []*model.User, err error) {
	db := DB()
	if limit == 0 {
		limit = 20
	}

	if order == "" {
		order = "created_at"
	}

	if memberShip.Role != model.Admin {
		org := &model.Organization{Model: model.Model{ID: memberShip.OrgID}}
		if err = db.Set("gorm:auto_preload", true).Take(org).Error; err != nil {
			log.Println("Failed to query organization", err)
			return
		}
		var userIDs []int64
		if org.Members != nil {
			total = int64(len(org.Members))
			for _, member := range org.Members {
				userIDs = append(userIDs, member.UserID)
			}
		}
		db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
		if err = db.Where(userIDs).Find(&users).Error; err != nil {
			log.Println("DB failed to get user list, %v", err)
			return
		}
	} else {
		if err = db.Find(&users).Error; err != nil {
			log.Println("DB failed to get user list, %v", err)
			return
		}
	}

	return
}

func (a *UserAdmin) Validate(username, password string) (user *model.User, err error) {
	user = &model.User{}
	db := DB()
	err = db.Take(user, "username = ?", username).Error
	if err != nil {
		log.Println("DB failed to validate user, %v", err)
		return
	}
	err = a.CompareHashAndPassword(user.Password, password)
	return
}

func (a *UserAdmin) AccessToken(uid int64, username, organization string) (oid int64, role *model.Role, token string, issueAt, expiresAt int64, err error) {
	db := DB()
	member := &model.Member{}
	err = db.Take(member, "user_name = ? and org_name = ?", username, organization).Error
	if err != nil {
		log.Println("DB failed to get membership, %v", err)
		return
	}
	if member.Role == model.None {
		err = fmt.Errorf("user %s has no role under organization %s", username, organization)
		return
	}
	oid = member.OrgID
	role = &member.Role
	orgInstance := &model.Organization{
		Model: model.Model{ID: oid},
	}
	userInstance := &model.User{
		Model: model.Model{ID: uid},
	}
	if err = db.First(orgInstance).Error; err != nil {
		return
	}
	if err = db.First(userInstance).Error; err != nil {
		return
	}
	token, issueAt, expiresAt, err = NewToken(username, organization, userInstance.UUID, orgInstance.UUID, role)
	return
}

// GenerateFromPassword is slow by design, do not call it too offen.
func (a *UserAdmin) GenerateFromPassword(password string) (hash string, err error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return
	}
	hash = string(b)
	return
}

// CompareHashAndPassword is slow by design, do not call it too offen.
func (a *UserAdmin) CompareHashAndPassword(hash, password string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return
}

func (v *UserView) LoginGet(c *macaron.Context, store session.Store) {
	adminInit(c.Req.Context())
	logout := c.Query("logout")
	if logout == "" {
		c.Data["PageIsSignIn"] = true
		c.HTML(200, "login")
	} else {
	}
}

func (v *UserView) LoginPost(c *macaron.Context, store session.Store) {
	username := c.Query("username")
	password := c.Query("password")
	user, err := userAdmin.Validate(username, password)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(401, "401")
		return
	}
	organization := username
	uid := user.ID
	oid, role, token, _, _, err := userAdmin.AccessToken(uid, username, organization)
	if err != nil {
		log.Println("Failed to get token", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(403, "403")
		return
	}
	members := []*model.Member{}
	err = dbs.DB().Where("user_name = ?", username).Find(&members).Error
	if err != nil {
		log.Println("Failed to query organizations, ", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(403, "403")
	}
	org, err := orgAdmin.Get(organization)
	store.Set("login", username)
	store.Set("uid", uid)
	store.Set("oid", oid)
	store.Set("role", role)
	store.Set("act", token)
	store.Set("org", organization)
	store.Set("defsg", org.DefaultSG)
	store.Set("members", members)
	redirectTo := "/instances"
	c.Redirect(redirectTo)
}

func (v *UserView) List(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, users, err := userAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list user(s)", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Users"] = users
	c.Data["Total"] = total
	c.HTML(200, "users")
}

func (v *UserView) Edit(c *macaron.Context, store session.Store) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckUser(int64(userID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	db := DB()
	user := &model.User{Model: model.Model{ID: int64(userID)}}
	err = db.Set("gorm:auto_preload", true).Take(user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	c.Data["User"] = user
	c.HTML(200, "users_patch")
}

func (v *UserView) Change(c *macaron.Context, store session.Store) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	orgName := c.Query("org")
	db := DB()
	user := &model.User{Model: model.Model{ID: int64(userID)}}
	err = db.Set("gorm:auto_preload", true).Take(user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "/instances"
	orgName = strings.TrimSpace(orgName)
	if orgName != "" {
		for _, em := range user.Members {
			if em.OrgName == orgName {
				org := &model.Organization{Model: model.Model{ID: em.OrgID}}
				err = db.Take(org).Error
				if err != nil {
					log.Println("Failed to query organization")
				} else {
					store.Set("oid", org.ID)
					store.Set("role", em.Role)
					store.Set("org", org.Name)
					store.Set("defsg", org.DefaultSG)
					memberShip.OrgID = org.ID
					memberShip.OrgName = org.Name
					memberShip.Role = em.Role
				}
				break
			}
		}
	}
	c.Data["User"] = user
	c.Redirect(redirectTo)
}

func (v *UserView) Patch(c *macaron.Context, store session.Store) {
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckUser(int64(userID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../users/" + id
	password := c.Query("password")
	members := c.QueryStrings("members")
	_, err = userAdmin.Update(int64(userID), password, members)
	if err != nil {
		log.Println("Failed to update password, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.Redirect(redirectTo)
}

func (v *UserView) Delete(c *macaron.Context, store session.Store) (err error) {
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	id := c.Params("id")
	if id == "" {
		log.Println("User id is empty, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get user id, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	err = userAdmin.Delete(int64(userID))
	if err != nil {
		log.Println("Failed to delete user, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "/users",
	})
	return
}

func (v *UserView) New(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	c.HTML(200, "users_new")
}

func (v *UserView) Create(c *macaron.Context, store session.Store) {
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "/users"
	username := c.Query("username")
	password := c.Query("password")
	confirm := c.Query("confirm")

	if confirm != password {
		log.Println("Passwords do not match")
		c.HTML(http.StatusBadRequest, "Passwords do not match")
	}
	_, err := userAdmin.Create(username, password)
	if err != nil {
		log.Println("Failed to create user, %v", err)
		c.HTML(500, "500")
	}
	_, err = orgAdmin.Create(c.Req.Context(), username, username)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
