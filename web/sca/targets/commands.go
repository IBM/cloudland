/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package targets

import (
	fmt "fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/IBM/cloudland/web/sca/pkgs"
	"github.com/IBM/cloudland/web/sca/releases"
	"github.com/IBM/cloudland/web/sca/tables"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

func DisplayTarget(target *Target) {
	w := NewTargetTable()
	AppendTarget(w, target)
	w.Render()
}

func NewTargetTable() tables.Table {
	w := tables.NewTable()
	keys := []string{}
	keys = append(keys, "name", "release", "version", "state")
	w.SetHeader(keys)
	return w
}

func AppendTarget(w tables.Table, target *Target) {
	row := []string{}
	row = append(row,
		target.GetName(),
		target.GetRelease().GetName(),
		target.GetRelease().GetVersion(),
		target.GetState().String())
	w.Append(row)

}

func Commands(getContext func() context.Context,
	getClientConn func() *grpc.ClientConn) (
	commands []*cobra.Command) {
	targetCmd := &cobra.Command{
		Use: "target",
	}

	createCmd := &cobra.Command{
		Use:  "create <target> <release[=version]>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			release := args[1]
			release, version := pkgs.NameVersion(release)
			if version == "" {
				version = "latest"
			}
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Create(getContext(), &CreateRequest{
				Name: name,
				Release: &releases.Release{
					Name:    release,
					Version: version,
				},
			})
			if err != nil {
				return
			}

			target := rep
			DisplayTarget(target)
			return
		},
	}
	updateCmd := &cobra.Command{
		Use:  "update <target> <version>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			version := args[1]
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Update(getContext(), &UpdateRequest{
				Name:    name,
				Version: version,
			})
			if err != nil {
				return
			}
			target := rep
			DisplayTarget(target)
			return
		},
	}
	deleteCmd := &cobra.Command{
		Use:  "delete <target>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Delete(getContext(), &DeleteRequest{Name: name})
			if err != nil {
				return
			}
			target := rep
			DisplayTarget(target)
			return
		},
	}

	showCmd := &cobra.Command{
		Use:  "show <target>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Get(getContext(), &GetRequest{Name: name})
			if err != nil {
				return
			}
			DisplayTarget(rep)
			return
		},
	}

	listCmd := &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.List(getContext(), &ListRequest{})
			if err != nil {
				return
			}
			w := NewTargetTable()
			for {
				var target *Target
				target, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				AppendTarget(w, target)
			}
			if w.NumLines() > 0 {
				w.Render()
			}
			return
		},
	}

	deployCmd := &cobra.Command{
		Use:  "deploy <target>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Deploy(getContext(), &DeployRequest{
				Name: name,
			})
			if err != nil {
				return
			}
			DisplayTarget(rep)
			return
		},
	}

	envsCmd := &cobra.Command{
		Use:  "envs <target> [envname] [env value]",
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			key, value := "", ""
			switch len(args) {
			case 2:
				key = args[1]
			case 3:
				key = args[1]
				value = args[2]
			}
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Envs(getContext(), &EnvsRequest{
				Name: name,
				Env: &Env{
					Name:  key,
					Value: value,
				},
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			for {
				var env *Env
				env, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				w.Append([]string{env.GetName(), env.GetValue()})
			}
			if w.NumLines() > 0 {
				w.SetHeader([]string{"name", "value"})
				w.Render()
			}
			return
		},
	}
	keysCmd := &cobra.Command{
		Use:  "keys <target> [keyname] [content]",
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			keyname, private := "", ""
			switch len(args) {
			case 2:
				keyname = args[1]
			case 3:
				keyname, private = args[1], args[2]
			}
			if strings.HasPrefix(private, "@") {
				private = private[1:]
				if private != "" && private[0] == '~' {
					private = fmt.Sprintf("%s/%s", os.ExpandEnv("$HOME"), private[1:])
				}
				var b []byte
				b, err = ioutil.ReadFile(private)
				if err != nil {
					return
				}
				private = string(b)
			}
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Keys(getContext(), &KeysRequest{
				Name: name,
				Key: &Key{
					Name:    keyname,
					Private: private,
				},
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			for {
				var key *Key
				key, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				w.Append([]string{key.GetName(), key.GetPrivate()})
			}
			if w.NumLines() > 0 {
				w.SetHeader([]string{"name", "sha1sum"})
				w.Render()
			}
			return
		},
	}
	hostsCmd := &cobra.Command{
		Use:  "hosts <target> [hostname] [ip=group]",
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			hostname, ip, group := "", "", ""
			switch len(args) {
			case 2:
				hostname = args[1]
			case 3:
				hostname = args[1]
				ip = args[2]
			}
			if ip != "" {
				items := strings.Split(ip, "=")
				if len(items) == 2 {
					ip = items[0]
					group = items[1]
				}
			}
			client := NewTargetAdminClient(getClientConn())
			rep, err := client.Hosts(getContext(), &HostsRequest{
				Name: name,
				Host: &Host{
					Name:  hostname,
					Ip:    ip,
					Group: group,
				},
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			for {
				var host *Host
				host, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				w.Append([]string{host.GetName(), host.GetIp(),
					host.GetGroup()})
			}
			if w.NumLines() > 0 {
				w.SetHeader([]string{"name", "ip", "group"})
				w.Render()
			}
			return
		},
	}
	targetCmd.AddCommand(createCmd, updateCmd, deleteCmd, showCmd,
		listCmd, deployCmd, envsCmd, hostsCmd, keysCmd)
	commands = append(commands, targetCmd)
	return
}
