/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/src/model"
	"github.com/IBM/cloudland/web/src/dbs"
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
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "tcp", 8080, 8080)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "tcp", 2379, 2380)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "tcp", 2049, 2049)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "tcp", 9000, 9999)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "tcp", 10249, 10259)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "udp", 9000, 9999)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "udp", 4789, 4789)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "udp", 2049, 2049)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "udp", 6081, 6081)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	_, err = secruleAdmin.Create(ctx, secgroup.ID, owner, cidr, "ingress", "udp", 30000, 32767)
	if err != nil {
		log.Println("Failed to create security rule", err)
		return
	}
	return
}

func (a *OpenshiftAdmin) State(ctx context.Context, id int64, status string) (err error) {
	db := DB()
	openshift := &model.Openshift{Model: model.Model{ID: id}}
	err = db.Model(openshift).Update("status", status).Error
	if err != nil {
		log.Println("Failed to update openshift cluster status", err)
		return
	}
	return
}

func (a *OpenshiftAdmin) GetState(ctx context.Context, id int64) (status string, err error) {
	db := DB()
	openshift := &model.Openshift{Model: model.Model{ID: id}}
	err = db.Take(openshift).Error
	if err != nil {
		log.Println("Failed to update openshift cluster status", err)
		return
	}
	status = openshift.Status
	return
}

func (a *OpenshiftAdmin) Launch(ctx context.Context, id int64, hostname, ipaddr string) (instance *model.Instance, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	openshift := &model.Openshift{Model: model.Model{ID: id}}
	if err = db.Preload("Subnet").Preload("Subnet.Netlink").Take(openshift).Error; err != nil {
		log.Println("Failed to query openshift cluster", err)
		return
	}
	flavorID := openshift.MasterFlavor
	if strings.Contains(hostname, "worker") {
		flavorID = openshift.WorkerFlavor
	}
	flavor := &model.Flavor{Model: model.Model{ID: flavorID}}
	if err = db.Take(flavor).Error; err != nil {
		log.Println("Failed to query flavor", err)
		return
	}
	subnet := openshift.Subnet
	if subnet == nil {
		log.Println("Cluster has no built-in subnet")
		err = fmt.Errorf("Cluster has no built-in subnet")
		return
	}
	secgroup := &model.SecurityGroup{Model: model.Model{Owner: memberShip.OrgID}, Name: "openshift"}
	err = db.Where(secgroup).Take(secgroup).Error
	if err != nil {
		log.Println("No existing openshift security group", err)
		return
	}
	secGroups := []*model.SecurityGroup{secgroup}
	instance = &model.Instance{Model: model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID}, Hostname: hostname, FlavorID: flavorID, Status: "pending", ZoneID: openshift.ZoneID}
	err = db.Create(instance).Error
	if err != nil {
		log.Println("DB create instance failed", err)
		return
	}
	metadata := ""
	var ifaces []*model.Interface
	ifaces, metadata, err = instanceAdmin.buildMetadata(ctx, subnet, "", "", nil, nil, instance, "", secGroups, openshift.ZoneID, ipaddr)
	if err != nil {
		log.Println("Build instance metadata failed", err)
		return
	}
	instance.Interfaces = ifaces
	count := 0
	err = db.Model(&model.Instance{}).Where("cluster_id = ? and hostname like ?", id, "%worker%").Count(&count).Error
	if err != nil {
		return
	}
	openshift.WorkerNum = int32(count)
	rcNeeded := fmt.Sprintf("cpu=%d memory=%d disk=%d network=%d", flavor.Cpu, flavor.Memory*1024, flavor.Disk*1024*1024, 0)
	hyperGroup, err := instanceAdmin.getHyperGroup("", openshift.ZoneID)
	control := "select=" + hyperGroup + " " + rcNeeded
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/oc_vm.sh '%d' '%d' '%d' '%d' '%s'<<EOF\n%s\nEOF", instance.ID, flavor.Cpu, flavor.Memory, flavor.Disk, hostname, metadata)
	err = hyperExecute(ctx, control, command)
	if err != nil {
		log.Println("Launch vm command execution failed", err)
		return
	}
	if strings.Contains(hostname, "worker") {
		err = db.Save(openshift).Error
		if err != nil {
			log.Println("Failed to update openshift cluster")
			return
		}
	}
	return
}

func (a *OpenshiftAdmin) Update(ctx context.Context, id, flavorID int64, nworkers int32) (openshift *model.Openshift, err error) {
	db := DB()
	openshift = &model.Openshift{Model: model.Model{ID: id}}
	err = db.Take(openshift).Error
	if err != nil {
		log.Println("DB failed to query openshift", err)
		return
	}
	if flavorID != openshift.WorkerFlavor {
		flavor := &model.Flavor{Model: model.Model{ID: flavorID}}
		if err = db.Take(flavor).Error; err != nil {
			log.Println("Failed to query flavor", err)
			return
		}
		openshift.WorkerFlavor = flavorID
		if err = db.Save(openshift).Error; err != nil {
			log.Println("Failed to save openshift", err)
			return
		}
	}
	err = a.State(ctx, id, "updating")
	if err != nil {
		log.Println("DB failed to update cluster status", err)
		return
	}
	maxIndex := 0
	if openshift.WorkerNum > 0 {
		instances := []*model.Instance{}
		err = db.Where("cluster_id = ? and hostname like ?", id, "%worker%").Find(&instances).Error
		if err != nil {
			log.Println("Failed to query cluster instances", err)
			return
		}
		if len(instances) > 0 {
			for _, inst := range instances {
				prefix := strings.Split(inst.Hostname, ".")[0]
				name := strings.Split(prefix, "-")
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
	if nworkers > openshift.WorkerNum {
		for i := 0; i < int(nworkers-openshift.WorkerNum); i++ {
			maxIndex++
			hostname := fmt.Sprintf("worker-%d.%s.%s", maxIndex, openshift.ClusterName, openshift.BaseDomain)
			lb :=strings.Split(openshift.LoadBalancer, `/`)
			_, err = openshiftAdmin.Launch(ctx, id, hostname, lb[0])
			if err != nil {
				log.Println("Failed to launch a worker", err)
				return
			}
		}
	} else {
		for i := 0; i < int(openshift.WorkerNum-nworkers); i++ {
			hostname := fmt.Sprintf("worker-%d", maxIndex)
			fqdn1 := hostname + "." + openshift.ClusterName + "." + openshift.BaseDomain
			fqdn2 := hostname + "." + openshift.BaseDomain
			instance := &model.Instance{}
			err = db.Where("(hostname = ? or hostname = ? or hostname = ?) and cluster_id = ?", hostname, fqdn1, fqdn2, id).Take(instance).Error
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

func (a *OpenshiftAdmin) Create(ctx context.Context, cluster, domain, cookie, haflag, extIP string, nworkers int32, lflavor, mflavor, wflavor, key, zoneID, subnetID, registryID int64, hostrec, infrtype string) (openshift *model.Openshift, err error) {
	memberShip := GetMemberShip(ctx)
	log.Println("it's in Creating openshift")
	db := DB()
	log.Println("registry in  openshift")
	registry := model.Registry{}
	if err = db.First(&registry, registryID).Error; err != nil {
		return
	}
	registryContent := registry.RegistryContent
	log.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~registryContent==" + registryContent)
	log.Println("version in openshift")
	version := registry.OcpVersion
	log.Println("version=%s", version)

	lbIP := ""
	log.Println("subnet in openshift")
	subnet := &model.Subnet{Model: model.Model{ID: subnetID}}
	if subnetID == 0 {
		lbIP = "192.168.91.8"
		subnetname := cluster + "-sn"
		search := cluster + "." + domain
		zone := fmt.Sprintf("%d", zoneID)
		subnet, err = subnetAdmin.Create(ctx, subnetname, "", "192.168.91.0", "255.255.255.0", zone, "", "", "", "", lbIP, search, "yes", "", "", memberShip.OrgID)
		if err != nil {
			log.Println("Failed to create openshift subnet", err)
			return
		}
		routerName := cluster + "-gw"
		_, err = routerAdmin.Create(ctx, routerName, "", 0, memberShip.OrgID, zoneID)
		if err != nil {
			log.Println("Failed to create router", err)
			return
		}
	} else {
		if err = db.Take(subnet).Error; err != nil {
			log.Println("Subnet query failed", err)
			return
		}
	}
	log.Println("openshift = &model.Openshift")
	openshift = &model.Openshift{
		Model:              model.Model{Creater: memberShip.UserID, Owner: memberShip.OrgID},
		ClusterName:        cluster,
		BaseDomain:         domain,
		Status:             "creating",
		Haflag:             haflag,
		Version:            version,
		Flavor:             lflavor,
		MasterFlavor:       mflavor,
		WorkerFlavor:       wflavor,
		SubnetID:           subnet.ID,
		Key:                key,
		ZoneID:             zoneID,
		InfrastructureType: infrtype,
	}
	log.Println("~~~~creating~~~")
	err = db.Create(openshift).Error
	if err != nil {
		log.Println("DB failed to create openshift", err)
		return
	}
	preSize, _ := net.IPMask(net.ParseIP(subnet.Netmask).To4()).Size()
	cidr := fmt.Sprintf("%s/%d", subnet.Network, preSize)
	secgroup, err := a.createSecgroup(ctx, "openshift", cidr, memberShip.OrgID)
	lbname := "lb"
	keyIDs := []int64{key}
	sgIDs := []int64{secgroup.ID}
	endpoint := viper.GetString("api.endpoint")
	userdata := getUserdata("ocd")
	userdata = fmt.Sprintf("%s\ncurl -k -O '%s/misc/openshift/ocd.sh'\nchmod +x ocd.sh", userdata, endpoint)
	/*
		        parts := fmt.Sprintf("pullSecret: '%s'\n", secret)
			if atbundle != "" {
				parts = fmt.Sprintf("%s\n%s\n", parts, atbundle)
			}
			if icsources != "" {
				parts = fmt.Sprintf("%s\n%s\n", parts, icsources)
			}
	*/
	parts := fmt.Sprintf("\n%s\n", registryContent)
	encParts := base64.StdEncoding.EncodeToString([]byte(parts))
	infraType := openshift.InfrastructureType
	log.Println("invoking ocd.sh")
	userdata = fmt.Sprintf("%s\n./ocd.sh '%d' '%s' '%s' '%s' '%s' '%s' '%d' '%s' '%s' '%s' '%s'<<EOF\n%s\nEOF", userdata, openshift.ID, cluster, domain, endpoint, cookie, haflag, nworkers, version, infraType, extIP, hostrec, encParts)
	image := &model.Image{}
	err = db.Where("open_shift_lb = ? and virt_type = ?", true, infraType).Take(image).Error
	if err != nil {
		log.Println("No valid LB image exists", err)
		return
	}
	lbImg := image.ID
	instance, err := instanceAdmin.Create(ctx, 1, lbname, userdata, int64(lbImg), lflavor, subnet.ID, zoneID, lbIP, "", nil, keyIDs, sgIDs, -1)
	if err != nil {
		log.Println("Failed to create oc first instance", err)
		return
	}
	openshift.LoadBalancer = instance.Interfaces[0].Address.Address
	err = db.Save(openshift).Error
	if err != nil {
		log.Println("Failed to update openshift cluster")
		return
	}
	return
}

func (a *OpenshiftAdmin) Delete(ctx context.Context, id int64) (err error) {
	db := DB()
	db = db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()
	openshift := &model.Openshift{Model: model.Model{ID: id}}
	err = db.Set("gorm:auto_preload", true).Take(openshift).Error
	if err != nil {
		log.Println("Failed to query openshift cluster", err)
		return
	}
	if openshift.Instances != nil && len(openshift.Instances) > 0 {
		log.Println("There are instances in this cluster")
		err = fmt.Errorf("There are instances in this cluster")
		return
	}
	subnet := openshift.Subnet
	if subnet != nil && subnet.Type == "internal" && subnet.Name == openshift.ClusterName+"-sn" && subnet.Network == "192.168.91.0" {
		if subnet.RouterID != 0 {
			err = routerAdmin.Delete(ctx, subnet.RouterID)
			if err != nil {
				log.Println("Failed to delete router", err)
				return
			}
		}
		err = subnetAdmin.Delete(ctx, subnet.ID)
		if err != nil {
			log.Println("Failed to delete subnet", err)
			return
		}
	}
	if err = db.Delete(&model.Openshift{Model: model.Model{ID: id}}).Error; err != nil {
		return
	}
	return
}

func (a *OpenshiftAdmin) List(ctx context.Context, offset, limit int64, order, query string) (total int64, openshifts []*model.Openshift, err error) {
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
	total, openshifts, err := openshiftAdmin.List(c.Req.Context(), offset, limit, order, query)
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
	c.Data["Openshifts"] = openshifts
	c.Data["Total"] = total
	c.Data["Pages"] = pages
	c.Data["Query"] = query
	if c.Req.Header.Get("X-Json-Format") == "yes" {
		c.JSON(200, map[string]interface{}{
			"openshifts": openshifts,
			"total":      total,
			"pages":      pages,
			"query":      query,
		})
		return
	}
	c.HTML(200, "openshifts")
}

func (v *OpenshiftView) Delete(c *macaron.Context, store session.Store) (err error) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	err = openshiftAdmin.Delete(c.Req.Context(), id)
	if err != nil {
		log.Println("Failed to Delete Openshift~")
		err = fmt.Errorf("There are instances in this cluster,Failed to delete")
		return
	}
	c.JSON(200, map[string]interface{}{
		"redirect": "openshifts",
	})
	return
}

func (v *OpenshiftView) Edit(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	db := DB()
	openshift := &model.Openshift{Model: model.Model{ID: id}}
	err = db.Take(openshift).Error
	if err != nil {
		log.Println("Failed ro query openshift", err)
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
	}
	_, flavors, err := flavorAdmin.List(0, -1, "", "")
	if err := db.Find(&flavors).Error; err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Openshift"] = openshift
	c.Data["Flavors"] = flavors
	c.HTML(200, "openshifts_patch")
}

func (v *OpenshiftView) Patch(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	flavor := c.QueryInt64("flavor")
	nworkers := c.QueryInt("nworkers")
	if nworkers < 2 {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Number of worker must be at least 2"
		c.HTML(code, "error")
		return
	}
	openshift, err := openshiftAdmin.Update(ctx, id, flavor, int32(nworkers))
	if err != nil {
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
		c.JSON(200, openshift)
		return
	}
	c.Redirect("../openshifts")
}

func (v *OpenshiftView) New(c *macaron.Context, store session.Store) {
	log.Println("go to openshiftView in New func")
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	log.Println("starting to connect to DB in openshift")
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
	log.Println("go to openshiftView in New Func ~~~~")
	sql := fmt.Sprintf("type = 'public' or owner = %d", memberShip.OrgID)
	_, subnets, err := subnetAdmin.List(ctx, 0, -1, "", "", sql)
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	zones := []*model.Zone{}
	err = db.Find(&zones).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	registrys := []*model.Registry{}
	err = db.Find(&registrys).Error
	if err != nil {
		c.Data["ErrorMsg"] = err.Error()
		c.HTML(500, "500")
		return
	}
	c.Data["Flavors"] = flavors
	c.Data["Keys"] = keys
	c.Data["Subnets"] = subnets
	c.Data["Zones"] = zones
	c.Data["Registrys"] = registrys
	c.HTML(200, "openshifts_new")
}

func (v *OpenshiftView) State(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	status := c.QueryTrim("status")
	err = openshiftAdmin.State(c.Req.Context(), id, status)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})
	}
	c.JSON(200, "ack")
}

func (v *OpenshiftView) Launch(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.ParamsInt64("id")
	permit, err := memberShip.CheckOwner(model.Owner, "openshifts", id)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	hostname := c.QueryTrim("hostname")
	ipaddr := c.QueryTrim("ipaddr")
	log.Println("------------------------------------------------------")
	log.Println("ipaddr===" + ipaddr)
	instance, err := openshiftAdmin.Launch(c.Req.Context(), id, hostname, ipaddr)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})
	}
	c.JSON(200, instance)
}

func (v *OpenshiftView) Create(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	permit := memberShip.CheckPermission(model.Owner)
	if !permit {
		log.Println("Not authorized for this operation")
		c.Data["ErrorMsg"] = "Not authorized for this operation"
		c.HTML(http.StatusBadRequest, "error")
		return
	}
	log.Println("starting to redirect to openshift")
	redirectTo := "../openshifts"
	name := c.QueryTrim("clustername")
	domain := c.QueryTrim("basedomain")
	haflag := c.QueryTrim("haflag")
	if haflag == "" {
		haflag = "no"
	}
	//secret := c.QueryTrim("secret")
	hostrec := c.QueryTrim("hostrec")
	nworkers := c.QueryInt("nworkers")
	if nworkers < 2 {
		code := http.StatusBadRequest
		c.Data["ErrorMsg"] = "Number of worker must be at least 2"
		c.HTML(code, "error")
		return
	}
	registry := c.QueryInt64("registry")
	//version := c.QueryTrim("version")
	extIP := c.QueryTrim("extip")
	lflavor := c.QueryInt64("lflavor")
	mflavor := c.QueryInt64("mflavor")
	wflavor := c.QueryInt64("wflavor")
	subnet := c.QueryInt64("subnet")
	key := c.QueryInt64("key")
	zone := c.QueryInt64("zone")
	infrtype := c.QueryTrim("infrtype")
	//sback := c.QueryTrim("sback")
	//atbundle := c.QueryTrim("atbundle")
	//icsources := c.QueryTrim("icsources")
	cookie := "MacaronSession=" + c.GetCookie("MacaronSession")
	permit, err := memberShip.CheckOwner(model.Writer, "subnets", int64(subnet))
	log.Println("openshift create in viewCreate")
	openshift, err := openshiftAdmin.Create(c.Req.Context(), name, domain, cookie, haflag, extIP, int32(nworkers), lflavor, mflavor, wflavor, key, zone, subnet, registry, hostrec, infrtype)
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
		c.JSON(200, openshift)
		return
	}
	c.Redirect(redirectTo)
}
