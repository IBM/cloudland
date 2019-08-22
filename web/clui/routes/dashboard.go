/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
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
	CpuUsed     int64 `json:"cpu_used"`
	CpuAvail    int64 `json:"cpu_avail"`
	MemUsed     int64 `json:"mem_used"`
	MemAvail    int64 `json:"mem_avail"`
	DiskUsed    int64 `json:"disk_used"`
	DiskAvail   int64 `json:"disk_avail"`
	VolumeUsed  int64 `json:"volume_used"`
	VolumeAvail int64 `json:"volume_avail"`
	PubipUsed   int64 `json:"pubip_used"`
	PubipAvail  int64 `json:"pubip_avail"`
	PrvipUsed   int64 `json:"prvip_used"`
	PrvipAvail  int64 `json:"prvip_avail"`
}

type Dashboard struct{}

func (a *Dashboard) Show(c *macaron.Context, store session.Store) {
	c.HTML(200, "dashboard")
	return
}

func (a *Dashboard) GetData(c *macaron.Context, store session.Store) {
	// memberShip := GetMemberShip(c.Req.Context())
	rcData := &ResourceData{}
	db := DB()
	//	if memberShip.OrgName == "admin" {
	resource := &model.Resource{}
	err := db.Take(resource).Error
	if err != nil {
		log.Println("Failed to query system resource")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	rcData.CpuUsed = resource.CpuTotal - resource.Cpu
	rcData.CpuAvail = resource.Cpu
	rcData.MemUsed = (resource.MemoryTotal - resource.Memory) >> 20
	rcData.MemAvail = resource.Memory >> 20
	rcData.DiskUsed = (resource.DiskTotal - resource.Disk) >> 30
	rcData.DiskAvail = resource.Disk >> 30
	rcData.VolumeUsed = 160
	rcData.VolumeAvail = 1100
	pubnets := []*model.Subnet{}
	err = db.Where("type = 'public'").Find(&pubnets).Error
	if err != nil {
		log.Println("Failed to query public subnets")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	pubipTotal := 0
	where := "subnet_id in ("
	for i, pub := range pubnets {
		if i == 0 {
			where = fmt.Sprintf("%s%d", where, pub.ID)
		} else {
			where = fmt.Sprintf("%s,%d", where, pub.ID)
		}
	}
	where = where + ")"
	err = db.Model(&model.Address{}).Where(where).Count(&pubipTotal).Error
	if err != nil {
		log.Println("Failed to count total public ips")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	pubipUsed := 0
	err = db.Model(&model.Address{}).Where(where).Where("allocated = ?", true).Count(&pubipUsed).Error
	if err != nil {
		log.Println("Failed to count used public ips")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	rcData.PubipUsed = int64(pubipUsed)
	rcData.PubipAvail = int64(pubipTotal - pubipUsed)

	prvnets := []*model.Subnet{}
	err = db.Where("type = 'private'").Find(&prvnets).Error
	if err != nil {
		log.Println("Failed to query private subnets")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	prvipTotal := 0
	where = "subnet_id in ("
	for i, prv := range prvnets {
		if i == 0 {
			where = fmt.Sprintf("%s%d", where, prv.ID)
		} else {
			where = fmt.Sprintf("%s,%d", where, prv.ID)
		}
	}
	where = where + ")"
	err = db.Model(&model.Address{}).Where(where).Count(&prvipTotal).Error
	if err != nil {
		log.Println("Failed to count total public ips")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	prvipUsed := 0
	err = db.Model(&model.Address{}).Where(where).Where("allocated = ?", true).Count(&prvipUsed).Error
	if err != nil {
		log.Println("Failed to count used public ips")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	rcData.PrvipUsed = int64(prvipUsed)
	rcData.PrvipAvail = int64(prvipTotal - prvipUsed)
	//	}
	c.JSON(200, rcData)
	return
}
