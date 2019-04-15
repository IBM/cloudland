#ifndef _NETLAYER_HPP
#define _NETLAYER_HPP

#include <string>
#include <sci.h>
#include <vector>
#include <map>

#define SCHEDULE_FILTER 1
#define SCHEDULE_SO_FILE "/opt/cloudland/lib64/scheduler.so"

#include <pthread.h>

using namespace std;

class RpcWorker;

struct groupDesc {
    sci_group_t group;
    string      desc;
};

typedef map<string, groupDesc> GROUP_MAP;

class NetLayer {
    private:
        sci_info_t sciInfo;
        string bePath;
        string hFile;
        pthread_mutex_t mtx;
        pthread_mutex_t ser;
        GROUP_MAP groupMap;
        vector<string> string2Array(const string& str, char splitter);

    public:
        NetLayer();
        int initFE(char *backend, char *hostfile, RpcWorker *rpcWorker);
        int sendMessage(char *message, int length, char *grpName = NULL, bool useFilter = true);
        int sendMessage(int beID, char *message, int length);
        int createGroup(char *grpDesc);
        int freeGroup(char *grpName);
        void terminate();
        void lock();
        void unlock();
        void serialize();
        void deserialize();
        int groupMessage(char *message, int length, char *grpDesc);
        string listGroup();
        ~NetLayer();
};

#endif
