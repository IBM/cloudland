/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package rpcs

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"web/src/model"

	"golang.org/x/net/context"
	"gopkg.in/macaron.v1"
)

var frontbackService = &FrontbackService{}

type FrontbackService struct{}

func (fb *FrontbackService) CallbackAgent(ctx context.Context, control, command string, duration time.Duration) (ids []int32, err error) {
	values := &model.Hyper{}
	values.LoadControl(control)
	agent := &model.Hyper{}
	values.Duration = int64(duration)
	values.Status = 1
	if err = agent.Updates(ctx, values); err != nil {
		log.Println("Update hyper value error: ", err)
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
			log.Println("Update agent error: ", err)
			return
		}
		ids = append(ids, v.Hostid)
	}
	return
}

func (fb *FrontbackService) doExecute(ctx context.Context, id, extra int32, command, control string) (reply *ExecuteReply, err error) {
	reply = &ExecuteReply{}
	firstToken := strings.Split(control, " ")[0]
	if id < 0 && firstToken == "callback=agent" {
		fb.CallbackAgent(ctx, control, command, 0)
		return
	}
	if firstToken == "callback=agent" {
		var ids []int32
		ids, err = fb.CallbackAgent(ctx, control, command, 0)
		if err != nil {
			log.Println("Call back agent error: ", err)
			return
		}
		ss := []string{}
		for _, id := range ids {
			if id < 0 {
				ss = append(ss, fmt.Sprintf("agent %d synced in %v", id, 0))
			} else {
				ss = append(ss, fmt.Sprintf("hyper %d synced in %v", id, 0))
			}
		}
		reply.Status = strings.Join(ss, "\n")
	} else if firstToken == "error=resource" {
		cmd, args := DecodeCommand(command)
		if cmd != "" {
			ctx2 := context.WithValue(ctx, "error", "resource")
			reply.Status, err = fb.dispatchExecute(ctx2, cmd, args)
		}
	} else {
		cmd, args := DecodeCommand(command)
		if cmd != "" {
			ctx2 := context.WithValue(ctx, "hostid", id)
			reply.Status, err = fb.dispatchExecute(ctx2, cmd, args)
		}
	}
	return
}

// Execute main entry
func (fb *FrontbackService) Execute(c *macaron.Context) {
	ctx := c.Req.Context()
	request, _ := c.Req.Body().Bytes()
	execReq := &ExecuteRequest{}
	err := json.Unmarshal(request, execReq)
	if err != nil {
		log.Println("Json unmarshal error:", err)
		c.JSON(404, &ExecuteReply{Status: "Bad request"})
	}
	id := execReq.Id
	extra := execReq.Extra
	command := execReq.Command
	control := execReq.Control
	reply, err := fb.doExecute(ctx, id, extra, command, control)
	c.JSON(200, reply)
	return
}

func (fb *FrontbackService) dispatchExecute(ctx context.Context, cmd string, args []string) (status string, err error) {
	if command := Get(cmd); command != nil {
		status, err = command(ctx, args)
	} else {
		err = fmt.Errorf("no command %s found", cmd)
		log.Println("Command dispatch error: ", err)
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

type Command func(ctx context.Context, args []string) (string, error)
