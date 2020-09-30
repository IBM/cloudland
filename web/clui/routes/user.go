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
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
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

func (a *UserAdmin) Create(ctx context.Context, username, password string) (user *model.User, err error) {
	memberShip := GetMemberShip(ctx)
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
	if err = db.Where("user_id = ?", id).Delete(&model.Member{}).Error; err != nil {
		log.Println("DB failed to delete members", err)
		return
	}
	if err = db.Delete(&model.User{Model: model.Model{ID: id}}).Error; err != nil {
		log.Println("DB failed to delete members", err)
		return
	}
	return
}

func (a *UserAdmin) Update(ctx context.Context, id int64, password string, members []string) (user *model.User, err error) {
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

func (a *UserAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, users []*model.User, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 16
	}

	if order == "" {
		order = "created_at"
	}

	if query != "" {
		query = fmt.Sprintf("username like '%%%s%%'", query)
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
		if err = db.Model(&model.User{}).Where(userIDs).Where(query).Count(&total).Error; err != nil {
			return
		}
		db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
		if err = db.Where(userIDs).Where(query).Find(&users).Error; err != nil {
			log.Println("DB failed to get user list, %v", err)
			return
		}
	} else {
		if err = db.Model(&model.User{}).Where(query).Count(&total).Error; err != nil {
			return
		}
		db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
		if err = db.Where(query).Find(&users).Error; err != nil {
			log.Println("DB failed to get user list, %v", err)
			return
		}
	}

	return
}

func (a *UserAdmin) Validate(ctx context.Context, username, password string) (user *model.User, err error) {
	db := DB()
	hasUser := true
	user = &model.User{}
	err = db.Take(user, "username = ?", username).Error
	if err != nil {
		log.Println("DB failed to qeury user", err)
		hasUser = false
	}
	if strings.Contains(username, "@") && strings.Contains(username, "ibm.com") {
		cmd := exec.Command("/opt/cloudland/scripts/frontend/ldap_auth.sh", username, password)
		err = cmd.Start()
		if err != nil {
			log.Println("cmd.Start: ", err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Println("cmd.Wait: ", err)
			return
		} else if hasUser {
			return
		}
		_, err = a.Create(ctx, username, password)
		if err != nil {
			log.Println("Failed to create user", err)
			return
		}
		_, err = orgAdmin.Create(ctx, username, username)
		if err != nil {
			log.Println("Failed to create organization", err)
			return
		}
		return
	}
	err = a.CompareHashAndPassword(user.Password, password)
	return
}

func (a *UserAdmin) AccessToken(uid int64, username, organization string) (oid int64, role model.Role, token string, issueAt, expiresAt int64, err error) {
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
	role = member.Role
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
	logout := c.QueryTrim("logout")
	if logout == "" {
		c.Data["PageIsSignIn"] = true
		c.HTML(200, "login")
	} else {
	}
}

func (v *UserView) LoginPost(c *macaron.Context, store session.Store) {
	username := c.QueryTrim("username")
	password := c.QueryTrim("password")
	user, err := userAdmin.Validate(c.Req.Context(), username, password)
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
	redirectTo := UrlBefore
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		cookie := "MacaronSession=" + c.GetCookie("MacaronSession")
		c.JSON(200, map[string]interface{}{
			"user":   username,
			"uid":    uid,
			"oid":    oid,
			"token":  token,
			"cookie": cookie,
		})
		return
	}
	c.Redirect(redirectTo)
}

func (v *UserView) List(c *macaron.Context, store session.Store) {
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
	total, users, err := userAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list user(s)", err)
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
	c.Data["Users"] = users
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"users": users,
			"total": total,
			"pages": pages,
			"query": query,
		})
		return
	}
	c.HTML(200, "users")
}

func (v *UserView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckUser(int64(userID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	user := &model.User{Model: model.Model{ID: int64(userID)}}
	err = db.Set("gorm:auto_preload", true).Take(user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.Data["User"] = user
	c.HTML(200, "users_patch")
}

func (v *UserView) Change(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	orgName := c.QueryTrim("org")
	db := DB()
	user := &model.User{Model: model.Model{ID: int64(userID)}}
	err = db.Set("gorm:auto_preload", true).Take(user).Error
	if err != nil {
		log.Println("Failed to query user", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "/dashboard"
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
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get input id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckUser(int64(userID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	password := c.QueryTrim("password")
	members := c.QueryStrings("members")
	user, err := userAdmin.Update(c.Req.Context(), int64(userID), password, members)
	if err != nil {
		log.Println("Failed to update password, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, user)
		return
	}
	c.HTML(200, "ok")
}

func (v *UserView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	id := c.Params("id")
	if id == "" {
		log.Println("User id is empty, %v", err)
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to get user id, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = userAdmin.Delete(int64(userID))
	if err != nil {
		log.Println("Failed to delete user, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "/users",
	})
	return
}

func (v *UserView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "users_new")
}

func (v *UserView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Admin)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	redirectTo := "/users"
	username := c.QueryTrim("username")
	password := c.QueryTrim("password")
	confirm := c.QueryTrim("confirm")

	if confirm != password {
		log.Println("Passwords do not match")
		c.HTML(http.StatusBadRequest, "Passwords do not match")
		return
	}
	user, err := userAdmin.Create(c.Req.Context(), username, password)
	if err != nil {
		log.Println("Failed to create user, %v", err)
		c.HTML(500, "500")
		return
	}
	_, err = orgAdmin.Create(c.Req.Context(), username, username)
	if err != nil {
		log.Println("Failed to create organization, %v", err)
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.HTML(500, "500")
		return
	}
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, user)
		return
	}
	c.Redirect(redirectTo)
}
