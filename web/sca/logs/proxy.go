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
	context "context"
	"io"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type traceProxy struct {
	jaegerURI string
}

func (tp *traceProxy) List(req *ListRequest, rep TraceAgentAdmin_ListServer) (err error) {
	cc, err := NewClientConn(tp.jaegerURI)
	if err != nil {
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	defer cc.Close()
	client := NewTraceAgentAdminClient(cc)
	ctx := context.Background()
	stream, err := client.List(ctx, &ListRequest{})
	if err != nil {
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	for {
		var out *ListReply
		out, err = stream.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			} else {
				err = grpc.Errorf(codes.Internal, "%v", err)
			}
			return
		}
		err = rep.Send(&ListReply{Data: out.GetData()})
		if err != nil {
			err = grpc.Errorf(codes.Aborted, "%v", err)
			return
		}
	}
	return
}

func RegisterProxy(server *grpc.Server, jaegerURI string) {
	RegisterTraceAgentAdminServer(server, &traceProxy{jaegerURI: jaegerURI})
}
