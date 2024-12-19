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

	. "web/src/common"
	"web/src/dbs"
	"web/src/model"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	secgroupAdmin = &SecgroupAdmin{}
	secgroupView  = &SecgroupView{}
)

type SecgroupAdmin struct{}
type SecgroupView struct{}

func (a *SecgroupAdmin) Switch(ctx context.Context, newSg *model.SecurityGroup, router *model.Router) (err error) {
	if router == nil {
		log.Println("Not authorized to change system default security group")
		err = fmt.Errorf("Not authorized")
		return
	}
	db := DB()
	oldSg := &model.SecurityGroup{Model: model.Model{ID: router.DefaultSG}}
	err = db.Take(oldSg).Error
	if err != nil {
		log.Println("Failed to query default security group", err)
	}
	oldSg.IsDefault = false
	err = db.Save(oldSg).Error
	if err != nil {
		log.Println("Failed to save new security group", err)
	}
	return
	router.DefaultSG = newSg.ID
	err = db.Save(router).Error
	if err != nil {
		log.Println("Failed to save router", err)
	}
	newSg.IsDefault = true
	err = db.Save(newSg).Error
	if err != nil {
		log.Println("Failed to save new security group", err)
	}
	return
}

func (a *SecgroupAdmin) Update(ctx context.Context, sgID int64, name string, isDefault bool) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{ID: sgID}}
	err = db.Take(secgroup).Error
	if err != nil {
		log.Println("Failed to query security group", err)
		return
	}
	if name != "" && secgroup.Name != name {
		secgroup.Name = name
	}
	if isDefault && secgroup.IsDefault != isDefault {
		secgroup.IsDefault = isDefault
	}
	err = db.Save(secgroup).Error
	if err != nil {
		log.Println("Failed to save security group", err)
		return
	}
	return
}

func (a *SecgroupAdmin) Get(ctx context.Context, id int64) (secgroup *model.SecurityGroup, err error) {
	if id < 0 {
		return a.GetSecgroupByName(ctx, SystemDefaultSGName)
	}
	memberShip := GetMemberShip(ctx)
	db := DB()
	where := memberShip.GetWhere()
	secgroup = &model.SecurityGroup{Model: model.Model{ID: id}}
	err = db.Preload("Router").Where(where).Take(secgroup).Error
	if err != nil {
		log.Println("DB failed to query secgroup ", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, secgroup.Owner)
	if !permit {
		log.Println("Not authorized to get security group")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *SecgroupAdmin) GetSecgroupByUUID(ctx context.Context, uuID string) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	secgroup = &model.SecurityGroup{}
	err = db.Preload("Router").Where(where).Where("uuid = ?", uuID).Take(secgroup).Error
	if err != nil {
		log.Println("Failed to query secgroup ", err)
		return
	}
	permit := memberShip.ValidateOwner(model.Reader, secgroup.Owner)
	if !permit {
		log.Println("Not authorized to get security group")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *SecgroupAdmin) GetSecgroupByName(ctx context.Context, name string) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	memberShip := GetMemberShip(ctx)
	where := memberShip.GetWhere()
	secgroup = &model.SecurityGroup{}
	err = db.Where(where).Where("name = ?", name).Take(secgroup).Error
	if err != nil {
		log.Println("Failed to query secgroup ", err)
		return
	}
	if secgroup.RouterID > 0 {
		err = db.Where("router_id = ?", secgroup.ID).Take(secgroup.Router).Error
		if err != nil {
			log.Println("Failed to query router ", err)
			return
		}
	}
	permit := memberShip.ValidateOwner(model.Reader, secgroup.Owner)
	if !permit {
		log.Println("Not authorized to get security group")
		err = fmt.Errorf("Not authorized")
		return
	}
	return
}

func (a *SecgroupAdmin) GetSecurityGroup(ctx context.Context, reference *BaseReference) (secgroup *model.SecurityGroup, err error) {
	if reference == nil || (reference.ID == "" && reference.Name == "") {
		err = fmt.Errorf("Security group base reference must be provided with either uuid or name")
		return
	}
	if reference.ID != "" {
		secgroup, err = a.GetSecgroupByUUID(ctx, reference.ID)
		return
	}
	if reference.Name != "" {
		secgroup, err = a.GetSecgroupByName(ctx, reference.Name)
		return
	}
	return
}

func (a *SecgroupAdmin) GetSecgroupInterfaces(ctx context.Context, secgroup *model.SecurityGroup) (err error) {
	db := DB()
	err = db.Model(secgroup).Preload("Address").Related(&secgroup.Interfaces, "Interfaces").Error
	if err != nil {
		log.Println("Failed to query secgroup, %v", err)
		return
	}
	return
}

func (a *SecgroupAdmin) Create(ctx context.Context, name string, isDefault bool, router *model.Router) (secgroup *model.SecurityGroup, err error) {
	memberShip := GetMemberShip(ctx)
	owner := memberShip.OrgID
	var routerID int64
	if router != nil {
		permit := memberShip.ValidateOwner(model.Writer, router.Owner)
		if !permit {
			log.Println("Not authorized for this operation")
			err = fmt.Errorf("Not authorized")
			return
		}
		routerID = router.ID
	} else {
		permit := memberShip.CheckPermission(model.Admin)
		if !permit {
			log.Println("Not authorized for this operation")
			err = fmt.Errorf("Not authorized")
			return
		}
	}
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{Creater: memberShip.UserID}, Owner: owner, Name: name, IsDefault: isDefault, RouterID: routerID}
	err = db.Create(secgroup).Error
	if err != nil {
		log.Println("DB failed to create security group, %v", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, "0.0.0.0/0", "egress", "tcp", 1, 65535, secgroup)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, "0.0.0.0/0", "egress", "udp", 1, 65535, secgroup)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, "0.0.0.0/0", "ingress", "tcp", 22, 22, secgroup)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, "0.0.0.0/0", "ingress", "udp", 68, 68, secgroup)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, "0.0.0.0/0", "egress", "icmp", -1, -1, secgroup)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, "0.0.0.0/0", "ingress", "icmp", -1, -1, secgroup)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	if router != nil {
		var subnets []*model.Subnet
		err = db.Where("router_id = ?", router.ID).Find(&subnets).Error
		if err != nil {
			log.Println("Failed to create security rule", err)
			return
		}
		for _, subnet := range subnets {
			_, err = secruleAdmin.Create(ctx, subnet.Network, "ingress", "tcp", 1, 65535, secgroup)
			if err != nil {
				log.Println("Failed to create security rule", err)
				return
			}
			_, err = secruleAdmin.Create(ctx, subnet.Network, "ingress", "udp", 1, 65535, secgroup)
			if err != nil {
				log.Println("Failed to create security rule", err)
				return
			}
		}
		if isDefault {
			err = a.Switch(ctx, secgroup, router)
			if err != nil {
				log.Println("Failed to set default security group", err)
				return
			}
		}
	}
	return
}

func (a *SecgroupAdmin) Delete(ctx context.Context, secgroup *model.SecurityGroup) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.ValidateOwner(model.Writer, secgroup.Owner)
	if !permit {
		log.Println("Not authorized to delete the router")
		err = fmt.Errorf("Not authorized")
		return
	}
	if len(secgroup.Interfaces) > 0 {
		log.Println("Security group has associated interfaces")
		err = fmt.Errorf("Security group has associated interfaces")
		return
	}
	err = db.Where("secgroup = ?", secgroup.ID).Delete(&model.SecurityRule{}).Error
	if err != nil {
		log.Println("DB failed to delete security group rules", err)
		return
	}
	secgroup.Name = fmt.Sprintf("%s-%d", secgroup.Name, secgroup.CreatedAt.Unix())
	err = db.Save(secgroup).Error
	if err != nil {
		log.Println("DB failed to update security group name", err)
		return
	}
	if err = db.Delete(secgroup).Error; err != nil {
		log.Println("DB failed to delete security group", err)
		return
	}
	return
}

func (a *SecgroupAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, secgroups []*model.SecurityGroup, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Reader)
	if !permit {
		log.Println("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
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
	secgroups = []*model.SecurityGroup{}
	if err = db.Model(&model.SecurityGroup{}).Where(where).Where(query).Count(&total).Error; err != nil {
		log.Println("DB failed to count security group(s), %v", err)
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Preload("Router").Where(where).Where(query).Find(&secgroups).Error; err != nil {
		log.Println("DB failed to query security group(s), %v", err)
		return
	}
	permit = memberShip.CheckPermission(model.Admin)
	if permit {
		db = db.Offset(0).Limit(-1)
		for _, sg := range secgroups {
			sg.OwnerInfo = &model.Organization{Model: model.Model{ID: sg.Owner}}
			if err = db.Take(sg.OwnerInfo).Error; err != nil {
				log.Println("Failed to query owner info", err)
				err = nil
				continue
			}
		}
	}

	return
}

func (v *SecgroupView) List(c *macaron.Context, store session.Store) {
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
	total, secgroups, err := secgroupAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		log.Println("Failed to list security group(s), %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	pages := GetPages(total, limit)
	c.Data["SecurityGroups"] = secgroups
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	c.HTML(200, "secgroups")
}

func (v *SecgroupView) Delete(c *macaron.Context, store session.Store) (err error) {
	ctx := c.Req.Context()
	id := c.Params("id")
	if id == "" {
		c.Data["ErrorMsg"] = "Id is Empty"
		c.Error(http.StatusBadRequest)
		return
	}
	secgroupID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid security group ID, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	secgroup, err := secgroupAdmin.Get(ctx, int64(secgroupID))
	if err != nil {
		log.Println("Failed to get security group", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	err = secgroupAdmin.Delete(ctx, secgroup)
	if err != nil {
		log.Printf("Failed to delete security group, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.Error(http.StatusBadRequest)
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "secgroups",
	})
	return
}
func (v *SecgroupView) New(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	c.HTML(200, "secgroups_new")
}

func (v *SecgroupView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	db := DB()
	id := c.Params(":id")
	sgID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	secgroup := &model.SecurityGroup{Model: model.Model{ID: int64(sgID)}}
	err = db.Take(secgroup).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, err.Error())
		return
	}
	c.Data["Secgroup"] = secgroup
	c.HTML(200, "secgroups_patch")
}

func (v *SecgroupView) Patch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	redirectTo := "../secgroups"
	id := c.Params(":id")
	name := c.QueryTrim("name")
	sgID, err := strconv.Atoi(id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	permit, err := memberShip.CheckOwner(model.Writer, "security_groups", int64(sgID))
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	isdefStr := c.QueryTrim("isdefault")
	isDef := false
	if isdefStr == "" || isdefStr == "no" {
		isDef = false
	} else if isdefStr == "yes" {
		isDef = true
	}
	_, err = secgroupAdmin.Update(c.Req.Context(), int64(sgID), name, isDef)
	if err != nil {
		c.HTML(500, err.Error())
		return
	}
	/*
		if isDef {
			err = secgroupAdmin.Switch(c.Req.Context(), secgroup, store)
			if err != nil {
				log.Println("Failed to switch security group", err)
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
		}
	*/
	c.Redirect(redirectTo)
	return
}

func (v *SecgroupView) Create(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	redirectTo := "../secgroups"
	name := c.QueryTrim("name")
	isdefStr := c.QueryTrim("isdefault")
	isDef := false
	if isdefStr == "" || isdefStr == "no" {
		isDef = false
	} else if isdefStr == "yes" {
		isDef = true
	}
	routerID := c.QueryInt64("router")
	router, err := routerAdmin.Get(ctx, routerID)
	if err != nil {
		log.Println("Failed to get vpc", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(404, "404")
	}
	_, err = secgroupAdmin.Create(ctx, name, isDef, router)
	if err != nil {
		log.Println("Failed to create security group, %v", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Redirect(redirectTo)
}
