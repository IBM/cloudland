/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package releases

import (
	fmt "fmt"
	"io"

	"github.com/spf13/cobra"
	pkgs "github.com/IBM/cloudland/web/sca/pkgs"
	"github.com/IBM/cloudland/web/sca/tables"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

func Commands(getContext func() context.Context,
	getClientConn func() *grpc.ClientConn) (commands []*cobra.Command) {
	releaseCmd := &cobra.Command{
		Use: "release",
	}
	createCmd := &cobra.Command{
		Use:  "create <release>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			client := NewReleaseAdminClient(getClientConn())
			rep, err := client.Create(getContext(), &CreateRequest{
				Name: name,
			})
			if err != nil {
				return
			}
			name = rep.GetName()
			w := tables.NewTable()
			w.Append([]string{name, rep.GetVersion(), pkgs.Names(rep.GetPkgs())})
			w.SetHeader([]string{"name", "version", "pkgs"})
			w.Render()
			return
		},
	}
	deleteCmd := &cobra.Command{
		Use:  "delete <release> <version>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			version := args[1]
			client := NewReleaseAdminClient(getClientConn())
			rep, err := client.Delete(getContext(), &DeleteRequest{
				Name:    name,
				Version: version,
			})
			if err != nil {
				return
			}
			r := rep
			w := tables.NewTable()
			w.Append([]string{r.GetName(), r.GetVersion(), pkgs.Names(r.GetPkgs())})
			w.SetHeader([]string{"name", "version", "pkgs"})
			w.Render()
			return
		},
	}
	showCmd := &cobra.Command{
		Use:  "show <release> <version>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			version := args[1]
			client := NewReleaseAdminClient(getClientConn())
			r, err := client.Get(getContext(), &GetRequest{
				Name:    name,
				Version: version,
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			w.Append([]string{r.GetName(), r.GetVersion(), pkgs.Names(r.GetPkgs())})
			w.SetHeader([]string{"name", "version", "pkgs"})
			w.Render()
			return
		},
	}
	refreshCmd := &cobra.Command{
		Use:  "refresh <release>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			client := NewReleaseAdminClient(getClientConn())
			r, err := client.Refresh(getContext(), &RefreshRequest{
				Name: name,
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			w.Append([]string{r.GetName(), r.GetVersion(), pkgs.Names(r.GetPkgs())})
			w.SetHeader([]string{"name", "version", "pkgs"})
			w.Render()
			return
		},
	}
	listCmd := &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			options := &tables.Options{cmd}
			name := options.GetString("name")
			client := NewReleaseAdminClient(getClientConn())
			rep, err := client.List(getContext(), &ListRequest{
				Name: name,
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			for {
				var r *Release
				r, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				w.Append([]string{r.GetName(), r.GetVersion(), pkgs.Names(r.GetPkgs())})
			}
			if w.NumLines() > 0 {
				w.SetHeader([]string{"name", "version", "pkgs"})
				w.Render()
			}
			return
		},
	}
	listCmd.Flags().String("name", "", "release name")
	addCmd := &cobra.Command{
		Use:  "add <release> <pkg=version>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			pkg, version := pkgs.NameVersion(args[1])
			if name == "" || pkg == "" || version == "" {
				err = fmt.Errorf(
					"no release name, pkg name or pkg version speficified")
				return
			}
			client := NewReleaseAdminClient(getClientConn())
			rep, err := client.Add(getContext(), &AddRequest{
				Name: name,
				Pkg: &pkgs.Pkg{
					Name:    pkg,
					Version: version,
				},
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			w.Append([]string{rep.GetName(), pkgs.Names([]*pkgs.Pkg{rep.GetPkg()})})
			w.SetHeader([]string{"name", "pkgs"})
			w.Render()
			return
		},
	}
	removeCmd := &cobra.Command{
		Use:  "remove <release> <pkg>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			pkg, _ := pkgs.NameVersion(args[1])
			client := NewReleaseAdminClient(getClientConn())
			rep, err := client.Remove(getContext(), &RemoveRequest{
				Name: name,
				Pkg:  &pkgs.Pkg{Name: pkg},
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			w.Append([]string{rep.GetName(), pkgs.Names([]*pkgs.Pkg{rep.GetPkg()})})
			w.SetHeader([]string{"name", "pkgs"})
			w.Render()
			return
		},
	}
	publishCmd := &cobra.Command{
		Use:  "publish <release>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			client := NewReleaseAdminClient(getClientConn())
			rep, err := client.Publish(getContext(), &PublishRequest{
				Name: name,
			})
			if err != nil {
				return
			}
			w := tables.NewTable()
			w.Append([]string{rep.GetName(), rep.GetVersion(), pkgs.Names(rep.GetPkgs())})
			w.SetHeader([]string{"name", "version", "pkgs"})
			w.Render()
			return
		},
	}
	releaseCmd.AddCommand(createCmd, deleteCmd, listCmd, showCmd, addCmd, removeCmd, publishCmd, refreshCmd)
	commands = append(commands, releaseCmd)
	return
}
