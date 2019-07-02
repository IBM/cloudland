/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package selfs

import (
	"context"
	fmt "fmt"
	"os"
	"strings"
	"time"

	"github.com/IBM/cloudland/web/sca/logs"
	"github.com/IBM/cloudland/web/sca/pkgs"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	Version = "v1.0.0"
)

func init() {
	os.Setenv("CLADMIN_PID", pkgs.Pid)
}

var (
	tracing *logs.Tracing
)

func getTracing() *logs.Tracing {
	if tracing == nil {
		tracing = logs.StartTracing("SelfAdmin")
	}
	return tracing
}

func startLogging(ctx context.Context, name string) (*logs.SpanLogger, context.Context) {
	return getTracing().StartLogging(ctx, name)
}

type selfAdmin struct {
}

func Register(s *grpc.Server) {
	RegisterSelfAdminServer(s, &selfAdmin{})
}

func (adm *selfAdmin) Upgrade(ctx context.Context,
	req *UpgradeRequest) (rep *UpgradeReply, err error) {
	logger, ctx := startLogging(ctx, "Upgrade")
	defer logger.Finish()
	rep = &UpgradeReply{}
	version := req.GetVersion()
	if version == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no version specified")
		logger.Error(err)
		return
	}

	if version == Version {
		return
	}

	tr := pkgs.LoadPkgSum("cladmin", version)
	if tr.NotFound() {
		err = grpc.Errorf(codes.NotFound, "pkg cladmin=%s not found", version)
		logger.Error(err)
		return
	}
	args := []string{
		fmt.Sprintf("%v", os.Getpid()), version, pkgs.Executable,
	}
	args = append(args, os.Args[1:]...)
	err = pkgs.RunPartsDetach("upgrade", args...)
	if err == nil {
		go func() {
			time.Sleep(time.Second * 1)
			os.Exit(0)
		}()
	}
	return
}

func (adm *selfAdmin) Runtime(ctx context.Context,
	req *RuntimeRequest) (rep *RuntimeReply, err error) {
	logger, ctx := startLogging(ctx, "Version")
	defer logger.Finish()
	rep = &RuntimeReply{}
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	rep.Version = Version
	rep.Executable = pkgs.Executable
	rep.Pid = pkgs.Pid
	rep.Args = append(rep.Args, os.Args[1:]...)
	rep.Environ = func() (environ []string) {
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "CLADMIN_") {
				environ = append(environ, env)
			}
		}
		return
	}()
	rep.Pwd = pwd
	rep.Netrc = pkgs.NetrcEntries()
	return
}

func (adm *selfAdmin) Set(ctx context.Context,
	req *SetRequest) (rep *SetReply, err error) {
	logger, ctx := startLogging(ctx, "Set")
	defer logger.Finish()
	key := req.GetKey()
	value := req.GetValue()
	rep = &SetReply{}
	keyType := KeyClassify(key)
	switch keyType {
	case Unknown:
		err = grpc.Errorf(codes.InvalidArgument, "no valid key specified: %s", key)
		logger.Error(err)
		return
	case Env:
		if value == "" { // unset
			err = os.Unsetenv(key)
		} else {
			err = os.Setenv(key, value)
		}
	case Netrc:
		if value == "" { // unset
			err = pkgs.RemoveNetrc(key)
		} else {
			err = pkgs.AddNetrc(fmt.Sprintf("%s=%s", key, value))
		}
	}
	if err != nil {
		err = grpc.Errorf(codes.Internal, "%v", err)
	}
	return
}
