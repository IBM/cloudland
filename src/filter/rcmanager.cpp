/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <unistd.h>

#include <sci.h>

#include "rcmanager.hpp"

ResourceManager::ResourceManager()
    : updated(false)
{
    memset(&totalRC, 0, sizeof(totalRC));
}

ResourceManager::~ResourceManager()
{
    map<int, struct Resource *>::iterator it = rcMap.begin();
    for (; it != rcMap.end(); ++it) {
        delete it->second;
    }
    rcMap.clear();
}

double ResourceManager::testResource(Resource *resc, long cpu, long memory, long disk, long network)
{
    double current = 0.0;
    if (resc->cpu < 0 || resc->memory < 0 || resc->disk < 0) {
        return current;
    }
    current = (1 - cpu / (double)(resc->cpu + 1));
    if (current <= 0.0) {
        return current;
    }
    current *= (1 - memory / (double)(resc->memory + 1));
    if (current <= 0.0) {
        return current;
    }
    current *= (1 - disk / (double)(resc->disk + 1));

    return current;
}

int ResourceManager::getBestBranch(long cpu, long memory, long disk, long network, int group)
{
    int bestID = -1;
    static int count = 0;
    double current = 0.0;
    int i, j, idx;
    int num, rnum;
    int *children = NULL;
    int *recovery = NULL;
    int child;

    SCI_Group_query(group, GROUP_SUCCESSOR_NUM, &num);
    if (num > 0) {
        children = new int[num];
        SCI_Group_query(group, GROUP_SUCCESSOR, children);
    }
    if (children == NULL) {
        return -1;
    }
    SCI_Query(NUM_RECOVERY, &rnum);
    if (rnum > 0) {
        recovery = new int[rnum];
        SCI_Query(RECOVERY_LIST, recovery);
    }

    int fastID = count % num;
    count++;
    for (i = fastID; i < num + fastID; i++) {
        idx = i % num;
        child = children[idx];
        if (rcMap.find(child) == rcMap.end()) {
            count++;
            continue;
        }
        for (j = 0; j < rnum; j++) {
            if (child == recovery[j]) {
                break;
            }
        }
        if (j < rnum) {
            count++;
            continue;
        }

        current = testResource(rcMap[child], cpu, memory, disk, network);
        if (current > 0.0) {
            bestID = child;
            break;
        }
    }
    if (children != NULL) {
        delete []children;
    }
    if (recovery != NULL) {
        delete []recovery;
    }

    return bestID;
}

int ResourceManager::setAvailibility(int id, long cpu, long total_cpu, long memory, long total_memory, long disk, long total_disk, long network, long total_network, long load, long total_load)
{
    map<int, struct Resource *>::iterator it = rcMap.find(id);
    Resource *resource = NULL;
    if (it == rcMap.end()) {
        resource = new Resource();
        rcMap[id] = resource;
    } else {
        resource = it->second;
    }
    memset(resource, 0, sizeof(Resource));
    resource->cpu = cpu;
    resource->total_cpu = total_cpu;
    resource->memory = memory;
    resource->total_memory = total_memory;
    resource->disk = disk;
    resource->total_disk = total_disk;
    resource->network = network;
    resource->total_network = total_network;
    resource->load = load;
    resource->total_load = total_load;
    totalResource();
    updated = true;

    return 0;
}

int ResourceManager::totalResource()
{
    memset(&totalRC, 0, sizeof(totalRC));

    map<int, struct Resource *>::iterator it = rcMap.begin();
    for (; it != rcMap.end(); ++it) {
        totalRC.cpu += it->second->cpu;
        totalRC.total_cpu += it->second->total_cpu;
        totalRC.memory+= it->second->memory;
        totalRC.total_memory+= it->second->total_memory;
        totalRC.disk+= it->second->disk;
        totalRC.total_disk+= it->second->total_disk;
        totalRC.network+= it->second->network;
        totalRC.total_network+= it->second->total_network;
        totalRC.load += it->second->load;
        totalRC.total_load += it->second->total_load;
    }

    return 0;
}

int ResourceManager::getTotalMsg(char *rcMsg, int size)
{
    memset(rcMsg, '\0', size);
    snprintf(rcMsg, size - 1, "report_rc.sh 'cpu=%d/%d' 'memory=%ld/%ld' 'disk=%ld/%ld' 'network=%d/%d' 'load=%d/%d'\n", totalRC.cpu, totalRC.total_cpu, totalRC.memory, totalRC.total_memory, totalRC.disk, totalRC.total_disk, totalRC.network, totalRC.total_network, totalRC.load, totalRC.total_load);

    return 0;
}
