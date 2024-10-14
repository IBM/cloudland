#ifndef _RPCWORKER_HPP
#define _RPCWORKER_HPP

#include "netlayer.hpp"
#include "exception.hpp"
#include "httplib.h"
#include "threadpool.hpp" /* Added by nanjj */

using namespace std;
using namespace httplib;

class RemoteExecServiceImpl {
    private:
        NetLayer & sciNet;
        bool running;
    public:
        RemoteExecServiceImpl(NetLayer & sci);
        bool getState() { return running; }
        void Execute(const Request &request, Response &response);
        int exec_cmd(int id, char *cmd);
};

class FrontBack {
    public:
	FrontBack(string remoteHost, int remotePort);
        string Execute(int be_id, int msg_id, char *ctl, char *cmd, char *trace);
        void ExecuteAsync(int be_id, int msg_id, char *ctl, char *cmd, char *trace);

    private:
	int remotePort;
	string remoteHost;
        ::ThreadPool<100> threadpool;
};

class RpcWorker {
    private:
        NetLayer sciNet;
        FrontBack *rpcClient;
        void initConn();

    public:
        RpcWorker();
        ~RpcWorker();

        void runServer();
        FrontBack *getClient() { return rpcClient; }
};

#endif
