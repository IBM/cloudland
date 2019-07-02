/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package selfs

import (
	"context"
	fmt "fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/IBM/cloudland/web/sca/tables"
	grpc "google.golang.org/grpc"
)

func Commands(getContext func() context.Context,
	getClientConn func() *grpc.ClientConn) (commands []*cobra.Command) {
	selfCmd := &cobra.Command{
		Use:  "self",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			client := NewSelfAdminClient(getClientConn())
			rep, err := client.Runtime(getContext(), &RuntimeRequest{})
			if err != nil {
				return
			}
			version, pid, executable, args, environ, netrc, pwd := rep.GetVersion(),
				rep.GetPid(), rep.GetExecutable(), rep.GetArgs(),
				rep.GetEnviron(), rep.GetNetrc(), rep.GetPwd()
			w := tables.NewTable()
			w.Append([]string{"version", version})
			w.Append([]string{"pid", pid})
			w.Append([]string{"pwd", pwd})
			w.Append([]string{"executable", executable})
			w.Append([]string{"args", strings.Join(args, " ")})
			w.Append([]string{"environs", strings.Join(environ, "\n")})
			w.Append([]string{"netrc", strings.Join(netrc, "\n")})
			w.SetHeader([]string{"item", "value"})
			w.Render()
			return
		},
	}

	upgradeCmd := &cobra.Command{
		Use:  "upgrade <version>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			version := args[0]
			req := &UpgradeRequest{Version: version}
			client := NewSelfAdminClient(getClientConn())
			_, err = client.Upgrade(getContext(), req)
			return
		},
	}
	setCmd := &cobra.Command{
		Use:  "set <key>[=<value>]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			key := args[0]
			if KeyClassify(key) == Unknown {
				err = fmt.Errorf("for env set, key should prefix with CLADMIN_, for netrc set, key should be a hostname")
				return
			}
			value := ""
			if idx := strings.Index(key, "="); idx != -1 {
				value = key[idx+1:]
				key = key[0:idx]
			}
			req := &SetRequest{
				Key:   key,
				Value: value,
			}
			client := NewSelfAdminClient(getClientConn())
			_, err = client.Set(getContext(), req)
			return
		},
	}
	selfCmd.AddCommand(upgradeCmd, setCmd)
	commands = append(commands, selfCmd)
	return
}
