package main

import (
	"context"

	"github.com/gabecloud/sca/clients"
	"github.com/gabecloud/sca/selfs"
	"google.golang.org/grpc"
)

func init() {
	RootCmd.AddCommand(selfs.Commands(context.Background,
		func() *grpc.ClientConn {
			return clients.GetClientConn("admin")
		})...)
}
