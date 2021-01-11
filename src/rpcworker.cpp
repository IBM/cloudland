/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <unistd.h>

#include "rpcworker.hpp"
#include "packer.hpp"
#include "log.hpp"

#include <thread>

#define CLOUDLET_PATH "/opt/cloudland/bin/cloudlet"
#define CLOUD_HOST_FILE "/opt/cloudland/etc/host.list"
#define GRPC_SERVER_ENDPOINT "0.0.0.0:50051"
#define GRPC_REMOTE_ENDPOINT "localhost:50050"

Status RemoteExecServiceImpl::Transmit(ServerContext* context, ServerReader<FileChunk>* reader, TransmitAck* ack)
{
    FileChunk chunk;
    bool first = true;
    string control;
    int node = -1;
    char *group = NULL;
    while (reader->Read(&chunk)) {
        if (first) {
            control = chunk.control();
            char *inter = strstr((char *)control.c_str(), "inter=");
            char *toall = strstr((char *)control.c_str(), "toall=");
            if (inter != NULL) {
                char *pnode = inter + strlen("inter=");

                if ((*pnode != '\0') && (*pnode != ' ')) {
                    node = atoi((const char *)pnode);
                }
                if (node < 0) {
                    ack->set_status("Invalid node");
                    return Status::OK;
                }
            } else if (toall != NULL) {
                char *name = toall + strlen("toall=");
                char *p = strchr(name, ' ');
                if (p != NULL) {
                    *p = '\0';
                }
                if (strcmp(name, "agent") == 0) {
                    ack->set_status("Invalid group");
                    return Status::OK;
                }
                group = name;
            } else {
                ack->set_status("Not supported");
                return Status::OK;
            }

            log_info("Transmit File: %d, control: %s, path: %s, size %lld", chunk.id(), (char *)control.c_str(), (char *)chunk.filepath().c_str(), chunk.filesize());
            if (control.find("type=") == string::npos) {
                control += " type=file";
            }
            first = false;
        }
        Packer packer;
        packer.packInt(chunk.id());
        packer.packInt(chunk.extra());
        packer.packStr(control);
        packer.packStr(chunk.filepath());
        packer.packInt(chunk.filesize());
        packer.packInt(chunk.checksum());
        packer.packInt(chunk.fileseek());
        packer.packStr(chunk.content(), chunk.content().size());
        char *message = packer.getPackedMsg();
        int length = packer.getPackedMsgLen();
        if (node >= 0) {
            sciNet.sendMessage(node, message, length);
        } else if (group != NULL) {
            sciNet.sendMessage(message, length, group, false);
        }
    }
    ack->set_status("OK");
    return Status::OK;
}

int RemoteExecServiceImpl::exec_cmd(int id, char *cmd)
{
    int bytes = 0;
    FILE *fp = NULL;
    char result[1024] = {0};
    char command[1024] = {0};

    snprintf(command, sizeof(command), "%s %s", cmd, "2>&1");
    fp = popen(cmd, "r");
    bytes = fread(result, sizeof(char), sizeof(result) - 1, fp);
    pclose(fp);
    log_info("Backend or agent %d responded command %s, result %s", id, cmd, result);

    return 0;
}

Status RemoteExecServiceImpl::Execute(ServerContext* context, const ExecuteRequest* request, ExecuteReply* reply)
{
    int msgID = request->id();
    int extra = request->extra();
    char *control = (char *)request->control().c_str();
    char *command = (char *)request->command().c_str();
    string trace = "";
    Packer packer;
    packer.packInt(msgID);
    packer.packInt(extra);
    packer.packStr(control);
    packer.packStr(command);
    char *inter = strstr(control, "inter=");
    char *select = strstr(control, "select=");
    char *group = strstr(control, "group=");
    char *toall = strstr(control, "toall=");
    char *mkgrp = strstr(control, "mkgrp=");
    char *rmgrp = strstr(control, "rmgrp=");
    char *lsgrp = strstr(control, "lsgrp=");
    char *term = strstr(control, "term=");
    char *callback = strstr(control, "callback");
    multimap<grpc::string_ref, grpc::string_ref> metadata = context->client_metadata();
    std::multimap<grpc::string_ref, grpc::string_ref>::iterator it = metadata.find("uber-trace-id");
    if (it != metadata.end()) {
        trace = string(it->second.data(), it->second.length());
    }
    packer.packStr(trace);
    char *message = packer.getPackedMsg();
    int length = packer.getPackedMsgLen();

    reply->set_status("OK");
    log_info("Received message id: %d, extra: %d, control: %s, command: %s, trace %s: ", msgID, extra, control, command, trace.c_str());
    try {
        if (inter != NULL) {
            int node = -1;
            char *pnode = inter + strlen("inter=");

            if ((*pnode != '\0') && (*pnode != ' ')) {
                node = atoi((const char *)pnode);
            }
            if (node >= 0) {
                sciNet.sendMessage(node, message, length);
                log_info("Message ID: %d control: %s command: %s was sent to node %d", msgID, control, command, node);
            } else {
                sciNet.sendMessage(message, length);
                log_info("Message ID: %d control: %s command: %s was sent to scheduler", msgID, control, command);
            }
        } else if (group != NULL) {
            log_info("Message ID: %d control: %s command: %s was sent to group", msgID, control, command);
            char *name = group + strlen("group=");
            char *p = strchr(group, ' ');
            if (p != NULL) {
                *p = '\0';
            }
            sciNet.sendMessage(message, length, name);
        } else if (select != NULL) {
            log_info("Message ID: %d control: %s command: %s was sent to group member", msgID, control, command);
            char *desc = select + strlen("select=");
            sciNet.createGroup(desc);
            char *p = strchr(select, ':');
            if (p != NULL) {
                *p = '\0';
            }
	    char *name = desc;
            sciNet.sendMessage(message, length, name, true);
        } else if (toall != NULL) {
            log_info("Message ID: %d control: %s command: %s was sent to all", msgID, control, command);
            char *name = toall + strlen("toall=");
            char *p = strchr(name, ' ');
            if (p != NULL) {
                *p = '\0';
            }
            if (strcmp(name, "agent") == 0) {
                sciNet.sendMessage(message, length, name);
            } else {
                sciNet.groupMessage(message, length, name);
            }
        } else if (mkgrp != NULL) {
            char *desc = mkgrp + strlen("mkgrp=");
            sciNet.createGroup(desc);
        } else if (rmgrp != NULL) {
            char *name = rmgrp + strlen("rmgrp=");
            sciNet.freeGroup(name);
        } else if (lsgrp != NULL) {
            string groupStr = sciNet.listGroup();
            reply->set_status(groupStr);
        } else if (term != NULL) {
            running = false;
        } else if (callback != NULL) {
            exec_cmd(extra, command);
        } else {
            reply->set_status("Not Found");
        }
    } catch (CommonException &e) {
        log_error(e.getErrMsg());
    }

    return Status::OK;
}

void FrontBack::ExecuteAsync(int msg_id, int extra, char *ctl, char *cmd, char *trace)
{
    ctl = strdup(ctl);
    cmd = strdup(cmd);
    trace = strdup(trace);
    threadpool.AddJob( [this, msg_id, extra, ctl, cmd, trace](){
            string reply = this->Execute(msg_id, extra, ctl, cmd, trace);
            free(ctl);
            free(cmd);
            free(trace);
            log_info("rpc_replied %s", reply.c_str());
        });
}

string FrontBack::Execute(int msg_id, int extra, char *ctl, char *cmd, char *trace)
{
	ExecuteRequest request;

	request.set_id(msg_id);
	request.set_extra(extra);
	request.set_control(ctl);
	request.set_command(cmd);
	ExecuteReply reply;

	grpc_connectivity_state state;
	int retried = 0;
	state = connChannel->GetState(true);
	while ((state != GRPC_CHANNEL_READY) && (state != GRPC_CHANNEL_CONNECTING) && (retried < 10)) {
		usleep(1000);
		state = connChannel->GetState(true);
		retried++;
	} 
	ClientContext context;
    context.AddMetadata("uber-trace-id", trace);
	Status status = stub_->Execute(&context, request, &reply);

	// Act upon its status.
	if (status.ok()) {
		return reply.status();
	} else {
        string errMsg = "RPC failed: " + status.error_message();
		log_info("RPC failed with code = %d, message = %s", status.error_code(), status.error_message().c_str());
		return errMsg;
	}
}

RemoteExecServiceImpl::RemoteExecServiceImpl(NetLayer & sci)
    : sciNet(sci), running(true)
{
}

RpcWorker::RpcWorker()
    : service(sciNet), rpcClient(NULL)
{
    initConn();
}

void RpcWorker::initConn() {
    char *envp = getenv("GRPC_REMOTE_ENDPOINT");
    if (envp == NULL) {
        envp = GRPC_REMOTE_ENDPOINT;
    }
    if (rpcClient != NULL) {
        delete rpcClient;
    }
    rpcClient = new FrontBack(grpc::CreateChannel(envp, grpc::InsecureChannelCredentials()));
}

RpcWorker::~RpcWorker()
{
    if (rpcClient != NULL) {
        delete rpcClient;
    }
}

void RpcWorker::runServer()
{
    char *bePath = getenv("CLOUDLET_PATH");
    char *hFile = getenv("CLOUD_HOST_FILE");
    if (bePath == NULL) {
        bePath = CLOUDLET_PATH;
    }
    if (hFile == NULL) {
        hFile = CLOUD_HOST_FILE;
    }
    try {
        sciNet.initFE(bePath, hFile, this);
    } catch (CommonException &e) {
        log_error(e.getErrMsg());
    }

    char *sAddr = getenv("GRPC_SERVER_ENDPOINT");
    if (sAddr == NULL) {
        sAddr = GRPC_SERVER_ENDPOINT;
    }
    string server_address(sAddr);
    ServerBuilder builder;
    builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
    builder.RegisterService(&service);
    unique_ptr<Server> server(builder.BuildAndStart());
    log_info("Server listening on %s", sAddr);
    thread server_thread{[&] {
        server->Wait();
    }};
    while (service.getState()) {
        sleep(2);
    }
    server->Shutdown();
    server_thread.join();
    sciNet.terminate();
    log_info("RPC worker terminated normally");
}
