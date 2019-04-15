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
    int cpu;
    int total_cpu;
    int memory;
    int total_memory;
    int disk;
    int total_disk;
    int network;
    int total_network;
    int load;
    int total_load;
    int hitRate;
};

class ResourceManager {
    private:
        map<int, struct Resource *> rcMap;
        Resource totalRC;
        bool updated;

    public:
        ResourceManager();
        ~ResourceManager();
        int setAvailibility(int id, int cpu, int total_cpu, int memory, int total_memory, int disk, int total_disk, int network, int total_network, int load, int total_load);
        double testResource(Resource *resc, int cpu, int memory, int disk, int network);
        int totalResource();
        int getBestBranch(int cpu, int memory, int disk, int network, int group);
        int getTotalMsg(char *rcMsg, int size);
        bool ifUpdated() { return updated; }
};

#endif
