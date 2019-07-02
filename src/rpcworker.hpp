#ifndef _RPCWORKER_HPP
#define _RPCWORKER_HPP

#include <grpc++/grpc++.h>
#include "remotexec.grpc.pb.h"

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::ServerReader;
using grpc::Channel;
using grpc::ClientContext;
using grpc::ClientWriter;
using grpc::Status;
using com::ibm::cloudland::scripts::ExecuteRequest;
using com::ibm::cloudland::scripts::ExecuteReply;
using com::ibm::cloudland::scripts::FileChunk;
using com::ibm::cloudland::scripts::TransmitAck;
using com::ibm::cloudland::scripts::RemoteExec;

#include "netlayer.hpp"
#include "exception.hpp"
#include "threadpool.hpp" /* Added by nanjj */
using namespace std;

class RemoteExecServiceImpl final : public RemoteExec::Service
{
    private:
        Status Execute(ServerContext* context, const ExecuteRequest* request, ExecuteReply* reply) override;
        Status Transmit(ServerContext* context, ServerReader<FileChunk>* reader, TransmitAck* ack) override;
        int exec_cmd(int id, char *cmd);
        NetLayer & sciNet;
        bool running;
    public:
        RemoteExecServiceImpl(NetLayer & sci);
        bool getState() { return running; }
};

class FrontBack {
    public:
		shared_ptr<Channel> connChannel;
        FrontBack(shared_ptr<Channel> channel)
            : stub_(RemoteExec::NewStub(channel)) { connChannel = channel; }
        string Execute(int be_id, int msg_id, char *ctl, char *cmd, char *trace);
        void ExecuteAsync(int be_id, int msg_id, char *ctl, char *cmd, char *trace);
    private:
        unique_ptr<RemoteExec::Stub> stub_;
        ThreadPool<100> threadpool;
};

class RpcWorker {
    private:
        NetLayer sciNet;
        RemoteExecServiceImpl service;
        FrontBack *rpcClient;
        void initConn();

    public:
        RpcWorker();
        ~RpcWorker();

        void runServer();
        FrontBack *getClient() { return rpcClient; }
};

#endif
