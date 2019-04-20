package main

import (
	"context"

	"github.com/gabecloud/sca/clients"
	"github.com/gabecloud/sca/targets"
	"google.golang.org/grpc"
)

func init() {
	RootCmd.AddCommand(targets.Commands(context.Background,
		func() *grpc.ClientConn {
			return clients.GetClientConn("admin")
		})...)
}
