/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package grpcs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/IBM/cloudland/web/clui/jobs"
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/IBM/cloudland/web/sca/logs"
	"github.com/IBM/cloudland/web/clui/model"
)

var (
	JobService *jobService
)

type jobService struct {
	*logs.Tracing
	chans  map[int32]chan *jobs.Event
	locker sync.Mutex
}

func GetJobService() *jobService {
	if JobService == nil {
		JobService = &jobService{
			Tracing: logs.StartTracing("Job"),
			locker: sync.Mutex{},
			chans:  map[int32]chan *jobs.Event{},
		}
	}
	return JobService
}

func (js *jobService) Register(jobId int32, max int32) (err error) {
	js.locker.Lock()
	defer js.locker.Unlock()
	if _, ok := js.chans[jobId]; ok {
		err = fmt.Errorf("chan for job %d exists", jobId)
		return
	}
	ch := make(chan *jobs.Event, max)
	js.chans[jobId] = ch
	return
}

func (js *jobService) GetChan(jobId int32) (ch chan *jobs.Event, ok bool) {
	js.locker.Lock()
	ch, ok = js.chans[jobId]
	js.locker.Unlock()
	return
}

func (js *jobService) Wait(ctx context.Context, jobId int32, max int, timeout time.Duration, rs jobs.JobService_RunServer) (jobStatus string, err error) {
	sp, ctx := js.StartLogging(ctx, "JobWait")
	defer sp.Finish()
	ch, ok := js.GetChan(jobId)
	if !ok {
		err = fmt.Errorf("chan for job %d does not exist", jobId)
		return
	}
	succeed := true
	for i := 0; i < max; i++ {
		var event *jobs.Event
		event, err = js.WaitOnce(ch, timeout)
		if err != nil {
			sp.Error(err)
			return
		}
		sp.WithFields(map[string]interface{}{
			"event": event,
			"index": i,
			"max":   max,
		}).Info("received event")
		rs.Send(event)
		if !event.Succeed {
			succeed = false
		}
	}
	if succeed && max > 1 {
		db := dbs.DB()
		job := &model.Job{}
		job.ID = uint(jobId)
		db.Delete(job)
	}
	return
}

func (js *jobService) WaitOnce(ch chan *jobs.Event, timeout time.Duration) (
	event *jobs.Event, err error,
) {
	select {
	case event = <-ch:
	case <-time.After(timeout):
		err = fmt.Errorf("timeout after %v", timeout)
	}
	return
}

func (js *jobService) remove(jobId int32) (ch chan *jobs.Event, err error) {
	js.locker.Lock()
	defer js.locker.Unlock()
	ok := false
	if ch, ok = js.chans[jobId]; !ok {
		err = fmt.Errorf("chan for job %d does not exist", jobId)
	}
	delete(js.chans, jobId)
	return
}

func (js *jobService) GetEndpoint() (endpoint string) {
	const (
		API_LISTEN   = "grpc.listen"
		API_ENDPOINT = "grpc.endpoint"
	)
	if endpoint = viper.GetString(API_ENDPOINT); endpoint == "" {
		endpoint = viper.GetString(API_LISTEN)
	}
	return
}

func (js *jobService) FirstOrCreate(ctx context.Context, pj *jobs.Job) (job *model.Job, err error) {
	sp, ctx := js.StartLogging(ctx, "JobFirstOrCreate")
	defer sp.Finish()
	job = &model.Job{}
	job.LoadRequest(pj)
	db := dbs.DB()
	if err = db.FirstOrCreate(job, job.ID).Error; err != nil {
		sp.Error(err)
		return
	}
	if err = db.Model(job).Update(&model.Job{
		Hooks:  js.GetEndpoint(),
		Status: int32(jobs.Status_RUNNING),
	}).Error; err != nil {
		sp.Error(err)
		return
	}
	return
}

func (js *jobService) Run(req *jobs.RunRequest, rs jobs.JobService_RunServer) (err error) {
	ctx := rs.Context()
	sp, ctx:= js.StartLogging(ctx, "JobRun")
	sp.Info(sp.Context())
	defer sp.Finish()
	wait := req.GetWait()
	var job *model.Job
	defer sp.Finish()

	if job, err = js.FirstOrCreate(ctx, req.GetJob()); err != nil {
		sp.Error(err)
		return
	}
	jobId := int32(job.ID)
	echoNumber := job.EchoNumber
	if wait > 0 {
		if err = js.Register(jobId, echoNumber); err != nil {
			sp.Error(err)
			return
		}
		defer js.remove(jobId)
	}
	if err = js.Execute(ctx, job); err != nil {
		sp.Error(err.Error())
		return
	}
	if wait > 0 {
		js.Wait(ctx, jobId, int(echoNumber), time.Second*time.Duration(wait), rs)
	}
	return
}

func (js *jobService) RunOnce(ctx context.Context, req *jobs.Job) (
	event *jobs.Event, err error,
) {
	sp, ctx := js.StartLogging(ctx, "JobRunOnce")
	defer sp.Finish()

	var job *model.Job
	if job, err = js.FirstOrCreate(ctx, req); err != nil {
		sp.Error(err)
		return
	}
	err = js.Register(int32(job.ID), 1)
	if err != nil {
		sp.Error(err.Error())
		return
	}
	err = js.Execute(ctx, job)
	if err != nil {
		sp.Error(err.Error())
		return
	}
	ch, ok := js.GetChan(int32(job.ID))
	if !ok {
		err = fmt.Errorf("chan for job %v does not exist", job.ID)
		sp.Error(err.Error())
		return
	}
	timeout := 10 * time.Second
	event, err = js.WaitOnce(ch, timeout)
	if err != nil {
		sp.Error(err.Error())
		return
	}
	db := dbs.DB()
	jb := &model.Job{}
	jb.ID = job.ID
	db.Delete(jb)
	return
}

func (js *jobService) Execute(ctx context.Context, job *model.Job) (err error) {
	sp, ctx:= js.StartLogging(ctx, "JobExecute")
	defer sp.Finish()
	req := &scripts.ExecuteRequest{
		Id:      int32(job.ID),
		Control: job.Control,
		Command: job.Command,
		Extra:   job.Extra,
	}
	if _, err = RemoteExecClient().Execute(ctx, req); err != nil {
		sp.Error(err)
		db := dbs.DB()
		db.Model(job).Update(&model.Job{Status: int32(jobs.Status_FAILED)})
		return
	}
	return
}

func (js *jobService) notify(event *jobs.Event) (err error) {
	jobId := event.GetJobId()
	if ch, ok := js.GetChan(jobId); ok {
		ch <- event
	}
	return
}

func (js *jobService) Notify(ctx context.Context, req *jobs.NotifyRequest) (rep *jobs.NotifyReply, err error) {
	sp, ctx:= js.StartLogging(ctx, "JobNotify")
	defer sp.Finish()
	event := req.GetEvent()
	rep = &jobs.NotifyReply{}
	js.notify(event)
	return
}

func (js *jobService) Delete(ctx context.Context, req *jobs.DeleteRequest) (rep *jobs.DeleteReply, err error) {
	sp, ctx:= js.StartLogging(ctx, "JobDelete")
	defer sp.Finish()
	sp.Info(sp.Context())
	callback := req.GetCallback()
	query := req.GetJob()
	var value, where interface{}
	if callback {
		cb := &model.Callback{}
		cb.LoadRequest(query)
		value = cb
		where = cb
	} else {
		job := &model.Job{}
		job.LoadRequest(query)
		value = job
		where = job
	}
	rep = &jobs.DeleteReply{}
	db := dbs.DB().Delete(value, where)
	if err = db.Error; err != nil {
		sp.Error(err)
		return
	}
	rep.Deleted = db.RowsAffected
	return
}

func (js *jobService) Invoke(ctx context.Context, req *jobs.InvokeRequest) (rep *jobs.InvokeReply, err error) {
	sp, ctx:= js.StartLogging(ctx, "JobInvoke")
	defer sp.Finish()
	rep = &jobs.InvokeReply{}
	var job *model.Job
	job, err = js.FirstOrCreate(ctx, req.GetJob())
	if err != nil {
		sp.Error(err)
		return
	}
	rep.Job = job.ToReply()
	err = js.Execute(ctx, job)
	return
}

func (js *jobService) List(req *jobs.ListRequest, ls jobs.JobService_ListServer) (err error) {
	ctx := ls.Context()
	sp, ctx:= js.StartLogging(ctx, "JobList")
	defer sp.Finish()
	callback := req.GetCallback()
	unscoped := req.GetUnscoped()
	db := dbs.DB()
	if unscoped {
		db = db.Unscoped()
	}
	query := req.GetJob()
	out := []*model.Job{}
	if callback {
		cb := &model.Callback{}
		cb.LoadRequest(query)
		cbs := []*model.Callback{}
		if err = db.Find(&cbs, cb).Error; err != nil {
			sp.Error(err)
			return
		}
		max := len(cbs)
		for i := 0; i < max; i++ {
			out = append(out, cbs[i].ToJob())
		}
	} else {
		job := &model.Job{}
		job.LoadRequest(query)
		err = db.Find(&out, job).Error
	}
	if err != nil {
		sp.Error(err)
		return
	}
	for _, job := range out {
		if err = ls.Send(job.ToReply()); err != nil {
			sp.Error(err)
			return
		}
	}
	return
}

func (js *jobService) RemoteExec(ctx context.Context, req *jobs.RemoteExecRequest) (rep *jobs.RemoteExecReply, err error) {
	sp, ctx := js.StartLogging(ctx, "RemoteExec")
	defer sp.Finish()
	rep = &jobs.RemoteExecReply{}
	client := RemoteExecClient()
	execRep, err := client.Execute(ctx,
		&scripts.ExecuteRequest{
			Id:      req.GetId(),
			Control: req.GetControl(),
			Command: req.GetCommand(),
			Extra:   req.GetExtra(),
		})
	if err != nil {
		sp.Error(err)
		return
	}
	rep.Status = execRep.GetStatus()
	return
}
