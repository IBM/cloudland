/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package targets

import (
	fmt "fmt"
	"os"
	"sort"
	"strings"

	"github.com/IBM/cloudland/web/sca/logs"
	"github.com/IBM/cloudland/web/sca/pkgs"
	"github.com/IBM/cloudland/web/sca/releases"
	context "golang.org/x/net/context"
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

type targetAdmin struct {
}

func Register(s *grpc.Server) {
	RegisterTargetAdminServer(s, &targetAdmin{})
}

func (adm *targetAdmin) Create(ctx context.Context,
	req *CreateRequest) (rep *Target, err error) {
	logger, ctx := startLogging(ctx, "Create")
	defer logger.Finish()
	rep = &Target{}
	name := req.GetName()
	relName := req.GetRelease().GetName()
	relVersion := req.GetRelease().GetVersion()
	if name == "" || relName == "" || relVersion == "" {
		err = grpc.Errorf(codes.InvalidArgument,
			"no target name, release name or release version specified")
		logger.Error(err)
		return
	}
	release := pkgs.LoadRelease(relName, relVersion)
	if release.NotFound() {
		err = grpc.Errorf(codes.NotFound,
			"release %s=%s not found", relName, relVersion)
		logger.Error(err)
		return
	}
	target := pkgs.LoadTarget(name)
	if !target.NotFound() {
		err = grpc.Errorf(codes.AlreadyExists,
			"target %s already exists", name)
		logger.Error(err)
		return
	}
	_, err = pkgs.RunParts("target-create", name, relName, relVersion)
	rep.Name = name
	rep.Release = &releases.Release{
		Name:    relName,
		Version: relVersion,
	}
	rep.State = State_CREATED
	return
}

func (adm *targetAdmin) Update(ctx context.Context,
	req *UpdateRequest) (rep *Target, err error) {
	logger, ctx := startLogging(ctx, "Update")
	defer logger.Finish()
	name := req.GetName()
	version := req.GetVersion()
	rep = &Target{}
	if name == "" || version == "" {
		err = grpc.Errorf(codes.InvalidArgument,
			"no target name or release version specified")
		logger.Error(err)
		return
	}
	target := pkgs.LoadTarget(name)
	release := target.GetString(name, "release")
	_, err = pkgs.RunParts("target-create", name, release, version)
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	target = pkgs.LoadTarget(name)
	rep.Name = name
	rep.Release = &releases.Release{
		Name:    target.GetString(name, "release"),
		Version: target.GetString(name, "version"),
	}
	rep.State = State(target.GetInt(name, "state"))
	return
}

func (adm *targetAdmin) Delete(ctx context.Context,
	req *DeleteRequest) (rep *Target, err error) {
	logger, ctx := startLogging(ctx, "Delete")
	defer logger.Finish()
	name := req.GetName()
	rep = &Target{}
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no target name specified")
		logger.Error(err)
		return
	}
	target := pkgs.LoadTarget(name)
	if target.NotFound() {
		err = grpc.Errorf(codes.NotFound, "target %s not found", target)
		logger.Error(err)
		return
	}
	err = target.Remove()
	if err != nil {
		return
	}
	rep.Name = name
	rep.Release = &releases.Release{
		Name:    target.GetString(name, "release"),
		Version: target.GetString(name, "version"),
	}
	rep.State = State_DELETED
	return
}

func (adm *targetAdmin) List(req *ListRequest,
	s TargetAdmin_ListServer) (err error) {
	ctx := s.Context()
	logger, ctx := startLogging(ctx, "Create")
	defer logger.Finish()
	names := pkgs.ReadDirNames("targets")
	if len(names) == 0 {
		err = grpc.Errorf(codes.NotFound, "no target found")
		logger.Error(err)
		return
	}
	for _, name := range names {
		target := pkgs.LoadTarget(name)
		if target.NotFound() {
			continue
		}
		err = s.Send(&Target{
			Name:  name,
			State: State((target.GetInt(name, "state"))),
			Release: &releases.Release{
				Name:    target.GetString(name, "release"),
				Version: target.GetString(name, "version"),
			},
		})
		if err != nil {
			logger.Error(err)
			return
		}
	}
	return
}

func (adm *targetAdmin) Deploy(ctx context.Context,
	req *DeployRequest) (rep *Target, err error) {
	logger, ctx := startLogging(ctx, "Create")
	defer logger.Finish()
	rep = &Target{}
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no deploy name specified")
		logger.Error(err)
		return
	}
	target := pkgs.LoadTarget(name)
	if target.NotFound() {
		err = grpc.Errorf(codes.NotFound, "target %s not found", name)
		logger.Error(err)
		return
	}
	_, err = pkgs.RunParts("target-deploy", name)
	if err != nil {
		logger.Error(err)
		err = grpc.Errorf(codes.Internal, "%v", err)
		return
	}
	target = pkgs.LoadTarget(name)
	rep.Name = name
	rep.Release = &releases.Release{
		Name:    target.GetString(name, "release"),
		Version: target.GetString(name, "version"),
	}
	rep.State = State((target.GetInt(name, "state")))
	return
}

func (adm *targetAdmin) Get(ctx context.Context,
	req *GetRequest) (rep *Target, err error) {
	logger, ctx := startLogging(ctx, "Create")
	defer logger.Finish()
	rep = &Target{}
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no target name specified")
		logger.Error(err)
		return
	}
	target := pkgs.LoadTarget(name)
	if target.NotFound() {
		err = grpc.Errorf(codes.NotFound, "target %s not found", name)
		logger.Error(err)
		return
	}
	rep.Name = name
	rep.Release = &releases.Release{
		Name:    target.GetString(name, "release"),
		Version: target.GetString(name, "version"),
	}
	rep.State = State((target.GetInt(name, "state")))
	return
}

func (adm *targetAdmin) Envs(req *EnvsRequest, s TargetAdmin_EnvsServer) (err error) {
	ctx := s.Context()
	logger, ctx := startLogging(ctx, "Envs")
	defer logger.Finish()
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no target specified")
		logger.Error(err)
		return
	}
	environs := pkgs.LoadEnvirons(name)
	if len(environs) == 0 {
		err = grpc.Errorf(codes.NotFound, "No environs found")
		logger.Error(err)
		return
	}
	key := req.GetEnv().GetName()
	value := req.GetEnv().GetValue()
	if key != "" { // setenv
		if pkgs.EnvReserved(key) {
			err = grpc.Errorf(codes.InvalidArgument, "%s is a reserved key", key)
			logger.Error(err)
			return
		}
		if value != "" { // set env
			environs[key] = value
		} else { // unset env
			delete(environs, key)
		}
		err = pkgs.SaveEnvirons(name, environs)
		if err != nil {
			logger.Error(err)
			err = grpc.Errorf(codes.Internal, "%v", err)
			return
		}
	}

	keys := []string{}
	for key := range environs {
		if environs[key] != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := environs[key]
		err = s.Send(&Env{Name: key, Value: value})
		if err != nil {
			return
		}
	}
	keys = []string{}
	for key := range environs {
		if environs[key] == "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		err = s.Send(&Env{Name: key})
		if err != nil {
			return
		}
	}
	return
}

func (adm *targetAdmin) Hosts(req *HostsRequest,
	rep TargetAdmin_HostsServer) (err error) {
	ctx := rep.Context()
	logger, ctx := startLogging(ctx, "Hosts")
	defer logger.Finish()
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no target specified")
		logger.Error(err)
		return
	}
	hostname := req.GetHost().GetName()
	ip := req.GetHost().GetIp()
	group := req.GetHost().GetGroup()

	if hostname != "" {
		filename := fmt.Sprintf("targets/%s/hosts/%s", name, hostname)
		if ip == "" { // remove
			err = os.RemoveAll(filename)
			if err != nil {
				logger.Error(err)
				err = grpc.Errorf(codes.NotFound, "%v", err)
				return
			}
			logger.Info("host", hostname, "removed")
		} else { // add
			if group == "" {
				idx := strings.LastIndex(hostname, "-")
				if idx == -1 {
					err = grpc.Errorf(codes.InvalidArgument, "no group specified")
					logger.Error(err)
					return
				}
				group = hostname[0:idx]
			}
			host := &Host{
				Name:  hostname,
				Group: group,
				Ip:    ip,
			}
			err = host.Save(name)
			if err != nil {
				logger.Error(err)
				err = grpc.Errorf(codes.Internal, "%v", err)
				return
			}
			logger.Info("host", hostname, "added")
		}
	}
	// List
	hnames := LoadHostnames(name)
	hgroups := pkgs.Strings([]string{})
	for _, hname := range hnames {
		host := &Host{Name: hname}
		err = host.Load(name, hname)
		if err != nil {
			logger.Error(err)
			err = grpc.Errorf(codes.Internal, "%v", err)
			return
		}
		err = rep.Send(host)
		if err != nil {
			return
		}
		hgroups = hgroups.Append(host.GetGroup())
	}
	logger.Info(len(hnames), "sent")
	pgroups := FindGroups(name)
	logger.Info(len(pgroups), "found")
	for _, group := range pgroups {
		if !hgroups.Contains(group) {
			host := &Host{
				Name:  "-",
				Group: group,
				Ip:    "-",
			}
			err = rep.Send(host)
			if err != nil {
				return
			}
		}
	}
	return
}

func (adm *targetAdmin) Keys(req *KeysRequest,
	rep TargetAdmin_KeysServer) (err error) {
	ctx := rep.Context()
	logger, ctx := startLogging(ctx, "Keys")
	defer logger.Finish()
	name := req.GetName()
	if name == "" {
		err = grpc.Errorf(codes.InvalidArgument, "no target specified")
		logger.Error(err)
		return
	}
	keyname := req.GetKey().GetName()
	if keyname != "" {
		content := req.GetKey().GetPrivate()
		key := &Key{Name: keyname, Private: content}
		if content == "" { // remove
			err = key.Remove(name)
		} else { // add
			err = key.Save(name)
		}

		if err != nil {
			logger.Error(err)
			grpc.Errorf(codes.Internal, "%v", err)
			return
		}
	}
	// List
	knames := pkgs.ReadFileNames(fmt.Sprintf("targets/%s/keys", name))
	for _, kname := range knames {
		key := &Key{Name: kname}
		err = key.Load(name, kname)
		if err != nil {
			logger.Error(err)
			err = grpc.Errorf(codes.Internal, "%v", err)
			return
		}
		err = rep.Send(key)
		if err != nil {
			logger.Error(err)
			err = grpc.Errorf(codes.Internal, "%v", err)
			return
		}
	}
	return
}
