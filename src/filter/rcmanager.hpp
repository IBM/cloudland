/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#ifndef _RCMANAGER_HPP
#define _RCMANAGER_HPP

#include <map>
#include <string>

using namespace std;

struct Resource {
    long cpu;
    long total_cpu;
    long memory;
    long total_memory;
    long disk;
    long total_disk;
    long network;
    long total_network;
    long load;
    long total_load;
    long hitRate;
};

class ResourceManager {
    private:
        map<int, struct Resource *> rcMap;
        Resource totalRC;
        bool updated;

    public:
        ResourceManager();
        ~ResourceManager();
        int setAvailibility(int id, long cpu, long total_cpu, long memory, long total_memory, long disk, long total_disk, long network, long total_network, long load, long total_load);
        double testResource(Resource *resc, long cpu, long memory, long disk, long network);
        int totalResource();
        int getBestBranch(long cpu, long memory, long disk, long network, int group);
        int getTotalMsg(char *rcMsg, int size);
        bool ifUpdated() { return updated; }
};

#endif
