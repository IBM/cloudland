/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>
#include <signal.h>
#include <time.h>
#include <pthread.h>

#include "sci.h"

#include "rcmanager.hpp"
#include "packer.hpp"

#define SCHEDULE_FILTER 1
#define CLOCKID CLOCK_REALTIME
#define MAXHOST 512

extern "C" {

extern int report_availibility(int group, ResourceManager *rcMgr, int myID);
extern int report_topology(int msgID, int target);
extern long getValue(char *message, char *key, long *second);
extern int bcast_message(int msgID, int group, void *buffer, int size);

pthread_t rthread;

void *report_thread(void *param)
{
    int myID;
    int status = 0;
    SCI_Query(SCI_AGENT_ID, &myID);
    ResourceManager * rcManager = (ResourceManager *)param;
    while (status == 0) {
        sleep(5);
        report_availibility(SCI_GROUP_ALL, rcManager, myID);
        report_topology(0, myID);
        SCI_Query(HEALTH_STATUS, &status);
    }

    return NULL;
}

int filter_initialize(void **user_param)
{
    ResourceManager * rcManager = new ResourceManager();
    *user_param = rcManager;
    pthread_create(&rthread, NULL, &report_thread, rcManager);

    return SCI_SUCCESS;
}

int filter_terminate(void *user_param)
{
    pthread_join(rthread, NULL);
    ResourceManager *rcManager = (ResourceManager *)user_param;
    delete rcManager;

    return SCI_SUCCESS;
}

int bcast_message(int msgID, int group, void *buffer, int size) 
{
    int rc, num;
    int *children = NULL;

    rc = SCI_Group_query(group, GROUP_SUCCESSOR_NUM, &num);
    if (num > 0) {
        children = new int[num];
        rc = SCI_Group_query(group, GROUP_SUCCESSOR, children);
        rc = SCI_Filter_bcast(SCHEDULE_FILTER, num, children, 1, &buffer, &size);
        delete []children;
    }

    return rc;
}

int upload_message(int msgID, int myID, int group, char *control, char *message)
{
    int rc = -1;
    Packer packer;
    void *bufs[1];
    int sizes[1];

    packer.packInt(msgID);
    packer.packInt(myID);
    packer.packStr(control);
    packer.packStr(message);
    packer.packStr("");
    bufs[0] = packer.getPackedMsg();
    sizes[0] = packer.getPackedMsgLen();
    rc = SCI_Filter_upload(SCI_FILTER_NULL, group, 1, bufs, sizes);

    return rc;
}

int upload_topology(int *children, int num, int *recovery, int rnum, int myID, int msgID)
{
    int i, j, hndl, port, status;
    string control = "callback=agent";
    string message;
    char hostname[MAXHOST] = {0};
    char tmp[1024] = {0};

    if (num <= 0) {
        return -1;
    }

    for (i = 0; i < num; i++) {
        hndl = children[i];
        SCI_Query_host(hndl, hostname, sizeof(hostname));
        status = 1;
        for (j = 0; j < rnum; j++) {
            if (hndl == recovery[j]) {
                status = 10;
                break;
            }
        }
        snprintf(tmp, sizeof(tmp) - 1, "%d,%s,%d\n", hndl, hostname, status);
        message += tmp;
    }
    snprintf(tmp, sizeof(tmp) - 1, " id=%d", myID);
    control += tmp;
    SCI_Query(SCI_LISTENER_PORT, &port);
    snprintf(tmp, sizeof(tmp) - 1, " port=%d", port);
    control += tmp;
    snprintf(tmp, sizeof(tmp) - 1, " num=%d", num);
    control += tmp;
    ::gethostname(hostname, sizeof(hostname));
    snprintf(tmp, sizeof(tmp) - 1, " hostname=%s", hostname);
    control += tmp;
    upload_message(msgID, myID, SCI_GROUP_ALL, (char *)control.c_str(), (char *)message.c_str());

    return 0;
}

int report_topology(int msgID, int myID)
{
    static bool update = true;
    int num, rnum, rc;
    int *children = NULL;
    int *recovery = NULL;

    rc = SCI_Query(NUM_RECOVERY, &rnum);
    if (rnum > 0) {
        update  = true;
        recovery = new int[rnum];
        rc = SCI_Query(RECOVERY_LIST, recovery);
    } else {
        if ((!update) && (msgID == 0)) {
            return 0;
        }
        update = false;
    }

    rc = SCI_Group_query(SCI_GROUP_ALL, GROUP_SUCCESSOR_NUM, &num);
    if (num > 0) {
        children = new int[num];
        SCI_Group_query(SCI_GROUP_ALL, GROUP_SUCCESSOR, children);
        upload_topology(children, num, recovery, rnum, myID, msgID);
    }
    if (recovery != NULL) {
        delete []recovery;
        return 0;
    }
    if (children != NULL) {
        delete []children;
    }

    return 0;
}

int report_availibility(int group, ResourceManager *rcMgr, int myID)
{
    int rc = 0;
    char rcMsg[1024] = {0};

    if (rcMgr->ifUpdated()) {
        rcMgr->getTotalMsg(rcMsg, sizeof(rcMsg));
        upload_message(-1, myID, group, "report", rcMsg);
    }

    return rc;
}

long getValue(char *message, char *key, long *second)
{
    char *p;
    long value = 0;

    p = strstr(message, key);
    if (p != NULL) {
        p += strlen(key);
        value = atol(p);
        if (second != NULL) {
            p = strchr(p, '/');
            if (p != NULL) {
                *second = atol(p + 1);
            }
        }
    }

    return value;
}

int filter_input(void *user_param, sci_group_t group, void *buffer, int size)
{
    if (buffer == NULL) {
        return SCI_SUCCESS;
    }

    int rc;
    int myID;
    ResourceManager *rcManager = (ResourceManager *)user_param;
    Packer packer((char *)buffer);

    rc = SCI_Query(SCI_AGENT_ID, &myID);
    int msgID = packer.unpackInt();
    int beID = packer.unpackInt();
    char *control = packer.unpackStr();
    char *message = packer.unpackStr();
    char *trace = packer.unpackStr();
    char *inter = strstr(control, "inter=");
    char *grp = strstr(control, "group=");
    char *report = strstr(control, "report");
    char *toall = strstr(control, "toall=");
    char *select = strstr(control, "select=");

    if (report != NULL) {
        long total_cpu;
        long cpu = getValue(message, "cpu=", &total_cpu);
        long total_memory;
        long memory = getValue(message, "memory=", &total_memory);
        long total_disk;
        long disk = getValue(message, "disk=", &total_disk);
        long total_network;
        long network = getValue(message, "network=", &total_network);
        long total_load;
        long load = getValue(message, "load=", &total_load);
        rcManager->setAvailibility(beID, cpu, total_cpu, memory, total_memory, disk, total_disk, network, total_network, load, total_load);
    } else if ((inter != NULL) || (grp != NULL) || (select != NULL)) {
        int target = (int)getValue(control, "inter=", NULL);
        if (target < 0) {
            if (target == myID) {
                rc = report_topology(msgID, myID);
            } else {
                rc = bcast_message(msgID, group, buffer, size);
            }
        } else {
            long cpu = getValue(control, "cpu=", NULL);
            long memory = getValue(control, "memory=", NULL);
            long disk = getValue(control, "disk=", NULL);
            long network = (int)getValue(control, "network=", NULL);
            int bestID = rcManager->getBestBranch(cpu, memory, disk, network, group);
            if (bestID != -1) {
                rc = SCI_Filter_bcast(SCHEDULE_FILTER, 1, &bestID, 1, &buffer, &size);
            } else {
                rc = upload_message(msgID, myID, group, "error=resource", message);
            }
        }
    } else if (toall != NULL) {
        char *agent = strstr(control, "toall=agent");
        if (agent != NULL) {
            rc = report_topology(msgID, myID);
        }
        rc = bcast_message(msgID, group, buffer, size);
    } else {
        rc = upload_message(msgID, myID, group, control, message);
    }

    return SCI_SUCCESS;
}

}
