/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package pkgs

import (
	"crypto/sha1"
	fmt "fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"context"

	"github.com/coduno/netrc"
	"github.com/IBM/cloudland/web/sca/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	tracing *logs.Tracing
)

func getTracing() *logs.Tracing {
	if tracing == nil {
		tracing = logs.StartTracing("PkgAdmin")
	}
	return tracing
}

func startLogging(ctx context.Context, name string) (*logs.SpanLogger, context.Context) {
	return getTracing().StartLogging(ctx, name)
}

type pkgAdmin struct {
}

func Register(s *grpc.Server) {
	RegisterPkgAdminServer(s, &pkgAdmin{})
	if filenames, err := Extract(); err == nil {
		fmt.Println("Extracted pkgNames:")
		for _, filename := range filenames {
			fmt.Println(filename)
		}
	}
}

func (adm *pkgAdmin) Upload(us PkgAdmin_UploadServer) (err error) {
	ctx := us.Context()
	logger, ctx := startLogging(ctx, "Upload")
	defer logger.Finish()
	var file *os.File
	h := sha1.New()
	name, version, deploy := "", "", false
	for {
		var pt *PkgChunk
		pt, err = us.Recv()
		if err == io.EOF {
			sha1sum := fmt.Sprintf("%x", h.Sum(nil))
			err = us.SendAndClose(&UploadReply{
				Pkg: &Pkg{
					Name:    name,
					Version: version,
					Sha1Sum: sha1sum,
					Deploy:  deploy,
				},
			})
			ioutil.WriteFile(
				fmt.Sprintf("packages/%s/%s/%s.sum", name, version, name),
				[]byte(PkgSum(name, version, sha1sum, deploy)), 0644)
			return
		} else if err != nil {
			err = grpc.Errorf(codes.Aborted, "%v", err)
			logger.Error(err)
			return
		}
		if file == nil {
			pkg := pt.GetPkg()
			name = pkg.GetName()
			version = pkg.GetVersion()
			deploy = pkg.GetDeploy()
			if name == "" || version == "" {
				err = grpc.Errorf(codes.InvalidArgument, "no pkg name or pkg version specified")
				logger.Error(err)
				return
			}
			d := fmt.Sprintf("packages/%s/%s", name, version)
			if err = os.MkdirAll(d, os.ModePerm); err != nil {
				err = grpc.Errorf(codes.Internal, "%v", err)
				logger.Error(err)
				return
			}

			filename := fmt.Sprintf("packages/%s/%s/%s.tgz", name, version, name)
			file, err = os.Create(filename)
			if err != nil {
				logger.Error(err)
				err = grpc.Errorf(codes.Internal, "%v", err)
				return
			}
			defer file.Close()
		}
		data := pt.GetChunk().GetData()
		h.Write(data)
		if _, err = file.Write(data); err != nil {
			logger.Error(err)
			err = grpc.Errorf(codes.Internal, "%v", err)
			return
		}
	}
	return
}

func (adm *pkgAdmin) Remove(ctx context.Context,
	req *RemoveRequest) (rep *RemoveReply, err error) {
	logger, ctx := startLogging(ctx, "Remove")
	defer logger.Finish()
	pkg := req.GetPkg()
	name := pkg.GetName()
	version := pkg.GetVersion()
	rep = &RemoveReply{}
	if name == "" || version == "" {
		err = grpc.Errorf(codes.InvalidArgument, "No pkg name or pkg version specified")
		logger.Error(err)
		return
	}
	tr := LoadPkgSum(name, version)
	if tr.NotFound() {
		err = grpc.Errorf(codes.NotFound, "%s=%s not found", name, version)
		logger.Error(err)
		return
	}
	p := fmt.Sprintf("packages/%s/%s", name, version)
	if err = os.RemoveAll(p); err != nil {
		err = grpc.Errorf(codes.Internal, "%v", err)
		logger.Error(err)
		return
	}
	rep.Pkg = &Pkg{
		Name:    name,
		Version: version,
		Sha1Sum: tr.GetString(name, "sha1sum"),
	}
	return
}

func (adm *pkgAdmin) List(req *ListRequest,
	rep PkgAdmin_ListServer) (
	err error) {
	ctx := rep.Context()
	logger, ctx := startLogging(ctx, "List")
	defer logger.Finish()
	pkgName := req.GetName()
	pkgNames := []string{}
	if pkgName != "" {
		pkgNames = append(pkgNames, pkgName)
	} else {
		pkgNames = ReadDirNames("packages")
	}

	if len(pkgNames) == 0 {
		err = grpc.Errorf(codes.NotFound, "No package found")
		logger.Error(err)
		return
	}

	for _, name := range pkgNames {
		versions := ReadDirNames(fmt.Sprintf("packages/%s", name))
		if len(versions) == 0 {
			logger.Warning(fmt.Sprintf("no version found in %s", name))
			continue
		}
		for _, version := range versions {
			tr := LoadPkgSum(name, version)
			sha1sum := tr.GetString(name, "sha1sum")
			if err != nil {
				err = nil
				continue
			}
			pkg := &Pkg{
				Name:    name,
				Version: version,
				Sha1Sum: sha1sum,
			}
			rep.Send(pkg)
		}
	}
	return
}

func (adm *pkgAdmin) Build(ctx context.Context,
	req *BuildRequest) (rep *BuildReply, err error) {
	logger, ctx := startLogging(ctx, "Build")
	defer logger.Finish()
	rep = &BuildReply{}
	p := req.GetProject()
	entries, err := netrc.Parse()
	url := "https://github.com"
	if err != nil {
		err = grpc.Errorf(codes.FailedPrecondition, "Run init before build: %v", err)
		logger.Error(err)
		return
	} else if _, ok := entries["github.com"]; !ok {
		err = grpc.Errorf(codes.FailedPrecondition, "Run init before build")
		logger.Error(err)
		return
	}
	if !strings.Contains(p.Name, "/") {
		p.Name = fmt.Sprintf("cland/%s", p.Name)
	}
	pkgs, err := p.Build(url)
	if err != nil {
		err = grpc.Errorf(codes.Internal, "build error: %v", err)
		logger.Error(err)
		return
	}
	rep.Pkgs = pkgs
	return
}
func (adm *pkgAdmin) Download(req *DownloadRequest,
	ds PkgAdmin_DownloadServer) (err error) {
	ctx := ds.Context()
	logger, ctx := startLogging(ctx, "Download")
	defer logger.Finish()
	pkg := req.GetPkg()
	name := pkg.GetName()
	version := pkg.GetVersion()
	if name == "" || version == "" {
		err = grpc.Errorf(codes.InvalidArgument, "No name or version specified")
		logger.Error(err)
		return
	}
	fp := fmt.Sprintf("packages/%s/%s/%s", name, version, name)
	tr := LoadPkgSum(name, version)
	if tr.NotFound() {
		err = grpc.Errorf(codes.NotFound, "No pkg found")
		logger.Error(err)
		return
	}
	sha1sum := tr.GetString(name, "sha1sum")
	if sha1sum == "" {
		err = grpc.Errorf(codes.NotFound, "No pkg content found")
		logger.Error(err)
		return
	}
	tgz := fp + ".tgz"
	file, err := os.Open(tgz)
	if err != nil {
		err = grpc.Errorf(codes.Internal, "%v", err)
		logger.Error(err)
		return
	}
	var buf [4096]byte
	position := int64(0)
	for {
		n := 0
		n, err = file.Read(buf[0:])
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = grpc.Errorf(codes.Internal, "%v", err)
			logger.Error(err)
			return
		}
		err = ds.Send(&PkgChunk{
			Pkg: &Pkg{
				Name:    name,
				Version: version,
				Sha1Sum: sha1sum,
			},
			Chunk: &Chunk{
				Data:     buf[0:n],
				Position: position,
			},
		})
		position += int64(n)
		if err != nil {
			err = grpc.Errorf(codes.Aborted, "%v", err)
			logger.Error(err)
			return
		}
	}
	return
}
