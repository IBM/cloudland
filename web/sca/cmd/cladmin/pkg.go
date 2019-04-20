package main

import (
	"context"

	"github.com/gabecloud/sca/clients"
	"github.com/gabecloud/sca/pkgs"
	"google.golang.org/grpc"
)

func init() {
	RootCmd.AddCommand(pkgs.Commands(context.Background,
		func() *grpc.ClientConn {
			return clients.GetClientConn("admin")
		})...)
}
