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

func (a *GlusterfsAdmin) State(ctx context.Context, id int64, status string, nworkers int32) (err error) {
	db := DB()
	glusterfs := &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Model(glusterfs).Update(map[string]interface{}{"status": status, "worker_num": nworkers}).Error
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

func (a *GlusterfsAdmin) Update(ctx context.Context, id, heketiKey, flavorID int64, nworkers int32) (glusterfs *model.Glusterfs, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	glusterfs = &model.Glusterfs{Model: model.Model{ID: id}}
	err = db.Take(glusterfs).Error
	if err != nil {
		log.Println("DB failed to query glusterfs", err)
		return
	}
	if flavorID > 0 && flavorID != glusterfs.Flavor {
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
	if glusterfs.HeketiKey == 0 && heketiKey > 0 {
		glusterfs.HeketiKey = heketiKey
		if err = db.Save(glusterfs).Error; err != nil {
			log.Println("Failed to save glusterfs", err)
			return
		}
	}
	err = a.State(ctx, id, "updating", glusterfs.WorkerNum)
	if err != nil {
		log.Println("DB failed to update cluster status", err)
		return
	}
	maxIndex := glusterfs.WorkerNum - 1
	if nworkers > glusterfs.WorkerNum {
		for i := 0; i < int(nworkers-glusterfs.WorkerNum); i++ {
			maxIndex++
			hostname := fmt.Sprintf("g%d-gluster-%d", glusterfs.ID, maxIndex)
			ipaddr := fmt.Sprintf("192.168.91.%d", maxIndex+200)
			secgroup := &model.SecurityGroup{Model: model.Model{Owner: memberShip.OrgID}, Name: "gluster"}
			err = db.Where(secgroup).Take(secgroup).Error
			if err != nil {
				log.Println("No existing gluster security group", err)
				return
			}
			endpoint := viper.GetString("api.endpoint")
			userdata := getUserdata("gluster")
			userdata = fmt.Sprintf("%s\ncurl -k -O '%s/misc/glusterfs/gluster.sh'\nchmod +x gluster.sh", userdata, endpoint)
			userdata = fmt.Sprintf("%s\n./gluster.sh '%d' '%s'", userdata, glusterfs.ID, glusterfs.Endpoint)
			sgIDs := []int64{secgroup.ID}
			keyIDs := []int64{glusterfs.Key, glusterfs.HeketiKey}
			_, err = instanceAdmin.Create(ctx, 1, hostname, userdata, 1, glusterfs.Flavor, glusterfs.SubnetID, glusterfs.ClusterID, ipaddr, "", nil, keyIDs, sgIDs, -1)
			if err != nil {
				log.Println("Failed to launch a worker", err)
				return
			}
		}
	} else {
		for i := 0; i < int(glusterfs.WorkerNum-nworkers); i++ {
			hostname := fmt.Sprintf("g%d-gluster-%d", glusterfs.ID, maxIndex)
			instance := &model.Instance{}
			err = db.Preload("Interface").Where("hostname = ?", hostname).Take(instance).Error
			if err != nil {
				log.Println("Failed to query gluster worker", err)
				return
			}
			if instance.Interfaces == nil || len(instance.Interfaces) == 0 || instance.Interfaces[0].Subnet != glusterfs.SubnetID {
				log.Println("Failed to query gluster worker", err)
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
	glusterfs.WorkerNum = nworkers
	err = a.State(ctx, id, "complete", glusterfs.WorkerNum)
	if err != nil {
		log.Println("DB failed to update cluster status", err)
		return
	}
	return
}

func getUserdata(name string) (userdata string) {
	userdata = fmt.Sprintf("#!/bin/bash\nexec >/tmp/%s.log 2>&1\n", name)
	userdata += `cd /opt
count=0
while [ "$count" -le 20 ]; do
    sleep 10
    nameserver=$(grep '^nameserver' /etc/resolv.conf | head -1 | awk '{print $2}')
    [ -n "$nameserver" ] && break
    let count=$count+1
done
[ -z "$nameserver" ] && nameserver=8.8.8.8 && echo nameserver $nameserver >> /etc/resolv.conf
while true; do
    ping -c 1 $nameserver
    [ $? -eq 0 ] && break
done
`
	return
}

func (a *GlusterfsAdmin) Create(ctx context.Context, name, cookie string, nworkers int32, cluster, flavor, key int64) (glusterfs *model.Glusterfs, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	glusterfs = &model.Glusterfs{
		Model:     model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID},
		Name:      name,
		Status:    "creating",
		Flavor:    flavor,
		Key:       key,
		ClusterID: cluster,
		Endpoint:  "http://192.168.91.199:8080",
	}
	err = db.Create(glusterfs).Error
	if err != nil {
		log.Println("Failed to create glusterfs", err)
		return
	}
	var subnet *model.Subnet
	var openshift *model.Openshift
	if cluster > 0 {
		openshift = &model.Openshift{Model: model.Model{ID: cluster}}
		err = db.Preload("Subnet").Take(openshift).Error
		if err != nil {
			log.Println("DB failed to query openshift cluster", err)
			return
		}
		subnet = openshift.Subnet
	} else {
		tmpName := fmt.Sprintf("g%d-sn", glusterfs.ID)
		subnet, err = subnetAdmin.Create(ctx, tmpName, "", "192.168.91.0", "255.255.255.0", "", "", "", "", "", "", "yes", "", 0, memberShip.OrgID)
		if err != nil {
			log.Println("Failed to create glusterfs subnet", err)
			return
		}
		subnetIDs := []int64{subnet.ID}
		tmpName = fmt.Sprintf("g%d-gw", glusterfs.ID)
		_, err = gatewayAdmin.Create(ctx, tmpName, "", 0, 0, subnetIDs, memberShip.OrgID)
		if err != nil {
			log.Println("Failed to create gateway", err)
			return
		}
	}
	secgroup, err := a.createSecgroup(ctx, "gluster", "192.168.91.0/24", memberShip.OrgID)
	keyIDs := []int64{key}
	sgIDs := []int64{secgroup.ID}
	endpoint := viper.GetString("api.endpoint")
	userdata := getUserdata("heketi")
	userdata = fmt.Sprintf("%s\ncurl -k -O '%s/misc/glusterfs/heketi.sh'\nchmod +x heketi.sh", userdata, endpoint)
	userdata = fmt.Sprintf("%s\n./heketi.sh '%d' '%s' '%s' '%d' '%d'", userdata, glusterfs.ID, endpoint, cookie, subnet.ID, nworkers)
	tmpName := fmt.Sprintf("g%d-heketi", glusterfs.ID)
	_, err = instanceAdmin.Create(ctx, 1, tmpName, userdata, 1, flavor, subnet.ID, cluster, "192.168.91.199", "", nil, keyIDs, sgIDs, -1)
	if err != nil {
		log.Println("Failed to create heketi instance", err)
		return
	}
	glusterfs.SubnetID = subnet.ID
	err = db.Save(glusterfs).Error
	if err != nil {
		log.Println("Failed to create glusterfs", err)
		return
	}
	if openshift != nil {
		openshift.GlusterID = glusterfs.ID
	}
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
	if glusterfs.HeketiKey > 0 {
		err = keyAdmin.Delete(glusterfs.HeketiKey)
		if err != nil {
			log.Println("Failed to delete heketi key", err)
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
		limit = 16
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
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	offset := c.QueryInt64("offset")
	limit := c.QueryInt64("limit")
	if limit == 0 {
		limit = 16
	}
	order := c.Query("order")
	if order == "" {
		order = "-created_at"
	}
	query := c.QueryTrim("q")
	total, glusterfses, err := glusterfsAdmin.List(c.Req.Context(), offset, limit, order, query)
	if err != nil {
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
	c.Data["Glusterfses"] = glusterfses
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"glusterfses": glusterfses,
			"total":       total,
			"pages":       pages,
			"query":       query,
		})
		return
	}
	c.HTML(200, "glusterfs")
}

func (v *GlusterfsView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "glusterfs", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = glusterfsAdmin.Delete(c.Req.Context(), id)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(http.StatusBadRequest, "error")
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
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
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
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor := c.QueryInt64("flavor")
	heketikey := c.QueryInt64("heketikey")
	nworkers := c.QueryInt("nworkers")
	if nworkers < 3 {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Number of workers must be at least 3"
		c.HTML(code, "error")
		return
	}
	glusterfs, err := glusterfsAdmin.Update(ctx, id, heketikey, flavor, int32(nworkers))
	if err != nil {
		log.Println("Failed to create glusterfs", err)
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
		c.JSON(200, glusterfs)
		return
	}
	c.Redirect("../glusterfs")
}

func (v *GlusterfsView) New(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
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
	where := memberShip.GetWhere()
	openshifts := []*model.Openshift{}
	err = db.Where(where).Where("gluster_id = 0").Find(&openshifts).Error
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
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	status := c.QueryTrim("status")
	err = glusterfsAdmin.State(c.Req.Context(), id, status, 0)
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
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
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
		c.Data["ErrorMsg"] = "Number of workers must be at least 3"
		c.HTML(code, "error")
		return
	}
	flavor := c.QueryInt64("flavor")
	if flavor <= 0 {
		log.Println("Invalid flavor ID")
		c.Data["ErrorMsg"] = "Invalid flavor ID"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	key := c.QueryInt64("key")
	permit, err := memberShip.CheckOwner(model.Writer, "keys", key)
	if !permit {
		log.Println("Not authorized to access key")
		c.Data["ErrorMsg"] = "Not authorized to access key"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	cluster := c.QueryInt64("cluster")
	if cluster < 0 {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Openshift cluster must be >= 0"
		c.HTML(code, "error")
		return
	}
	permit, err = memberShip.CheckOwner(model.Writer, "openshifts", cluster)
	if !permit {
		log.Println("Not authorized to access openshift cluser")
		c.Data["ErrorMsg"] = "Not authorized to access openshift cluser"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	cookie := "MacaronSession=" + c.GetCookie("MacaronSession")
	glusterfs, err := glusterfsAdmin.Create(c.Req.Context(), name, cookie, int32(nworkers), cluster, flavor, key)
	if err != nil {
		if c.Req.Header.Get("X-Json-Format") == "yes" {
			c.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "error")
		return
	} else if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, glusterfs)
		return
	}
	c.Redirect(redirectTo)
}
