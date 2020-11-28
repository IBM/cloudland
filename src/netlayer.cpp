/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <string.h>
#include <string>
#include <sstream>

#include "netlayer.hpp"
#include "handler.hpp"
#include "exception.hpp"
#include "rpcworker.hpp"

using namespace std;

NetLayer::NetLayer()
{
    ::pthread_mutex_init(&mtx, NULL);
    ::pthread_mutex_init(&ser, NULL);
}

NetLayer::~NetLayer()
{
    groupMap.clear();
    ::pthread_mutex_destroy(&mtx);
    ::pthread_mutex_destroy(&ser);
}

void NetLayer::terminate()
{
    SCI_Terminate();
}

void NetLayer::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void NetLayer::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

void NetLayer::serialize()
{
    ::pthread_mutex_lock(&ser);
}

void NetLayer::deserialize()
{
    ::pthread_mutex_unlock(&ser);
}

int NetLayer::initFE(char *backend, char *hostfile, RpcWorker *rpcWorker)
{
    int rc;
    char *envp = getenv("SCHEDULE_SO_FILE");
    if ((envp == NULL)) {
        envp = SCHEDULE_SO_FILE;
    }
    sci_filter_info_t filter = {SCHEDULE_FILTER, envp};
    sci_filter_list_t flist = {1, &filter};

    bePath = backend;
    hFile = hostfile;
    memset(&sciInfo, 0, sizeof(sciInfo));
    sciInfo.type = SCI_FRONT_END;
    sciInfo.fe_info.mode = SCI_INTERRUPT;
    sciInfo.fe_info.hostfile = (char *)hFile.c_str();
    sciInfo.fe_info.bepath = (char *)bePath.c_str();
    sciInfo.fe_info.hndlr = (SCI_msg_hndlr *)&frontHandler;
    sciInfo.fe_info.filter_list = flist;
    sciInfo.fe_info.param = rpcWorker;
    sciInfo.enable_recover = 1;

    rc = SCI_Initialize(&sciInfo);
    if (rc != SCI_SUCCESS) {
        throw CommonException(CommonException::SCI_INIT_ERROR);
    }

    return 0;
}

string NetLayer::listGroup()
{
    string groupStr;
    GROUP_MAP::iterator it;
    
    lock();
    for (it = groupMap.begin(); it != groupMap.end(); ++it) {
        groupStr += it->second.desc + "\n";
    }
    unlock();

    return groupStr;
}

int NetLayer::freeGroup(char *grpName) 
{
    int rc = -1;
    lock();
    if (groupMap.find(grpName) != groupMap.end()) {
        rc = SCI_Group_free(groupMap[grpName].group);
        groupMap.erase(grpName);
    }
    unlock();

    return rc;
}

vector<string> NetLayer::string2Array(const string& str, char splitter)
{
    vector<string> tokens;
    stringstream ss(str);
    string temp;
    while (getline(ss, temp, splitter)) {
        tokens.push_back(temp);
    }
    return tokens;
}

int NetLayer::createGroup(char *grpDesc) 
{
    string savedDesc = grpDesc;
    char *p = strchr(grpDesc, ':');
    char *name = grpDesc;
    char *desc = NULL;
    char *r = NULL;
    int size = 256;
    if (p == NULL) {
        return -1;
    }
    *p = 0;
    if (groupMap.find(name) != groupMap.end()
		    && (groupMap[name].desc == savedDesc)) {
	    return 0;
    }
    vector<int> result;
    vector<string> tokens = string2Array(p + 1, ',');
    vector<string>::const_iterator it;
    for (it = tokens.begin(); it != tokens.end(); ++it) {
        const string& token = *it;
        vector<string> range = string2Array(token, '-');
        if (range.size() == 1) {
            result.push_back(atoi(range[0].c_str()));
        } else if (range.size() == 2) {
            int start = atoi(range[0].c_str());
            int stop = atoi(range[1].c_str());
            for (int j = start; j <= stop; j++) {
                result.push_back(j);
            }
        }
    }
    lock();
    groupMap[name].group = SCI_GROUP_ALL;
    groupMap[name].desc = savedDesc;
    if (result.size() > 0) {
        sci_group_t group;
        SCI_Group_create(result.size(), &result[0], &group);
        groupMap[name].group = group;
    }
    unlock();

    return 0;
}

int NetLayer::sendMessage(int beID, char *message, int length)
{
    int rc = -1;
    void *bufs[1];
    int sizes[1];

    bufs[0] = message;
    sizes[0] = length;
    rc = SCI_Bcast(SCI_FILTER_NULL, beID, 1, bufs, sizes);
    if (rc != SCI_SUCCESS) {
        throw CommonException(CommonException::SCI_BCAST_ERROR); 
    }

    return 0;
}

int NetLayer::groupMessage(char *message, int length, char *grpDesc)
{
    string desc = grpDesc;
    string::size_type pos = desc.find(":");
    if (pos != desc.npos) {
        string name = desc.substr(0, pos);
        serialize();
        createGroup(grpDesc);
        sendMessage(message, length, (char *)name.c_str(), false);
        freeGroup((char *)name.c_str());
        deserialize();
    } else {
        sendMessage(message, length, (char *)desc.c_str(), false);
    }

    return 0;
}

int NetLayer::sendMessage(char *message, int length, char *grpName, bool useFilter)
{
    int rc = -1;
    void *bufs[1];
    int sizes[1];
    int group = SCI_GROUP_ALL;
    int filter = SCHEDULE_FILTER;

    bufs[0] = message;
    sizes[0] = length;
    if (!useFilter) {
        filter = SCI_FILTER_NULL;
    }
    lock();
    if ((grpName != NULL) && (groupMap.find(grpName) != groupMap.end())) {
        group = groupMap[grpName].group;
    }
    rc = SCI_Bcast(filter, group, 1, bufs, sizes);
    unlock();
    if (rc != SCI_SUCCESS) {
        throw CommonException(CommonException::SCI_BCAST_ERROR); 
    }

    return 0;
}

