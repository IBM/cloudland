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
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
)

var (
	dashboard = &Dashboard{}
)

type ResourceData struct {
	Title       string `json:"title"`
	CpuUsed     int64  `json:"cpu_used"`
	CpuAvail    int64  `json:"cpu_avail"`
	MemUsed     int64  `json:"mem_used"`
	MemAvail    int64  `json:"mem_avail"`
	DiskUsed    int64  `json:"disk_used"`
	DiskAvail   int64  `json:"disk_avail"`
	VolumeUsed  int64  `json:"volume_used"`
	VolumeAvail int64  `json:"volume_avail"`
	PubipUsed   int64  `json:"pubip_used"`
	PubipAvail  int64  `json:"pubip_avail"`
	PrvipUsed   int64  `json:"prvip_used"`
	PrvipAvail  int64  `json:"prvip_avail"`
}

type Dashboard struct{}

func (a *Dashboard) Show(c *macaron.Context, store session.Store) {
	c.HTML(200, "dashboard")
	return
}

func (a *Dashboard) GetData(c *macaron.Context, store session.Store) {
	ctx := c.Req.Context()
	memberShip := GetMemberShip(ctx)
	var rcData *ResourceData
	db := DB()
	if memberShip.OrgName == "admin" {
		resource := &model.Resource{}
		err := db.Where("hostid = ?", -1).Take(resource).Error
		if err != nil {
			log.Println("Failed to query system resource")
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
		pubipTotal, pubipUsed, err := a.getSystemIpUsage(ctx, "public")
		prvipTotal, prvipUsed, err := a.getSystemIpUsage(ctx, "private")
		rcData = &ResourceData{
			Title:       "System Resource Usage Ratio",
			CpuUsed:     resource.CpuTotal - resource.Cpu,
			CpuAvail:    resource.Cpu,
			MemUsed:     (resource.MemoryTotal - resource.Memory) >> 10,
			MemAvail:    resource.Memory >> 10,
			DiskUsed:    (resource.DiskTotal - resource.Disk) >> 30,
			DiskAvail:   resource.Disk >> 30,
			VolumeUsed:  160,
			VolumeAvail: 1100,
			PubipUsed:   int64(pubipUsed),
			PubipAvail:  int64(pubipTotal - pubipUsed),
			PrvipUsed:   int64(prvipUsed),
			PrvipAvail:  int64(prvipTotal - prvipUsed),
		}
	} else {
		quota := &model.Quota{}
		err := db.Where("owner = ?", memberShip.OrgID).Take(quota).Error
		if err != nil {
			log.Println("Failed to query quota")
			quota.Cpu = 6
			quota.Memory = 24
			quota.Disk = 200
			quota.Subnet = 10
			quota.PublicIp = 2
			quota.PrivateIp = 5
			quota.Gateway = 2
			quota.Volume = 100
			quota.Secgroup = 10
			quota.Secrule = 100
			quota.Instance = 4
			quota.Openshift = 1
			err = db.Create(quota).Error
			if err != nil {
				log.Println("Failed to query quota")
				c.Data["ErrorMsg"] = err.Error()
				c.HTML(http.StatusBadRequest, "error")
				return
			}
		}
		rcData, err = a.getOrgUsage(ctx, quota)
		if err != nil {
			log.Println("Failed to get usage")
			c.Data["ErrorMsg"] = err.Error()
			c.HTML(http.StatusBadRequest, "error")
			return
		}
	}
	c.JSON(200, rcData)
	return
}

func (a *Dashboard) getSystemIpUsage(ctx context.Context, ntype string) (ipTotal, ipUsed int, err error) {
	db := DB()
	subnets := []*model.Subnet{}
	err = db.Where("type = ?", ntype).Find(&subnets).Error
	if err != nil {
		log.Println("Failed to query subnets")
		return
	}
	where := "subnet_id in ("
	for i, sub := range subnets {
		if i == 0 {
			where = fmt.Sprintf("%s%d", where, sub.ID)
		} else {
			where = fmt.Sprintf("%s,%d", where, sub.ID)
		}
	}
	where = where + ")"
	err = db.Model(&model.Address{}).Where(where).Count(&ipTotal).Error
	if err != nil {
		log.Println("Failed to count total public ips")
		return
	}
	err = db.Model(&model.Address{}).Where(where).Where("allocated = ?", true).Count(&ipUsed).Error
	if err != nil {
		log.Println("Failed to count used public ips")
		return
	}
	return
}

func (a *Dashboard) getOrgIpUsage(ctx context.Context, ntype string) (ipUsed int, err error) {
	memberShip := GetMemberShip(ctx)
	db := DB()
	subnets := []*model.Subnet{}
	err = db.Where("type = ?", ntype).Find(&subnets).Error
	if err != nil {
		log.Println("Failed to query subnets")
		return
	}
	where := "subnet in ("
	for i, sub := range subnets {
		if i == 0 {
			where = fmt.Sprintf("%s%d", where, sub.ID)
		} else {
			where = fmt.Sprintf("%s,%d", where, sub.ID)
		}
	}
	where = where + ")"
	err = db.Model(&model.Interface{}).Where(where).Where("owner = ?", memberShip.OrgID).Count(&ipUsed).Error
	if err != nil {
		log.Println("Failed to count used ips")
		return
	}
	return
}

func (a *Dashboard) getOrgUsage(ctx context.Context, quota *model.Quota) (rcData *ResourceData, err error) {
	var cpu, memory, disk int32
	_, instances, err := instanceAdmin.List(ctx, 0, -1, "", "")
	for _, inst := range instances {
		flavor := inst.Flavor
		cpu += flavor.Cpu
		memory += flavor.Memory
		disk += flavor.Disk
	}
	pubip, err := a.getOrgIpUsage(ctx, "public")
	prvip, err := a.getOrgIpUsage(ctx, "private")
	rcData = &ResourceData{
		Title:       "Organization Quota Usage Ratio",
		CpuUsed:     int64(cpu),
		CpuAvail:    int64(quota.Cpu - cpu),
		MemUsed:     int64(memory),
		MemAvail:    int64(quota.Memory*1024 - memory),
		DiskUsed:    int64(disk),
		DiskAvail:   int64(quota.Disk - disk),
		VolumeUsed:  0,
		VolumeAvail: int64(quota.Volume),
		PubipUsed:   int64(pubip),
		PubipAvail:  int64(quota.PublicIp - int32(pubip)),
		PrvipUsed:   int64(prvip),
		PrvipAvail:  int64(quota.PrivateIp - int32(prvip)),
	}
	return
}
