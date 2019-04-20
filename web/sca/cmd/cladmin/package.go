package main

import (
	"context"

	"github.com/gabecloud/sca/releases"
	"github.com/gabecloud/sca/clients"
	"google.golang.org/grpc"
)

func init() {
	RootCmd.AddCommand(releases.Commands(context.Background,
		func() *grpc.ClientConn {
			return clients.GetClientConn("admin")
		})...)
}
