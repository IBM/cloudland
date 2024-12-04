/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import (
	"testing"

	"github.com/golang/protobuf/ptypes/any"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGrpcStatus(t *testing.T) {
	// Simulate to generate error on server side
	err := grpc.Errorf(codes.InvalidArgument, "Invalid argument: %s", "id")
	// Simulate to handle error in client side
	if st, ok := status.FromError(err); !ok {
		t.Fatal(ok)
	} else if code := st.Code(); code != codes.InvalidArgument {
		t.Fatal(code)
	} else if msg := st.Message(); msg != "Invalid argument: id" {
		t.Fatal(msg)
	} else if details := st.Details(); len(details) != 0 {
		t.Fatal(details)
	}
}

func TestStatusDetails(t *testing.T) {
	// generate status
	st := status.New(codes.InvalidArgument, "Invalid argument")
	st, err := st.WithDetails(&any.Any{
		TypeUrl: "http://any",
		Value:   []byte("invalid"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if st, ok := status.FromError(st.Err()); !ok {
		t.Fatal(ok)
	} else if details := st.Details(); len(details) != 1 {
		t.Fatal(details)
	}
}
