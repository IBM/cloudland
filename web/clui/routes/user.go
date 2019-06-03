/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	user = &model.User{Username: username, Password: password}
	err = db.Create(user).Error
	if err != nil {
		log.Println("DB failed to create user, %v", err)
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
	ids := []int64{}
	if id == 0 { // delete all
		users := []model.User{}
		if err = db.Select("id").Find(&users).Error; err != nil {
			log.Println("DB failed to find users, %v", err)
			return
		}
		for _, user := range users {
			ids = append(ids, user.ID)
		}
	} else {
		ids = append(ids, id)
	}
	for _, id = range ids {
		if err = db.Delete(&model.User{Model: model.Model{ID: id}}).Error; err != nil {
			log.Println("DB failed to delete user, %v", err)
			return
		}
	}
	return
}

func (a *UserAdmin) Update(id int64, password string) (user *model.User, err error) {
	if password, err = a.GenerateFromPassword(password); err != nil {
		return
	}
	db := DB()
	err = db.Model(&model.User{Model: model.Model{ID: id}}).Update("password", password).Error
	if err != nil {
		log.Println("DB failed to update user password, %v", err)
	}
	return
}

func (a *UserAdmin) Show(id int64) (user *model.User, err error) {
	db := DB()
	user = &model.User{}
	err = db.Take(user, id).Error
	if err != nil {
		log.Println("DB failed to get user, %v", err)
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

	users = []*model.User{}
	if err = db.Model(&model.User{}).Count(&total).Error; err != nil {
		log.Println("DB failed to count user, %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Find(&users).Error; err != nil {
		log.Println("DB failed to get user list, %v", err)
		return
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

func (a *UserAdmin) AccessToken(uid int64, username, organization string) (oid int64, role model.Role, token string, err error) {
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
	token, err = NewToken(username, organization, uid, oid, role)
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
	oid, role, token, err := userAdmin.AccessToken(uid, username, organization)
	if err != nil {
		log.Println("Failed to get token, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(403, "403")
		return
	}
	store.Set("login", username)
	store.Set("uid", uid)
	store.Set("oid", oid)
	store.Set("role", role)
	store.Set("act", token)
	store.Set("org", organization)
	redirectTo := "/instances"
	c.Redirect(redirectTo)
}

func (v *UserView) List(c *macaron.Context, store session.Store) {
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	total, users, err := userAdmin.List(offset, limit, order)
	if err != nil {
		log.Println("Failed to list user(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Users"] = users
	c.Data["Total"] = total
	c.HTML(200, "users")
}

func (v *UserView) Show(c *macaron.Context, store session.Store) {
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
	user, err := userAdmin.Show(int64(userID))
	if err != nil {
		log.Println("Failed to show user, %v", err)
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	c.Data["Username"] = user.Username
	c.Data["UserID"] = user.ID
	c.HTML(200, "users_show")
}

func (v *UserView) Update(c *macaron.Context, store session.Store) {
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
	password := c.Query("password")
	_, err = userAdmin.Update(int64(userID), password)
	if err != nil {
		log.Println("Failed to update password, %v", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.Redirect("/admin/users")
}

func (v *UserView) Delete(c *macaron.Context, store session.Store) (err error) {
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
	/*
		orgs, err := orgAdmin.List(int64(userID))
		if err != nil {
			code := http.StatusInternalServerError
			c.Error(code, http.StatusText(code))
			return
		}
		for _, org := range orgs {
			if err = orgAdmin.Delete(org.ID); err != nil {
				code := http.StatusInternalServerError
				c.Error(code, http.StatusText(code))
				return
			}
		}
	*/
	c.JSON(200, map[string]interface{}{
		"redirect": "/users",
	})
	return
}

func (v *UserView) New(c *macaron.Context, store session.Store) {
	c.HTML(200, "users_new")
}

func (v *UserView) Create(c *macaron.Context, store session.Store) {
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
	_, err = orgAdmin.Create(username, username)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		c.HTML(500, "500")
	}
	sgName := username + ":default"
	_, err = secgroupAdmin.Create(sgName, true)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
