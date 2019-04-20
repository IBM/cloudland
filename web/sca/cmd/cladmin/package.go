package main

import (
	"context"

	"github.com/IBM/cloudland/web/sca/releases"
	"github.com/IBM/cloudland/web/sca/clients"
	"google.golang.org/grpc"
)

func init() {
	RootCmd.AddCommand(releases.Commands(context.Background,
		func() *grpc.ClientConn {
			return clients.GetClientConn("admin")
		})...)
}
