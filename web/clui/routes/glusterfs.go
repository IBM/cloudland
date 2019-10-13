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
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/go-macaron/session"
	"github.com/spf13/viper"
	macaron "gopkg.in/macaron.v1"
)

var (
	glusterfsAdmin = &GlusterfsAdmin{}
	glusterfsView  = &GlusterfsView{}
)

type GlusterfsAdmin struct{}
type GlusterfsView struct{}

func (a *GlusterfsAdmin) createSecgroup(ctx context.Context, name, cidr string, owner int64) (secgroup *model.SecurityGroup, err error) {
	db := DB()
	secgroup = &model.SecurityGroup{Model: model.Model{Owner: owner}, Name: name}
	err = db.Where(secgroup).Take(secgroup).Error
	if err == nil {
		log.Println("Use existing glusterfs security group", err)
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
	return
}

func (a *GlusterfsAdmin) State(ctx context.Context, id int64, status string) (err error) {
	db := DB()
	glusterfs := &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Model(glusterfs).Update("status", status).Error
	if err != nil {
		log.Println("Failed to update glusterfs cluster status", err)
		return
	}
	return
}

func (a *GlusterfsAdmin) GetState(ctx context.Context, id int64) (status string, err error) {
	db := DB()
	glusterfs := &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Take(glusterfs).Error
	if err != nil {
		log.Println("Failed to update glusterfs cluster status", err)
		return
	}
	status = glusterfs.Status
	return
}

func (a *GlusterfsAdmin) Update(ctx context.Context, id, flavorID int64, nworkers int32) (glusterfs *model.Glusterfs, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	glusterfs = &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Take(glusterfs).Error
	if err != nil {
		log.Println("DB failed to query glusterfs", err)
		return
	}
	if flavorID != glusterfs.Flavor {
		flavor := &model.Flavor{Model: model.Model{ID: flavorID}}
		if err = db.Take(flavor).Error; err != nil {
			log.Println("Failed to query flavor", err)
			return
		}
		glusterfs.Flavor = flavorID
		if err = db.Save(glusterfs).Error; err != nil {
			log.Println("Failed to save glusterfs", err)
			return
		}
	}
	err = a.State(ctx, id, "updating")
	if err != nil {
		log.Println("DB failed to update cluster status", err)
		return
	}
	maxIndex := 0
	if glusterfs.WorkerNum > 0 {
		instances := []*model.Instance{}
		err = db.Where("subnet_id = ? and hostname like ?", glusterfs.SubnetID, "%gluster%").Find(&instances).Error
		if err != nil {
			log.Println("Failed to query cluster instances", err)
			return
		}
		if len(instances) > 0 {
			for _, inst := range instances {
				name := strings.Split(inst.Hostname, "-")
				if len(name) < 2 {
					log.Println("Wrong name pattern")
					continue
				}
				index, err := strconv.Atoi(name[1])
				if err != nil {
					log.Println("Failed to convert index")
					continue
				}
				if maxIndex < index {
					maxIndex = index
				}
			}
		}
	}
	if nworkers > glusterfs.WorkerNum {
		for i := 0; i < int(nworkers-glusterfs.WorkerNum); i++ {
			maxIndex++
			hostname := fmt.Sprintf("gluster-%d", maxIndex)
			ipaddr := fmt.Sprintf("192.168.91.%d", maxIndex+200)
			secgroup := &model.SecurityGroup{Model: model.Model{Owner: memberShip.OrgID}, Name: "gluster"}
			err = db.Where(secgroup).Take(secgroup).Error
			if err != nil {
				log.Println("No existing gluster security group", err)
				return
			}
			sgIDs := []int64{secgroup.ID}
			keyIDs := []int64{glusterfs.Key, glusterfs.HeketiKey}
			_, err = instanceAdmin.Create(ctx, 1, hostname, "", 1, int64(flavorID), glusterfs.SubnetID, 0, ipaddr, "", nil, keyIDs, sgIDs, -1)
			if err != nil {
				log.Println("Failed to launch a worker", err)
				return
			}
		}
	} else {
		for i := 0; i < int(glusterfs.WorkerNum-nworkers); i++ {
			hostname := fmt.Sprintf("gluster-%d", maxIndex)
			instance := &model.Instance{}
			err = db.Where("hostname = ? and subnet_id = ?", hostname, glusterfs.SubnetID).Take(instance).Error
			if err != nil {
				log.Println("Failed to query worker", err)
				return
			}
			err = instanceAdmin.Delete(ctx, instance.ID)
			if err != nil {
				log.Println("Failed to delete worker", err)
				return
			}
			maxIndex--
		}
	}
	err = a.State(ctx, id, "complete")
	if err != nil {
		log.Println("DB failed to update cluster status", err)
		return
	}
	return
}

func (a *GlusterfsAdmin) Create(ctx context.Context, name, cookie string, nworkers int32, cluster, flavor, key int64) (glusterfs *model.Glusterfs, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	var subnet *model.Subnet
	if cluster > 0 {
		glusterfs = &model.Glusterfs{Model: model.Model{ID: cluster}}
		err = db.Take(glusterfs).Error
		if err != nil {
			log.Println("DB failed to query glusterfs", err)
			return
		}
		subnet = glusterfs.Subnet
	} else {
		subnet, err = subnetAdmin.Create(ctx, "gluster-sn", "", "192.168.91.0", "255.255.255.0", "", "", "", "", "", "", "", 0, memberShip.OrgID)
		if err != nil {
			log.Println("Failed to create glusterfs subnet", err)
			return
		}
		subnetIDs := []int64{subnet.ID}
		_, err = gatewayAdmin.Create(ctx, "gatewayAdmin", "", 0, 0, subnetIDs, memberShip.OrgID)
		if err != nil {
			log.Println("Failed to create gateway", err)
			return
		}
	}
	secgroup, err := a.createSecgroup(ctx, "gluster", "192.168.91.0/24", memberShip.OrgID)
	keyIDs := []int64{key}
	sgIDs := []int64{secgroup.ID}
	endpoint := viper.GetString("api.endpoint")
	userdata := `#!/bin/bash
cd /opt
exec >/tmp/heketi.log 2>&1
yum -y install epel-release centos-release-gluster
yum -y install wget jq`
	userdata = fmt.Sprintf("%s\nwget '%s/misc/glusterfs/heketi.sh'\nchmod +x heketi.sh", userdata, endpoint)
	userdata = fmt.Sprintf("%s\n./heketi.sh '%s' '%s' '%d' '%d'", userdata, endpoint, cookie, subnet.ID, nworkers)
	_, err = instanceAdmin.Create(ctx, 1, "heketi", userdata, 1, flavor, subnet.ID, cluster, "192.168.91.199", "", nil, keyIDs, sgIDs, -1)
	if err != nil {
		log.Println("Failed to create heketi instance", err)
		return
	}
	glusterfs = &model.Glusterfs{
		Model:    model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID},
		Status:   "creating",
		Flavor:   flavor,
		Key:      key,
		SubnetID: subnet.ID,
		Endpoint: "http://192.168.91.199:8080",
	}
	err = db.Create(glusterfs).Error
	return
}

func (a *GlusterfsAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	glusterfs := &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Set("gorm:auto_preload", true).Take(glusterfs).Error
	if err != nil {
		log.Println("Failed to query glusterfs cluster", err)
		return
	}
	subnet := glusterfs.Subnet
	if subnet != nil {
		if subnet.Router != 0 {
			err = gatewayAdmin.Delete(ctx, subnet.Router)
			if err != nil {
				log.Println("Failed to delete gateway", err)
				return
			}
		}
		err = subnetAdmin.Delete(ctx, subnet.ID)
		if err != nil {
			log.Println("Failed to delete subnet", err)
			return
		}
	}
	if err = db.Delete(&model.Glusterfs{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *GlusterfsAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, glusterfses []*model.Glusterfs, err error) {
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
	glusterfses = []*model.Glusterfs{}
	if err = db.Model(&model.Glusterfs{}).Where(where).Where(query).Count(&total).Error; err != nil {
		return
	}
	db = dbs.Sortby(db.Offset(offset).Limit(limit), order)
	if err = db.Where(where).Where(query).Find(&glusterfses).Error; err != nil {
		return
	}

	return
}

func (v *GlusterfsView) List(c *macaron.Context, store session.Store) {
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
	total, glusterfses, err := glusterfsAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Glusterfses"] = glusterfses
	c.Data["Total"] = total
	c.Data["Pages"] = GetPages(total, limit)
	c.Data["Query"] = query
	c.HTML(200, "glusterfs")
}

func (v *GlusterfsView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "glusterfs", id)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	err = glusterfsAdmin.Delete(c.Req.Context(), id)
	if err != nil {
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "glusterfs",
	})
	return
}

func (v *GlusterfsView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "glusterfs", id)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	db := DB()
	glusterfs := &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Take(glusterfs).Error
	if err != nil {
		log.Println("Failed ro query glusterfs", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	flavors := []*model.Flavor{}
	if err := db.Where("ephemeral > 0").Where("ephemeral > 0").Find(&flavors).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Glusterfs"] = glusterfs
	c.Data["Flavors"] = flavors
	c.HTML(200, "glusterfs_patch")
}

func (v *GlusterfsView) Patch(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "glusterfs", id)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	flavor := c.QueryInt64("flavor")
	nworkers := c.QueryInt("nworkers")
	if nworkers < 3 {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Number of worker must be at least 2"
		c.HTML(code, "error")
		return
	}
	status, err := glusterfsAdmin.GetState(ctx, id)
	if status != "complete" {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Cluster can be updated only in complete status"
		c.HTML(code, "error")
		return
	}
	_, err = glusterfsAdmin.Update(ctx, id, flavor, int32(nworkers))
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	c.Redirect("../glusterfs")
}

func (v *GlusterfsView) New(c *macaron.Context, store session.Store) {
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
	flavors := []*model.Flavor{}
	if err := db.Where("ephemeral > 0").Find(&flavors).Error; err != nil {
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
	_, openshifts, err := openshiftAdmin.List(ctx, 0, -1, "", "")
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Flavors"] = flavors
	c.Data["Keys"] = keys
	c.Data["Openshifts"] = openshifts
	c.HTML(200, "glusterfs_new")
}

func (v *GlusterfsView) State(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "glusterfs", id)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	status := c.QueryTrim("status")
	err = glusterfsAdmin.State(c.Req.Context(), id, status)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})
	}
	c.JSON(200, "ack")
}

func (v *GlusterfsView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	redirectTo := "../glusterfs"
	name := c.QueryTrim("name")
	if name == "" {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Name can not be empty string"
		c.HTML(code, "error")
		return
	}
	nworkers := c.QueryInt("nworkers")
	if nworkers < 3 {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Number of worker must be at least 2"
		c.HTML(code, "error")
		return
	}
	flavor := c.QueryInt64("flavor")
	key := c.QueryInt64("key")
	cluster := c.QueryInt64("cluster")
	cookie := "MacaronSession=" + c.GetCookie("MacaronSession")
	_, err := glusterfsAdmin.Create(c.Req.Context(), name, cookie, int32(nworkers), cluster, flavor, key)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "error")
		return
	}
	c.Redirect(redirectTo)
}
