/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <string.h>
#include <stdio.h>

#include <string>

#include "handler.hpp"
#include "netlayer.hpp"
#include "rpcworker.hpp"
#include "log.hpp"
#include "packer.hpp"

using namespace std;

const char * RESCUE_CMD = "/opt/cloudland/scripts/frontend/rescue.sh";

int exec_cmd(int id, char *cmd)
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

void frontHandler(void *user_param, sci_group_t group, void *buffer, int size)
{
    if (size < 16) {
        return;
    }

    Packer packer((char *)buffer);
    int msg_id = packer.unpackInt();
    int be_id = packer.unpackInt();
    char *ctl = packer.unpackStr();
    char *msg = packer.unpackStr();
    char *trace = packer.unpackStr();
    char *callback = strstr(ctl, "callback");
    char *error = strstr(ctl, "error");
    char *report = strstr(ctl, "report");
    RpcWorker *rpcWorker = (RpcWorker *)user_param;

    log_info("Cloudlet %d responded message id: %d control: %s content: %s", be_id, msg_id, ctl, msg);
    if (error != NULL) {
	rpcWorker->getClient()->ExecuteAsync(msg_id, be_id, ctl, msg, trace);
    } else if (callback != NULL) {
        char *cmd = NULL;
        char *next = NULL;
        char *tail = NULL;

        if (strstr(callback, "callback=agent") != NULL) {
            rpcWorker->getClient()->ExecuteAsync(msg_id, be_id, ctl, msg, trace);
            return;
        }
        cmd = strstr(msg, "|:-COMMAND-:|");
        while (cmd != NULL) {
            cmd += strlen("|:-COMMAND:-|") + 1;
            next = strstr(cmd, "|:-COMMAND-:|");
            tail = strchr(cmd, '\n');
            if (tail != NULL) {
                *tail = '\0';
            }
            rpcWorker->getClient()->ExecuteAsync(msg_id, be_id, ctl, cmd, trace);
            cmd = next;
        }
    } else if (report != NULL) {
        if (be_id == -1) {
            rpcWorker->getClient()->ExecuteAsync(msg_id, be_id, ctl, msg, trace);
        }
    }
}
