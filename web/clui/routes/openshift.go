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
	"github.com/spf13/viper"
	macaron "gopkg.in/macaron.v1"
)

var (
	openshiftAdmin = &OpenshiftAdmin{}
	openshiftView  = &OpenshiftView{}
)

type OpenshiftAdmin struct{}
type OpenshiftView struct{}

func (a *OpenshiftAdmin) createSecgroup(ctx context.Context, name, cidr string, owner int64) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{Owner: owner}, Name: name}
	err = db.Where(secgroup).Take(secgroup).Error
	if err == nil {
		log.Println("Use existing openshift security group", err)
		return
	}
	secgroup, err = secgroupAdmin.Create(ctx, name, false, owner)
	if err != nil {
		log.Println("Failed to create security group with default rules", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "tcp", 1, 65535)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "udp", 1, 65535)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 8443, 8443)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 6443, 6443)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 22623, 22623)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 2379, 2379)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 9000, 9999)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 10249, 10259)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 30000, 32767)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 53, 53)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "udp", 53, 53)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 443, 443)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, "0.0.0.0/0", "ingress", "tcp", 80, 80)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	return
}

func (a *OpenshiftAdmin) Create(ctx context.Context, cluster, domain, secret, cookie string, haflag bool, nworkers int32, flavor, key int64) (openshift *model.Openshift, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	openshift = &model.Openshift{
		Model:       model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID},
		ClusterName: cluster,
		BaseDomain:  domain,
		Haflag:      haflag,
		WorkerNum:   nworkers,
		Flavor:      flavor,
		Key:         key,
	}
	err = db.Create(openshift).Error
	if err != nil {
		log.Println("DB failed to create openshift", err)
		return
	}
	name := fmt.Sprintf("oc%d-sn", openshift.ID)
	subnet, err := subnetAdmin.Create(ctx, name, "", "192.168.91.0", "255.255.255.0", "", "", "", "", "", memberShip.OrgID)
	if err != nil {
		log.Println("Failed to create openshift subnet", err)
		return
	}
	name = fmt.Sprintf("oc%d-gw", openshift.ID)
	subnetIDs := []int64{subnet.ID}
	_, err = gatewayAdmin.Create(ctx, name, 0, 0, subnetIDs, memberShip.OrgID)
	if err != nil {
		log.Println("Failed to create gateway", err)
		return
	}
	secgroup, err := a.createSecgroup(ctx, "openshift", "192.168.91.0/24", memberShip.OrgID)
	name = fmt.Sprintf("oc%d-lb", openshift.ID)
	keyIDs := []int64{key}
	sgIDs := []int64{secgroup.ID}
	endpoint := viper.GetString("api.endpoint")
	userdata := `#!/bin/bash
cd /opt
exec >/tmp/ocd.log 2>&1
sleep 15
grep nameserver /etc/resolv.conf
[ $? -ne 0 ] && echo nameserver 8.8.8.8 >> /etc/resolv.conf
yum -y install epel-release
yum -y install wget jq`
	userdata = fmt.Sprintf("%s\nwget '%s/misc/openshift/ocd.sh'\nchmod +x ocd.sh", userdata, endpoint)
	userdata = fmt.Sprintf("%s\n./ocd.sh '%s' '%s' '%s' '%s' <<EOF\n%s\nEOF", userdata, cluster, domain, endpoint, cookie, haflag, secret)
	_, err = instanceAdmin.Create(ctx, 1, name, userdata, 1, flavor, subnet.ID, "192.168.91.9", "", nil, keyIDs, sgIDs, -1)
	if err != nil {
		log.Println("Failed to create oc first instance", err)
		return
	}
	return
}

func (a *OpenshiftAdmin) Delete(id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	if err = db.Delete(&model.Openshift{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *OpenshiftAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, openshifts []*model.Openshift, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	if limit == 0 {
		limit = 10
	}

	if order == "" {
		order = "created_at"
	}
	if query != "" {
		query = fmt.Sprintf("name like '%%%s%%'", query)
	}

	where := memberShip.GetWhere()
	openshifts = []*model.Openshift{}
	if err = db.Model(&model.Openshift{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Where(query).Find(&openshifts).Error; err != nil {
		return
	}

	return
}

func (v *OpenshiftView) List(c *macaron.Context, store session.Store) {
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
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, openshifts, err := openshiftAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Openshifts"] = openshifts
	c.Data["Total"] = total
	c.Data["Pages"] = GetPages(total, limit)
	c.Data["Query"] = query
	c.HTML(200, "openshifts")
}

func (v *OpenshiftView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	openshiftID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", int64(openshiftID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = openshiftAdmin.Delete(int64(openshiftID))
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "openshifts",
	})
	return
}

func (v *OpenshiftView) New(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	db := DB()
	_, flavors, err := flavorAdmin.List(0, -1, "", "")
	if err := db.Find(&flavors).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	_, keys, err := keyAdmin.List(ctx, 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Flavors"] = flavors
	c.Data["Keys"] = keys
	c.HTML(200, "openshifts_new")
}

func (v *OpenshiftView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../openshifts"
	name := c.QueryTrim("clustername")
	domain := c.QueryTrim("basedomain")
	haflagStr := c.QueryTrim("haflag")
	secret := c.QueryTrim("secret")
	nworkers := c.QueryInt("nworkers")
	flavor := c.QueryInt64("flavor")
	key := c.QueryInt64("key")
	haflag := false
	if haflagStr == "" || haflagStr == "no" {
		haflag = false
	} else if haflagStr == "yes" {
		haflag = true
	}
	cookie := "MacaronSession=" + c.GetCookie("MacaronSession")
	_, err := openshiftAdmin.Create(c.Req.Context(), name, domain, secret, cookie, haflag, int32(nworkers), flavor, key)
	if err != nil {
		c.HTML(500, "500")
	}
	c.Redirect(redirectTo)
}
