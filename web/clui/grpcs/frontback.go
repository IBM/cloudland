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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/IBM/cloudland/web/clui/jobs"
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/logs"
	"golang.org/x/net/context"
)

var _frontbackService *frontbackService

func GetRemoteExecService() scripts.RemoteExecServer {
	if _frontbackService == nil {
		_frontbackService = &frontbackService{
			logs.StartTracing("Frontback"),
		}
	}
	return _frontbackService
}

type frontbackService struct {
	*logs.Tracing
}

func (fb *frontbackService) CallbackAgent(ctx context.Context, control, command string, duration time.Duration) (ids []int32, err error) {
	sp, ctx := fb.StartLogging(ctx, "CallbackAgent")
	defer sp.Finish()
	values := &model.Hyper{}
	values.LoadControl(control)
	agent := &model.Hyper{}
	values.Duration = int64(duration)
	values.Status = 1
	if err = agent.Updates(ctx, values); err != nil {
		sp.Error(err)
		return
	}
	ids = append(ids, agent.Hostid)
	items := strings.Split(command, "\n")
	for _, item := range items {
		if item == "" {
			continue
		}
		a := &model.Hyper{}
		v := &model.Hyper{}
		v.LoadCommand(item)
		v.Parentid = agent.Hostid
		v.Duration = int64(duration)
		if err = a.Updates(ctx, v); err != nil {
			sp.Error(err)
			return
		}
		ids = append(ids, v.Hostid)
	}
	return
}

func (fb *frontbackService) doExecute(ctx context.Context, id, extra int32, command, control string) (reply *scripts.ExecuteReply, err error) {
	sp, ctx := fb.StartLogging(ctx, "doExecute")
	defer sp.Finish()
	reply = &scripts.ExecuteReply{}
	firstToken := strings.Split(control, " ")[0]
	if id < 0 && firstToken == "callback=agent" {
		fb.CallbackAgent(ctx, control, command, 0)
		return
	}
	sp.Debug("id: ", id, ", command: ", command)
	sp.WithFields(map[string]interface{}{
		"id":      id,
		"command": command,
	}).Debug("Event received")
	job := &model.Job{}
	/*
		db := dbs.DB()
		if err = db.First(job, uint(id)).Error; err != nil {
			sp.WithError(err).Debug()
			cmd, args := DecodeCommand(command)
			sp.Debug("cmd:", cmd, ", args: ", args)
			if cmd != "" {
				reply.Status, err = fb.dispatchExecute(ctx, job, cmd, args)
			}
			return
		}
		succeed := true
		if strings.Index(control, "error") != -1 {
			succeed = false
		}
	*/
	sp.Debug(firstToken, control)
	if firstToken == "callback=agent" {
		duration := time.Now().Sub(job.CreatedAt)
		var ids []int32
		ids, err = fb.CallbackAgent(ctx, control, command, duration)
		if err != nil {
			sp.Error(err)
			return
		}
		ss := []string{}
		for _, id := range ids {
			if id < 0 {
				ss = append(ss, fmt.Sprintf("agent %d synced in %v", id, duration))
			} else {
				ss = append(ss, fmt.Sprintf("hyper %d synced in %v", id, duration))
			}
		}
		reply.Status = strings.Join(ss, "\n")
	} else if firstToken == "error=resource" {
		cmd, args := DecodeCommand(command)
		sp.Debug("cmd:", cmd, ", args: ", args)
		if cmd != "" {
			ctx2 := context.WithValue(ctx, "error", "resource")
			reply.Status, err = fb.dispatchExecute(ctx2, job, cmd, args)
		}
	} else {
		cmd, args := DecodeCommand(command)
		sp.Debug("cmd:", cmd, ", args: ", args)
		if cmd != "" {
			ctx2 := context.WithValue(ctx, "hostid", id)
			reply.Status, err = fb.dispatchExecute(ctx2, job, cmd, args)
		}
	}
	/*
		if err != nil {
			succeed = false
		}
		status := jobs.Status_FAILED
		if succeed {
			status = jobs.Status_DONE
		}
		if job.Status == int32(jobs.Status_RUNNING) {
			db.Model(job).Updates(&model.Job{
				Status: int32(status),
			})
		}
		if succeed && err == nil {
			if job.EchoNumber == 1 {
				err = db.Delete(job).Error
				if err != nil {
					sp.Error(err)
					return
				}
			}
		}
		event := &jobs.Event{
			Echo:    reply.Status,
			JobId:   int32(job.ID),
			Succeed: succeed,
		}
		if err = fb.Notify(ctx, job.Hooks, event); err != nil {
			sp.Error(err)
			return
		}
	*/
	return
}

// Execute main entry
func (fb *frontbackService) Execute(ctx context.Context,
	req *scripts.ExecuteRequest) (reply *scripts.ExecuteReply, err error) {
	sp, ctx := fb.StartLogging(ctx, "Execute")
	defer sp.Finish()
	reply = &scripts.ExecuteReply{}
	id := req.GetId()
	extra := req.GetExtra()
	command := req.GetCommand()
	control := req.GetControl()
	sp.WithFields(map[string]interface{}{
		"id":      id,
		"extra":   extra,
		"command": command,
		"control": control,
	}).Info("callback received")
	reply, err = fb.doExecute(ctx, id, extra, command, control)
	return
}

func (fb *frontbackService) Notify(ctx context.Context, targets string, event *jobs.Event) (err error) {
	sp, ctx := fb.StartLogging(ctx, "Notify")
	defer sp.Finish()
	hooks := strings.Split(targets, ",")
	for _, hook := range hooks {
		client := JobServiceClient(hook)
		if client == nil {
			sp.Debug("hook client is nil")
			continue
		}
		req := &jobs.NotifyRequest{
			Event: event,
		}
		if _, err = client.Notify(ctx, req); err != nil {
			sp.Error(err)
			return
		}
	}
	return
}

func (fb *frontbackService) dispatchExecute(ctx context.Context, job *model.Job, cmd string, args []string) (status string, err error) {
	sp, ctx := fb.StartLogging(ctx, "dispatchExecute")
	defer sp.Finish()
	if command := Get(cmd); command != nil {
		status, err = command(ctx, job, args)
	} else {
		err = fmt.Errorf("no command %s found", cmd)
		sp.WithError(err).Debug()
	}
	return

}

var (
	frontbacks = map[string]Command{}
	locker     = sync.Mutex{}
)

func Add(name string, command Command) {
	locker.Lock()
	frontbacks[name] = command
	locker.Unlock()
}

func Get(name string) (command Command) {
	ok := false
	locker.Lock()
	if command, ok = frontbacks[name]; !ok {
		command = nil
	}
	locker.Unlock()
	return
}

type Command func(ctx context.Context, job *model.Job, args []string) (string, error)
