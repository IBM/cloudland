/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package servers

import (
	"context"
	"net"
	"net/http"

	"github.com/IBM/cloudland/web/sca/interceptors/xrequestid"
	"github.com/IBM/cloudland/web/sca/logs"
	"github.com/IBM/cloudland/web/sca/pkgs"
	"github.com/IBM/cloudland/web/sca/releases"
	"github.com/IBM/cloudland/web/sca/selfs"
	"github.com/IBM/cloudland/web/sca/servers/internal/stone"
	"github.com/IBM/cloudland/web/sca/targets"
	mid "github.com/grpc-ecosystem/go-grpc-middleware"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	tracing *logs.Tracing
)

func getTracing() *logs.Tracing {
	if tracing == nil {
		tracing = logs.StartTracing("ServerAdmin")
	}
	return tracing
}

func startLogging(ctx context.Context, name string) (*logs.SpanLogger, context.Context) {
	return getTracing().StartLogging(ctx, name)
}

func ServeGrpc(lis net.Listener) (err error) {
	err = newGrpcServer().Serve(lis)
	return
}

func newGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{}
	unarys := []grpc.UnaryServerInterceptor{}
	streams := []grpc.StreamServerInterceptor{}
	jaegerEnabled := viper.GetBool("jaeger.enabled")
	if jaegerEnabled {
		tracer := getTracing()
		opentracing.SetGlobalTracer(tracer)
		unarys = append(unarys, otgrpc.OpenTracingServerInterceptor(tracer))
		streams = append(streams, otgrpc.OpenTracingStreamServerInterceptor(tracer))
	}
	unarys = append(unarys, xrequestid.UnaryServerInterceptor)
	streams = append(streams, xrequestid.StreamServerInterceptor)
	// recovery interceptor
	unarys = append(unarys, recovery.UnaryServerInterceptor())
	streams = append(streams, recovery.StreamServerInterceptor())
	opts = append(opts, grpc.UnaryInterceptor(mid.ChainUnaryServer(unarys...)),
		grpc.StreamInterceptor(mid.ChainStreamServer(streams...)))
	s := grpc.NewServer(opts...)
	reflection.Register(s)
	pkgs.Register(s)
	selfs.Register(s)
	releases.Register(s)
	targets.Register(s)
	return s
}

func ServeHttp(lis net.Listener) (err error) {
	err = newHttpServer().Serve(lis)
	return
}

func newHttpServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/webhook", stone.HandleWebHook)
	mux.HandleFunc("/webhook/", stone.HandleWebHook)
	return &http.Server{
		Handler: mux,
	}
}
