/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package releases

import (
	"fmt"
	"os"
	"time"

	"context"

	"github.com/IBM/cloudland/web/sca/logs"
	"github.com/IBM/cloudland/web/sca/pkgs"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	tracing *logs.Tracing
)

func getTracing() *logs.Tracing {
	if tracing == nil {
		tracing = logs.StartTracing("ReleaseAdmin")
	}
	return tracing
}

func startLogging(ctx context.Context, name string) (*logs.SpanLogger, context.Context) {
	return getTracing().StartLogging(ctx, name)
}

type releaseAdmin struct {
}

func Register(s *grpc.Server) {
	RegisterReleaseAdminServer(s, &releaseAdmin{})
}

func (adm *releaseAdmin) Create(ctx context.Context,
	req *CreateRequest) (rep *Release, err error) {
	logger, ctx := startLogging(ctx, "Create")
	defer logger.Finish()
	rep = &Release{}
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no name specified")
		logger.Error(err)
		return
	}
	version := "latest"
	p := fmt.Sprintf("releases/%s/%s", name, version)
	var fi os.FileInfo
	if fi, err = os.Stat(p); err == nil && fi.IsDir() {
		err = grpc.Errorf(codes.AlreadyExists, "%s already exists", name)
		logger.Error(err)
		return
	}
	err = os.MkdirAll(p, os.ModePerm)
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	p = fmt.Sprintf("%s/release.toml", p)
	f, err := os.Create(p)
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	f.Close()
	rep.Name = name
	rep.Version = version
	return
}

func (adm *releaseAdmin) Delete(ctx context.Context,
	req *DeleteRequest) (rep *Release, err error) {
	logger, ctx := startLogging(ctx, "Delete")
	defer logger.Finish()
	rep = &Release{}
	name := req.GetName()
	version := req.GetVersion()
	if name == "" || version == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no release name or version specified")
		logger.Error(err)
		return
	}
	p := fmt.Sprintf("releases/%s/%s", name, version)
	tr := pkgs.LoadRelease(name, version)
	if tr.NotFound() {
		err = grpc.Errorf(codes.NotFound, "%s/%s not found", name, version)
		logger.Error(err)
		return
	}
	err = os.RemoveAll(p)
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	rep.Name = name
	rep.Version = version
	rep.Pkgs = tr.GetPkgs()
	return
}

func (adm *releaseAdmin) List(req *ListRequest,
	svr ReleaseAdmin_ListServer) (err error) {
	ctx := svr.Context()
	logger, ctx := startLogging(ctx, "List")
	defer logger.Finish()
	name := req.GetName()
	names := []string{}
	if name == "" {
		names = pkgs.ReadDirNames("releases")
	} else {
		names = append(names, name)
	}

	if len(names) == 0 {
		return
	}
	for _, name := range names {
		versions := pkgs.ReadDirNames(fmt.Sprintf("releases/%s", name))
		if len(versions) == 0 {
			continue
		}
		for _, version := range versions {
			tr := pkgs.LoadRelease(name, version)
			if tr.NotFound() {
				continue
			}
			ps := tr.GetPkgs()
			err = svr.Send(&Release{
				Name:    name,
				Version: version,
				Pkgs:    ps,
			})
			if err != nil {
				logger.Error(err)
				grpc.Errorf(codes.Internal, "%v", err)
				return
			}
		}
	}
	return
}

func (adm *releaseAdmin) Refresh(ctx context.Context,
	req *RefreshRequest) (rep *Release, err error) {
	logger, ctx := startLogging(ctx, "Refresh")
	defer logger.Finish()
	name := req.GetName()
	rep = &Release{}
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no release name specified")
		logger.Error(err)
		return
	}
	release := pkgs.LoadRelease(name, "latest")
	if release.NotFound() {
		err = grpc.Errorf(codes.NotFound, "release %s not found", name)
		logger.Error(err)
		return
	}
	ps := release.GetPkgs()
	for _, p := range ps {
		pkg := p.Name
		latest := pkgs.LatestVersion(pkg)
		sha1sum := p.Sha1Sum
		if latest != p.Version {
			sha1sum, err = adm.addPkg(ctx, name, pkg, latest)
			if err != nil {
				return
			}
			p.Version = latest
			p.Sha1Sum = sha1sum
		}
	}
	rep.Name = name
	rep.Version = "latest"
	rep.Pkgs = ps
	return
}

func (adm *releaseAdmin) Get(ctx context.Context,
	req *GetRequest) (rep *Release, err error) {
	logger, ctx := startLogging(ctx, "Get")
	defer logger.Finish()
	name := req.GetName()
	version := req.GetVersion()
	rep = &Release{}
	if name == "" || version == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no release name or version specified")
		logger.Error(err)
		return
	}
	tr := pkgs.LoadRelease(name, version)
	ps := tr.GetPkgs()
	rep.Name = name
	rep.Version = version
	rep.Pkgs = ps
	return
}

func (adm *releaseAdmin) addPkg(ctx context.Context,
	name string, pkg, version string) (sha1sum string, err error) {
	logger, ctx := startLogging(ctx, "addPkg")
	defer logger.Finish()
	sum := pkgs.LoadPkgSum(pkg, version)
	if sum.NotFound() {
		err = grpc.Errorf(codes.NotFound, "pkg %s=%s not found", pkg, version)
		logger.Error(err)
		return
	}
	deploy := fmt.Sprintf("%v", sum.GetDeploy())
	_, err = pkgs.RunParts("release-add", name, pkg, version, deploy)
	if err == nil {
		sha1sum = sum.GetString(pkg, "sha1sum")
	}
	return
}

func (adm *releaseAdmin) Add(ctx context.Context,
	req *AddRequest) (rep *AddReply, err error) {
	logger, ctx := startLogging(ctx, "Add")
	defer logger.Finish()
	rep = &AddReply{}
	name := req.GetName()
	pkg := req.GetPkg().GetName()
	version := req.GetPkg().GetVersion()
	if name == "" || pkg == "" || version == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no release name, or pkg specified")
		logger.Error(err)
		return
	}
	release := pkgs.LoadRelease(name, "latest")
	if release.NotFound() {
		err = grpc.Errorf(codes.NotFound, "release %s not found", name)
		logger.Error(err)
		return
	}
	sha1sum, err := adm.addPkg(ctx, name, pkg, version)
	rep.Name = name
	rep.Pkg = &pkgs.Pkg{
		Name:    pkg,
		Version: version,
		Sha1Sum: sha1sum,
	}
	return
}

func (adm *releaseAdmin) Remove(ctx context.Context,
	req *RemoveRequest) (rep *RemoveReply, err error) {
	logger, ctx := startLogging(ctx, "Remove")
	defer logger.Finish()
	rep = &RemoveReply{}
	name := req.GetName()
	pkgName := req.GetPkg().GetName()
	if name == "" || pkgName == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no release name, or pkg specified")
		logger.Error(err)
		return
	}
	release := pkgs.LoadRelease(name, "latest")
	if release.NotFound() {
		err = grpc.Errorf(codes.NotFound, "release %s not found", name)
		logger.Error(err)
		return
	}
	pkg := release.GetPkg(pkgName)
	if pkg == nil {
		err = grpc.Errorf(codes.NotFound, "pkg %s not found in release", pkg)
		logger.Error(err)
		return
	}

	deploy := fmt.Sprintf("%v", pkg.GetDeploy())
	_, err = pkgs.RunParts("release-remove", name, pkgName, deploy)
	if err != nil {
		return
	}
	rep.Name = name
	rep.Pkg = pkg
	return
}

func (adm *releaseAdmin) Publish(ctx context.Context,
	req *PublishRequest) (rep *Release, err error) {
	logger, ctx := startLogging(ctx, "Publish")
	defer logger.Finish()
	rep = &Release{}
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no release name specified")
		logger.Error(err)
		return
	}
	release := pkgs.LoadRelease(name, "latest")
	if release.NotFound() {
		err = grpc.Errorf(codes.NotFound, "release %s not found", name)
		logger.Error(err)
		return
	}
	version := time.Now().Format("20060102150405")
	if !pkgs.LoadRelease(name, version).NotFound() {
		err = grpc.Errorf(codes.AlreadyExists,
			"release %s=%s already being published", name, version)
		logger.Error(err)
		return
	}
	if !release.GetDeploy() {
		err = grpc.Errorf(codes.FailedPrecondition, "No deploy pkg found in latest release")
		logger.Error(err)
		return
	}
	_, err = pkgs.RunParts("release-publish", name, version)
	if err == nil {
		release = pkgs.LoadRelease(name, version)
		if release.NotFound() {
			err = grpc.Errorf(codes.Internal,
				"release %s=%s failed to publish", name, version)
			logger.Error(err)
			return
		}
		rep.Name = name
		rep.Version = version
		rep.Pkgs = release.GetPkgs()
	}
	return
}
