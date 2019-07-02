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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/IBM/cloudland/web/sca/tables"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
)

func Commands(getContext func() context.Context, getClient func() DBAdminClient) (commands []*cobra.Command) {
	dbCmd := &cobra.Command{
		Use: "db",
	}
	statsCmd := &cobra.Command{
		Use:  "stats",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			req := &StatsRequest{}
			var rep *StatsReply
			if rep, err = getClient().Stats(getContext(), req); err != nil {
				return
			}
			stats := rep.GetStats()
			w := tables.NewTable()
			w.SetHeader([]string{"name", "value"})
			type getValue func() int64
			names := []string{
				"MaxOpenConnections",
				"OpenConnections",
				"InUse",
				"Idle",
				"WaitCount",
				"WaitDuration",
				"MaxIdleClosed",
			}
			values := []getValue{
				stats.GetMaxOpenConnections,
				stats.GetOpenConnections,
				stats.GetInUse,
				stats.GetIdle,
				stats.GetWaitCount,
				stats.GetWaitDuration,
				stats.GetMaxIdleClosed,
			}
			for i := 0; i < len(names); i++ {
				name := names[i]
				get := values[i]
				v := get()
				value := ""
				if name == "WaitDuration" {
					value = fmt.Sprint(time.Duration(v))
				} else {
					value = fmt.Sprint(v)
				}
				w.Append([]string{name, value})
			}
			w.Render()
			return
		},
	}

	tablesCmd := &cobra.Command{
		Use:  "tables",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			req := &TablesRequest{}
			var rep *TablesReply
			if rep, err = getClient().Tables(getContext(), req); err != nil {
				return
			}
			ts := rep.GetTables()
			if len(ts) > 0 {
				names := []string{}
				for name := range ts {
					names = append(names, name)
				}
				sort.Strings(names)
				w := tables.NewTable()
				for _, name := range names {
					tab := ts[name]
					w.Append([]string{name, fmt.Sprint(tab.GetRows()), fmt.Sprint(tab.GetDeleted()), tab.GetError()})
				}
				w.SetHeader([]string{"table", "rows", "deleted", "errors"})
				w.Render()
			}
			return
		},
	}
	execCmd := &cobra.Command{
		Use: "exec",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			sql := strings.Join(args, " ")
			req := &ExecRequest{
				Sql: sql,
			}
			rep, err := getClient().Exec(getContext(), req)
			if err != nil {
				return
			}
			w := tables.NewTable()
			affected := rep.GetAffected()
			w.Append([]string{"affected", fmt.Sprint(affected)})
			w.Render()
			return
		},
	}
	queryCmd := &cobra.Command{
		Use: "query",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			sql := strings.Join(args, " ")
			req := &QueryRequest{
				Sql: sql,
			}
			rep, err := getClient().Query(getContext(), req)
			if err != nil {
				return
			}
			w := tables.NewTable()
			result := rep.GetResult()
			if len(result) > 0 {
				names := result[0].GetValues()
				w.SetHeader(names)
				for i := 1; i < len(result); i++ {
					w.Append(result[i].GetValues())
				}
				w.Render()
			}
			return
		},
	}
	dbCmd.AddCommand(statsCmd, tablesCmd, execCmd, queryCmd)
	commands = append(commands, dbCmd)
	return
}
