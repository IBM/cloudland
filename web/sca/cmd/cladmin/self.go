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

	"github.com/IBM/cloudland/web/sca/clients"
	"github.com/IBM/cloudland/web/sca/selfs"
	"google.golang.org/grpc"
)

func init() {
	RootCmd.AddCommand(selfs.Commands(context.Background,
		func() *grpc.ClientConn {
			return clients.GetClientConn("admin")
		})...)
}
