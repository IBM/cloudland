/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package clients

import (
	"crypto/tls"
	"fmt"
	"os"
	"sync"

	"github.com/IBM/cloudland/web/sca/interceptors/xrequestid"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

const (
	API = "api"
)

var (
	clientconns = &sync.Map{}
	locker      = &sync.Mutex{}
)

func GetClientConn(name string) (conn *grpc.ClientConn) {
	return getClientConn(name)
}

func getClientConn(name string) (conn *grpc.ClientConn) {
	conn = clientConn(getEndpoint(name))
	return
}

func clientConn(endpoint string) (conn *grpc.ClientConn) {
	conn = load(endpoint)
	if conn == nil {
		conn = dial(endpoint)
	}
	return
}

func load(endpoint string) (conn *grpc.ClientConn) {
	if v, ok := clientconns.Load(endpoint); ok {
		conn = v.(*grpc.ClientConn)
		state := conn.GetState()
		if state <= connectivity.Ready {
			return
		} else {
			conn.Close()
			conn = nil
		}
	}
	return
}

func dial(endpoint string) (conn *grpc.ClientConn) {
	locker.Lock()
	defer locker.Unlock()
	if conn = load(endpoint); conn != nil {
		return
	}
	var err error
	var opts []grpc.DialOption
	if os.Getenv("TLS_ENABLE") == "true" {
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
		opts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	} else {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}
	unarys := []grpc.UnaryClientInterceptor{}
	streams := []grpc.StreamClientInterceptor{}

	if tracer := opentracing.GlobalTracer(); tracer != nil {
		unarys = append(unarys, otgrpc.OpenTracingClientInterceptor(tracer))
		streams = append(streams, otgrpc.OpenTracingStreamClientInterceptor(tracer))
	}
	unarys = append(unarys, xrequestid.UnaryClientInterceptor)
	streams = append(streams, xrequestid.StreamClientInterceptor)
	opts = append(opts,
		grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(unarys...)),
		grpc.WithStreamInterceptor(middleware.ChainStreamClient(streams...)))
	if conn, err = grpc.Dial(endpoint, opts...); err != nil {
		logrus.Fatal(err)
		return
	}
	clientconns.Store(endpoint, conn)
	return
}

func getProxy(name string, fallbacks ...string) (proxy string) {
	return getConfigString(name, "proxy", fallbacks...)
}

func getEndpoint(name string, fallbacks ...string) (endpoint string) {
	endpoint = getConfigString(name, "endpoint", fallbacks...)
	return
}

func getDebug(name string) (debugging bool) {
	return viper.GetBool(fmt.Sprintf("%s.%s", name, "debug"))
}

func getConfigString(name, config string, fallbacks ...string) (value string) {
	value = viper.GetString(fmt.Sprintf("%s.%s", name, config))
	if value == "" { // fall back to api value
		if len(fallbacks) == 0 {
			fallbacks = append(fallbacks, API)
		}
		for _, name = range fallbacks {
			if name == "" {
				break
			}
			value = viper.GetString(fmt.Sprintf("%s.%s", name, config))
			if value != "" {
				break
			}
		}
	}
	return
}
