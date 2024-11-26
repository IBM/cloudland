/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package main

import (
	"context"
	"net"

	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/IBM/cloudland/web/sca/servers"
	"golang.org/x/sync/errgroup"
)

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			address := viper.GetString("admin.listen")
			listen, err := net.Listen("tcp", address)
			if err != nil {
				return
			}
			m := cmux.New(listen)
			grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
			httpL := m.Match(cmux.HTTP1Fast())
			g, _ := errgroup.WithContext(context.Background())
			g.Go(func() error { return servers.ServeGrpc(grpcL) })
			g.Go(func() error {
				return servers.ServeHttp(httpL)
			})
			g.Go(m.Serve)
			return g.Wait()
		},
	})
}
