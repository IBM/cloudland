/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package model

import (
	context "context"
	"strings"

	"github.com/IBM/cloudland/web/clui/jobs"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/jinzhu/gorm"
)

type Job struct {
	gorm.Model
	Status     int32
	Control    string `gorm:"type:varchar(256)"`
	Command    string `gorm:"type:text"`
	Extra      int32
	Hooks      string `gorm:"type:varchar(256)"`
	EchoNumber int32
}

type Callback struct {
	gorm.Model
	Status  int32
	Control string `gorm:"type:varchar(256)"`
	Command string `gorm:"type:varchar(65536)"`
	Extra   int32
	Job     Job `gorm:"ForeignKey:JobID"`
	JobID   uint
}

func init() {
	dbs.AutoMigrate(&Job{}, &Callback{})
	gradeName := "0001-Job-0001-Modify-Command"
	dbs.AutoUpgrade(gradeName, func(db *gorm.DB) (err error) {
		logger, _ := startLogging(context.Background(), gradeName)
		dbType := db.Dialect().GetName()
		if dbType == "postgres" {
			if err = db.Model(&Job{}).ModifyColumn("command", "text").Error; err != nil {
				logger.WithError(err).Debug("Error found when upgrading", gradeName)
			}
		}
		return
	})
}

func (job *Job) LoadRequest(p *jobs.Job) {
	job.ID = uint(p.GetId())
	job.Status = jobs.Status_value[strings.ToUpper(p.GetStatus())]
	job.Control = p.GetControl()
	job.Command = p.GetCommand()
	job.Extra = p.GetExtra()
	job.Hooks = p.GetHooks()
}

func (job *Job) ToReply() (p *jobs.Job) {
	p = &jobs.Job{
		Status:  strings.ToLower(jobs.Status(job.Status).String()),
		Control: job.Control,
		Command: job.Command,
		Extra:   job.Extra,
		Hooks:   job.Hooks,
		Echos:   job.EchoNumber,
	}
	p.Id = int32(job.ID)
	return
}

func (job *Job) BeforeCreate() (err error) {
	logger, _ := startLogging(context.Background(), "JobBeforeCreate")
	if job.Status == 0 {
		job.Status = int32(jobs.Status_CREATED)
	}
	if job.EchoNumber != 0 {
		return
	}
	job.EchoNumber = 1
	if strings.Index(job.Control, "toall=") != -1 {
		db := dbs.DB().Model(&Hyper{})
		if strings.Index(job.Control, "toall=agent") != -1 {
			db = db.Where("status = 1 and hostid < 0")
		} else {
			db = db.Where("status = 1 and hostid >= 0")
		}
		count := int32(0)
		if err = db.Count(&count).Error; err != nil {
			logger.Error(err)
			return
		}
		job.EchoNumber = count
	}
	return
}

func (cb *Callback) LoadRequest(p *jobs.Job) (err error) {
	cb.Command = p.GetCommand()
	cb.Control = p.GetControl()
	cb.JobID = uint(p.GetId())
	cb.Extra = p.GetExtra()
	cb.Status = jobs.Status_value[strings.ToUpper(p.GetStatus())]
	return
}

func (cb *Callback) ToJob() (job *Job) {
	job = &Job{}
	job.ID = cb.JobID
	job.Command = cb.Command
	job.Control = cb.Control
	job.Extra = cb.Extra
	job.Status = cb.Status
	return
}

func (cb *Callback) BeforeCreate() (err error) {
	if cb.Status == 0 {
		cb.Status = int32(jobs.Status_CREATED)
	}
	return
}
