#ifndef _REMOTE_SERVER_HPP
#define _REMOTE_SERVER_HPP

#include <grpc++/grpc++.h>

#ifdef BAZEL_BUILD
#include "examples/protos/remotexec.grpc.pb.h"
#else
#include "remotexec.grpc.pb.h"
#endif

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;
using remotexec::ExecuteRequest;
using remotexec::ExecuteReply;
using remotexec::RemoteExec;

class RemoteExecServiceImpl final : public RemoteExec::Service 
{
    Status Execute(ServerContext* context, const ExecuteRequest* request, ExecuteReply* reply) override;
};

#endif
