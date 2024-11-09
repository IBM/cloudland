/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import (
	"strings"

	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Register register db admin service in grpc server
func Register(server *grpc.Server) {
	RegisterDBAdminServer(server, &dbAdmin{})
}

type dbAdmin struct {
}

func (dba *dbAdmin) Stats(ctx context.Context,
	req *StatsRequest) (rep *StatsReply, err error) {
	logger, ctx := startLogging(ctx, "Stats")
	defer logger.Finish()
	rep = &StatsReply{}
	var stats *Stats
	stats, err = newStats()
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	rep.Stats = stats
	return
}

func (dba *dbAdmin) Tables(ctx context.Context,
	req *TablesRequest) (rep *TablesReply, err error) {
	logger, ctx := startLogging(ctx, "Tables")
	defer logger.Finish()
	rep = &TablesReply{}
	tables, err := newTables()
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	rep.Tables = tables
	return
}

func (dba *dbAdmin) Exec(ctx context.Context,
	req *ExecRequest) (rep *ExecReply, err error) {
	logger, ctx := startLogging(ctx, "Exec")
	defer logger.Finish()

	sql := strings.TrimSpace(req.GetSql())
	rep = &ExecReply{}
	if sql == "" {
		return
	}

	if rep.Affected, err = execSql(sql); err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.InvalidArgument, "%v", err)
	}
	return
}

func execSql(sql string) (affected int64, err error) {
	db := DB().Exec(sql)
	if err = db.Error; err == nil {
		affected = db.RowsAffected
	}
	return
}

func (dba *dbAdmin) Query(ctx context.Context,
	req *QueryRequest) (rep *QueryReply, err error) {
	logger, ctx := startLogging(ctx, "Query")
	defer logger.Finish()
	sql := strings.TrimSpace(req.GetSql())
	rep = &QueryReply{}
	if sql == "" {
		return
	}
	rs, err := query(sql)
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.InvalidArgument, "%v", err)
		return
	}
	for i := range rs {
		r := rs[i]
		rep.Result = append(rep.Result, &Result{
			Values: r,
		})
	}
	return
}

func query(sql string) (rs [][]string, err error) {
	db := DB().DB()
	rows, err := db.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	names, err := rows.Columns()
	if err != nil {
		return
	}

	n := len(names)
	if n == 0 {
		return
	}
	rs = append(rs, names)
	for rows.Next() {
		values := []interface{}{}
		for i := 0; i < n; i++ {
			values = append(values, new([]byte))
		}
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		result := make([]string, n)
		for i := 0; i < n; i++ {
			v := values[i]
			result[i] = string(*(v.(*[]byte)))
		}
		rs = append(rs, result)
	}
	return
}

func newStats() (stats *Stats, err error) {
	stats = &Stats{}
	db := DB().Model("")
	err = db.Error
	if err != nil {
		return
	}
	dbStats := db.DB().Stats()
	stats.MaxOpenConnections = int64(dbStats.MaxOpenConnections)
	stats.OpenConnections = int64(dbStats.OpenConnections)
	stats.InUse = int64(dbStats.InUse)
	stats.Idle = int64(dbStats.Idle)
	stats.WaitCount = dbStats.WaitCount
	stats.WaitDuration = int64(dbStats.WaitDuration)
	stats.MaxIdleClosed = dbStats.MaxIdleClosed
	stats.MaxLifetimeClosed = dbStats.MaxLifetimeClosed
	return
}

func newTables() (tables map[string]*Table, err error) {
	db := DB()
	tables = map[string]*Table{}
	names := TableNames()
	for i := 0; i < len(names); i++ {
		name := names[i]
		table := &Table{}
		tables[name] = table
		if emsg, ok := migrationErrors[name]; ok {
			table.Error = emsg
			continue
		}
		obj := objects[i]
		scope := db.NewScope(obj)
		udb := DB().Unscoped()
		var rows, deleted int64
		if err = udb.Table(name).Count(&rows).Error; err != nil {
			table.Error = err.Error()
			continue
		}
		table.Rows = rows
		if scope.HasColumn("deleted_at") {
			if err = udb.Table(name).Where(
				"deleted_at IS NOT NULL").Count(&deleted).Error; err != nil {
				table.Error = err.Error()
				continue
			}
		}
		table.Deleted = deleted
	}
	return
}
