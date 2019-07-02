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

	"github.com/IBM/cloudland/web/clui/hypers"
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/IBM/cloudland/web/sca/logs"
)

func GetHyperService() hypers.HyperServiceServer {
	return &hypervisor{
		logs.StartTracing("HyperAdmin"),
	}
}

type hypervisor struct {
	*logs.Tracing
}

func (s *hypervisor) Get(ctx context.Context,
	req *hypers.GetRequest) (rep *hypers.GetReply, err error) {
	sp, ctx := s.StartLogging(ctx, "HyperGet")
	defer sp.Finish()
	rep = &hypers.GetReply{}
	hostid := req.GetId()
	db := dbs.DB()
	hyper := &model.Hyper{
		Hostid: hostid,
	}
	rep = &hypers.GetReply{}
	if err = db.First(hyper, map[string]interface{}{
		"hostid": hyper.Hostid,
	}).Error; err != nil {
		sp.Error(err)
		return
	}
	rep.Hyper = hyper.ToReply()
	return
}

func (s *hypervisor) List(
	req *hypers.ListRequest, rep hypers.HyperService_ListServer) (err error) {
	ctx := rep.Context()
	sp, ctx := s.StartLogging(ctx, "HyperList")
	defer sp.Finish()
	parentid := req.GetParentid()
	status := req.GetStatus()
	where := &model.Hyper{
		Parentid: parentid,
		Status:   model.HyperStatusNames[status],
	}
	db := dbs.DB()
	if req.GetUnscoped() {
		db = db.Unscoped()
	}
	db = db.Model(&model.Hyper{})
	kind := req.GetKind()
	switch kind {
	case hypers.Kind_hyper:
		db = db.Where(where).Where("hostid >= 0")
	case hypers.Kind_agent:
		db = db.Where(where).Where("hostid < 0")
	}
	rows, err := db.Rows()
	if err = db.Error; err != nil {
		sp.WithError(err).Debug()
		return
	}
	defer rows.Close()
	for rows.Next() {
		hyper := &model.Hyper{}
		err = db.ScanRows(rows, hyper)
		if err != nil {
			sp.Error(err)
			return
		}
		if err = rep.Send(hyper.ToReply()); err != nil {
			return
		}
	}
	return
}

func (s *hypervisor) Delete(ctx context.Context,
	req *hypers.DeleteRequest) (rep *hypers.DeleteReply, err error) {
	sp, ctx := s.StartLogging(ctx, "HyperDelete")
	defer sp.Finish()
	rep = &hypers.DeleteReply{}
	hyper := &model.Hyper{}
	db := dbs.DB().Delete(hyper, hyper)
	if err = db.Error; err == nil {
		rep.Deleted = db.RowsAffected
	} else {
		sp.Error(err)
	}
	return
}
