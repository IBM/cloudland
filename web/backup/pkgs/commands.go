/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package pkgs

import (
	"errors"
	fmt "fmt"
	"io"
	"os"

	"github.com/IBM/cloudland/web/sca/tables"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

func DisplayPkg(pkg *Pkg) {
	w := tables.NewTable()
	w.Append([]string{pkg.GetName(), pkg.GetVersion(), pkg.GetSha1Sum()})
	w.SetHeader([]string{"name", "version", "sha1sum"})
	w.Render()

}

func Commands(getContext func() context.Context,
	getClientConn func() *grpc.ClientConn) (commands []*cobra.Command) {
	pkgCmd := &cobra.Command{
		Use: "pkg",
	}
	buildCmd := &cobra.Command{
		Use:     "build <project[=version]>",
		Example: "cladmin pkg build cladmin",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name, version := NameVersion(args[0])
			if version == "" {
				version = "latest"
			}
			ctx := getContext()
			client := NewPkgAdminClient(getClientConn())
			req := &BuildRequest{
				Project: &Project{
					Name:    name,
					Version: version,
				},
			}
			rep, err := client.Build(ctx, req)
			if err != nil {
				return
			}
			w := tables.NewTable()
			for _, pkg := range rep.GetPkgs() {
				w.Append([]string{pkg.GetName(), pkg.GetVersion(), pkg.GetSha1Sum()})
			}
			if w.NumLines() > 0 {
				w.SetHeader([]string{"name", "version", "sha1sum"})
				w.Render()
			}
			return
		},
	}
	downloadCmd := &cobra.Command{
		Use:     "download <pkg=version>",
		Example: "cladmin pkg download cladmin=v1.0.0",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			name, version := NameVersion(args[0])
			if version == "" || name == "" {
				err = errors.New("no pkg version or pkg name specified")
			}
			ctx := getContext()
			client := NewPkgAdminClient(getClientConn())
			req := &DownloadRequest{
				Pkg: &Pkg{
					Name:    name,
					Version: version,
				},
			}
			rep, err := client.Download(ctx, req)
			if err != nil {
				return
			}
			filename := fmt.Sprintf("%s_%s.tgz", name, version)
			file, err := os.Create(filename)
			if err != nil {
				return
			}
			defer file.Close()
			for {
				var chunk *PkgChunk
				chunk, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				data := chunk.GetChunk().GetData()
				_, err = file.Write(data)
				if err != nil {
					return
				}
			}
			return
		},
	}
	uploadCmd := &cobra.Command{
		Use:  "upload <pkg=version> <-|filename>",
		Args: cobra.ExactArgs(2),
		Example: `cladmin pkg upload cladmin=v1.0.0 cladmin.tgz
cat cladmin.tgz | cladmin pkg upload cladmin=v1.0.0 -`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name, version := NameVersion(args[0])
			filename := args[1]
			if name == "" || version == "" || filename == "" {
				err = errors.New("no pkg name or pkg version or pkg contents specified")
				return
			}
			options := &tables.Options{cmd}
			deploy := options.GetBool("deploy")
			var file *os.File
			if filename == "-" {
				file = os.Stdin
			} else {
				file, err = os.Open(filename)
				defer file.Close()
			}
			if err != nil {
				return
			}
			ctx := getContext()
			client := NewPkgAdminClient(getClientConn())
			us, err := client.Upload(ctx)
			if err != nil {
				return
			}
			var buf [1024]byte
			position := int64(0)
			var rep *UploadReply
			for {
				n := 0
				n, err = file.Read(buf[0:])
				if err == io.EOF {
					rep, err = us.CloseAndRecv()
					if err != nil {
						return
					}
					break
				} else if err != nil {
					return
				}
				chunk := &PkgChunk{
					Pkg: &Pkg{
						Name:    name,
						Version: version,
						Deploy:  deploy,
					},
					Chunk: &Chunk{
						Data:     buf[0:n],
						Position: position,
					},
				}
				position += int64(n)
				err = us.Send(chunk)
				if err != nil {
					return
				}
			}
			if rep != nil {
				DisplayPkg(rep.GetPkg())
			}
			return
		},
	}
	uploadCmd.Flags().Bool("deploy", false, "deploy pkg")
	removeCmd := &cobra.Command{
		Use:     "remove <pkg=version>",
		Args:    cobra.ExactArgs(1),
		Example: "cladmin pkg remove <cladmin=v1.0.0>",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name, version := NameVersion(args[0])
			if name == "" || version == "" {
				err = fmt.Errorf("No pkg name or version specified")
				return
			}
			req := &RemoveRequest{Pkg: &Pkg{
				Name:    name,
				Version: version,
			}}
			ctx := getContext()
			client := NewPkgAdminClient(getClientConn())
			rep, err := client.Remove(ctx, req)
			if err != nil {
				return
			}
			DisplayPkg(rep.GetPkg())
			return
		},
	}
	listCmd := &cobra.Command{
		Use:     "list",
		Args:    cobra.MaximumNArgs(1),
		Example: "cladmin pkg list [pkg]",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			ctx := getContext()
			client := NewPkgAdminClient(getClientConn())
			rep, err := client.List(ctx, &ListRequest{Name: name})
			if err != nil {
				return
			}
			w := tables.NewTable()
			for {
				var pkg *Pkg
				pkg, err = rep.Recv()
				if err == io.EOF {
					err = nil
					break
				} else if err != nil {
					return
				}
				pkgName := pkg.GetName()
				if pkg.GetDeploy() {
					pkgName += " ✔"
				}
				w.Append([]string{pkgName, pkg.GetVersion(), pkg.GetSha1Sum()})
			}
			if w.NumLines() > 0 {
				w.SetHeader([]string{"name", "version", "sha1sum"})
				w.Render()
			}
			return
		},
	}
	pkgCmd.AddCommand(buildCmd, downloadCmd, uploadCmd, removeCmd, listCmd)
	commands = append(commands, pkgCmd)
	return
}
