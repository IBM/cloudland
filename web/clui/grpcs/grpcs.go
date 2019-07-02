/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package grpcs

import (
	"github.com/IBM/cloudland/web/clui/hypers"
	"github.com/IBM/cloudland/web/clui/jobs"
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/IBM/cloudland/web/sca/logs"
	mid "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"net"
)

func Run() (err error) {
	opts := []grpc.ServerOption{}
	unarys := []grpc.UnaryServerInterceptor{}
	streams := []grpc.StreamServerInterceptor{}
	tracer := logs.StartTracing("HypercubeGrpc")
	opentracing.SetGlobalTracer(tracer)
	unarys = append(unarys, otgrpc.OpenTracingServerInterceptor(tracer))
	streams = append(streams, otgrpc.OpenTracingStreamServerInterceptor(tracer))
	// recovery interceptor
	unarys = append(unarys, grpc_recovery.UnaryServerInterceptor())
	streams = append(streams, grpc_recovery.StreamServerInterceptor())

	opts = append(opts, grpc.UnaryInterceptor(mid.ChainUnaryServer(unarys...)), grpc.StreamInterceptor(mid.ChainStreamServer(streams...)))
	address := viper.GetString("grpc.listen")
	s := grpc.NewServer(opts...)
	scripts.RegisterRemoteExecServer(s, GetRemoteExecService())
	jobs.RegisterJobServiceServer(s, GetJobService())
	hypers.RegisterHyperServiceServer(s, GetHyperService())
	dbs.Register(s)
	reflection.Register(s)
	db := dbs.DB()
	defer db.Close()
	var listen net.Listener
	listen, err = net.Listen("tcp", address)
	if err != nil {
		return
	}

	if err = s.Serve(listen); err != nil {
		return
	}
	return
}
