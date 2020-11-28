/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <sys/types.h>
#include <sys/stat.h>
#include <signal.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <fcntl.h>
#include <string.h>
#include <stdlib.h>
#include <sci.h>
#include <errno.h>
#include <unistd.h>

#include <fstream>

#include "log.hpp"
#include "exception.hpp"
#include "netlayer.hpp"
#include "packer.hpp"

const char * REPORT_RC_CMD = "/opt/cloudland/scripts/backend/report_rc.sh";

string pidFile;
string hostList;

void backHandler(void *user_param, sci_group_t group, void *buffer, int size)
{
    int rc, my_id;
    char *p = (char *)buffer;
    int ctLen = 0;
    Packer packer((char *)buffer);
    int id = packer.unpackInt();
    int extra = packer.unpackInt();
    char *control = packer.unpackStr();
    char *command = NULL;
    char *inter = strstr(control, "inter=");
    char *select = strstr(control, "select=");
    char *toall = strstr(control, "toall=");
    char *grp = strstr(control, "group=");
    char *type = strstr(control, "type=file");

    rc = SCI_Query(BACKEND_ID, &my_id);
    if (rc != SCI_SUCCESS) {
        return;
    }
    if (toall != NULL) {
        p = strstr(toall, "toall=agent");
        if (p != NULL) {
            return;
        }
    } else if (inter != NULL) {
        p = inter + strlen("inter=");
        if ((p == NULL) || (atoi(p) < 0)) {
            return;
        }
    }

    if (type != NULL) {
        char *filepath = packer.unpackStr();
        int filesize = packer.unpackInt();
        int checksum = packer.unpackInt();
        int fileseek = packer.unpackInt();
        int clen;
        char *content = packer.unpackStr(&clen);
        ofstream file;
        if (fileseek == 0) {
            file.open(filepath, fstream::binary | fstream::out);
        } else {
            file.open(filepath, fstream::binary | fstream::in | fstream::out);
        }
        file.seekp(fileseek);
        file.write(content, clen);
        file.close();
        return;
    }
    command = packer.unpackStr();
    if ((inter != NULL) || (toall != NULL) || (grp != NULL) || (select != NULL)) {
        int bytes = 0;
        void *bufs[1];
        int sizes[1];
        FILE *fp = NULL;
        char tmp[1024] = {0};
        int code;
        string cmdStr = command;
        char *trace = packer.unpackStr();

        setenv("JAEGER_TRACE_ID", trace, 1);
        cmdStr = "sudo -E " + cmdStr + " 2>&1";
        cmdStr = cmdStr + " 2>&1";
        fp = popen(cmdStr.c_str(), "r");
        char *p = fgets(tmp, sizeof(tmp), fp);
        while (p != NULL) {
            Packer resp;
            resp.packInt(id);
            resp.packInt(my_id);
            resp.packStr("callback");
            resp.packStr(tmp);
            resp.packStr(trace);
            bufs[0] = resp.getPackedMsg();
            sizes[0] = resp.getPackedMsgLen();
            rc = SCI_Upload(SCI_FILTER_NULL, group, 1, bufs, sizes);
            p = fgets(tmp, sizeof(tmp), fp);
        }
        rc = fclose(fp);
        code = WEXITSTATUS(rc);
        unsetenv("JAEGER_TRACE_ID");
        if (code != 0) {
            Packer resp;
            resp.packInt(id);
            resp.packInt(my_id);
            resp.packStr("error");
            resp.packStr(command);
            resp.packStr(trace);
            bufs[0] = resp.getPackedMsg();
            sizes[0] = resp.getPackedMsgLen();
            rc = SCI_Upload(SCI_FILTER_NULL, group, 1, bufs, sizes);
        }
    }
}

int startBE()
{
    int rc;

    sci_info_t sciInfo;
    memset(&sciInfo, 0, sizeof(sciInfo));
    sciInfo.type = SCI_BACK_END;
    sciInfo.be_info.mode = SCI_INTERRUPT;
    sciInfo.be_info.hndlr = (SCI_msg_hndlr *)&backHandler;
    sciInfo.be_info.param = NULL;
    sciInfo.enable_recover = 1;

    rc = SCI_Initialize(&sciInfo);
    if (rc != SCI_SUCCESS) {
        throw CommonException(CommonException::SCI_INIT_ERROR);
    }

    return 0;
}

void set_oom_adj(int s)
{
    char oomfname[256] = "";
    char oomadjstr[16] = "";
    int oomfd = -1;
    struct stat oomadjst = {0};
    int nbytes = -1;

    int score = s;
    score = (score < -1000) ? (-1000) : (score);
    score = (score > 1000) ? (1000) : (score);

    ::sprintf(oomfname, "/proc/%d/oom_score_adj", getpid());
    if (::stat(oomfname, &oomadjst) != 0) {
        /* could not find oom_score_adj, will try oom_adj instead */
        ::sprintf(oomfname, "/proc/%d/oom_adj", getpid());
        double oomadjval = (score < 0) ? (((double) score/1000.0)*17.0) : (((double) score/1000.0)*15.0);
        ::sprintf(oomadjstr, "%.0f", oomadjval);
    } else {
        ::sprintf(oomadjstr, "%d", score);
    }

    oomfd = ::open(oomfname, O_WRONLY, 0);
    if (oomfd < 0) {
        log_error("open() failed for %s: errno = %d\n", oomfname, errno);
        return;
    }

    nbytes = ::write(oomfd, oomadjstr, strlen(oomadjstr));
    if (nbytes < 0) {
        log_error("write() failed for %s: errno = %d", oomfname, errno);
    } else {
        log_crit("wrote %d bytes to %s: %s", nbytes, oomfname, oomadjstr);
    }
    ::close(oomfd);
}

int main(int argc, char *argv[])
{
    int rc, bytes, myID;
    int status = 0;
    int msgID = 0;
    char result[1024] = {0};
    char ctl[16] = "report";
    FILE *fp = NULL;
    sigset_t sigs_to_block;
    sigset_t old_sigs;

    set_oom_adj(-1000);

    startBE();

    rc = SCI_Query(BACKEND_ID, &myID);
    if (rc != SCI_SUCCESS) {
        exit(-1);
    }
    sigemptyset(&sigs_to_block);
    pthread_sigmask(SIG_SETMASK, &sigs_to_block, &old_sigs);
    setsid();

    while (status == 0) {
        void *bufs[1];
        int sizes[1];
        Packer packer;
        packer.packInt(msgID);
        packer.packInt(myID);
        packer.packStr("report");
        char *p = NULL;
        memset(result, '\0', sizeof(result));
        fp = popen(REPORT_RC_CMD, "r");
        p = fgets(result, sizeof(result) - 1, fp);
        packer.packStr(result);
        packer.packStr("");
        bufs[0] = packer.getPackedMsg();
        sizes[0] = packer.getPackedMsgLen();
        rc = SCI_Upload(SCHEDULE_FILTER, SCI_GROUP_ALL, 1, bufs, sizes);
        do {
            p = fgets(result, sizeof(result), fp);
            if (p == NULL) {
                break;
            }
            Packer resp;
            resp.packInt(msgID);
            resp.packInt(myID);
            resp.packStr("callback");
            resp.packStr(result);
            resp.packStr("");
            bufs[0] = resp.getPackedMsg();
            sizes[0] = resp.getPackedMsgLen();
            rc = SCI_Upload(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes);
        } while (true);
        pclose(fp);
        rc = SCI_Query(HEALTH_STATUS, &status);
        sleep(random() % 5 + 1);
    }
    SCI_Terminate();

    return 0;
}
