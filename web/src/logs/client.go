/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package logs

import (
	"google.golang.org/grpc"
)

func NewClientConn(endpoint string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	if len(opts) == 0 {
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.Dial(endpoint, opts...)
}
