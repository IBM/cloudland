#ifndef _PRAGMA_COPYRIGHT_
#define _PRAGMA_COPYRIGHT_
#pragma comment(copyright, "%Z% %I% %W% %D% %T%\0")
#endif /* _PRAGMA_COPYRIGHT_ */
/****************************************************************************

* Copyright (c) 2008, 2010 IBM Corporation.
* All rights reserved. This program and the accompanying materials
* are made available under the terms of the Eclipse Public License v1.0s
* which accompanies this distribution, and is available at
* http://www.eclipse.org/legal/epl-v10.html

 Classes: ExtLauncher

 Description: Support External Laucher such as POE
   
 Author: Tu HongJ, Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/06/09 tuhongj      Initial code (D155101)
   10/10/12 ronglli      Add oom_adj codes

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include <assert.h>
#include <signal.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <errno.h>
#include <string.h>
#include <stdio.h>

#include "tools.hpp"
#include "sshfunc.hpp"
#include "log.hpp"
#include "exception.hpp"
#include "locker.hpp"
#include "stream.hpp"
#include "socket.hpp"

#include "extlaunch.hpp"

const int MAX_FD = 256;

vector<ExtLauncher *> launcherList;

const int MAX_PWD_BUF_SIZE = 1024;
const int MAX_ENV_VAR_NUM = 1024;

void set_oom_adj(int s)
{
#ifdef _SCI_LINUX
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
#endif
}

ExtLauncher::ExtLauncher(Stream *s, bool auth)
    : stream(s)
{
    memset(&usertok, 0, sizeof(usertok));
    sshAuth = auth;
}

ExtLauncher::~ExtLauncher()
{
}

int ExtLauncher::verifyToken(bool suser) 
{
    int rc;
    struct passwd pwd;
    struct passwd *result = NULL;
    char *pwdBuf = new char[MAX_PWD_BUF_SIZE];

    while (1) {
        rc = ::getpwnam_r(userName.c_str(), &pwd, pwdBuf, MAX_PWD_BUF_SIZE, &result);
        if ((rc == EINTR) || (rc == EMFILE) || (rc == ENFILE)) {
            SysUtil::sleep(WAIT_INTERVAL);
            continue;
        }
        if (NULL == result) {
            delete []pwdBuf;
            throw Exception(Exception::INVALID_USER);
        } else {
            break;
        }
    } 
    if (suser) {
        ::setgid(pwd.pw_gid);
        ::setuid(pwd.pw_uid);
    }
    ::seteuid(pwd.pw_uid);
    rc = SSHFUNC->verify_id_token(pwd.pw_name, &usertok);
    delete []pwdBuf;

    return rc;
}

int ExtLauncher::verifyData(struct iovec &sign, int jobkey, int id, char *path, char *envStr)
{
    int rc = -1;

    ssKeyLen = sizeof(sessionKey);
    rc = SSHFUNC->get_key_from_token(NULL, &usertok, sessionKey, &ssKeyLen);
    if (rc != 0)
        return rc;

    if (path == NULL) {
        rc = SSHFUNC->verify_data(sessionKey, ssKeyLen, &sign, "%d%d%d", mode, jobkey, id);
    } else {
        rc = SSHFUNC->verify_data(sessionKey, ssKeyLen, &sign, "%d%d%d%s%s", mode, jobkey, id, path, envStr);
    }

    return rc;
}

void ExtLauncher::run()
{
    int id, jobKey, rc;
    string path, envStr;
    struct iovec sign;
    try {
        *stream >> userName >> usertok >> sign >> (int &)mode >> jobKey >> id;
        switch (mode) {
            case INTERNAL:
                *stream >> path >> envStr >> endl;
                log_crit("[%s] Launch %d.%d %s with %s internally", userName.c_str(), 
                    jobKey, id, path.c_str(), envStr.c_str());
                launchInt(jobKey, id, (char *)path.c_str(), (char *)envStr.c_str(), sign);
                break;
            case REGISTER:
                *stream >> path >> envStr >> endl;
                log_crit("[%s] Receive register info %d.%d %s", userName.c_str(), jobKey, 
                        id, envStr.c_str());
                rc = doVerify(sign, jobKey, id, (char *)path.c_str(), (char *)envStr.c_str());
                if (rc == 0) {
                    rc = launchReg(jobKey, id, (char *)envStr.c_str());
                }
                break;
            case REQUEST:
                *stream >> endl;
                log_crit("[%s] Handle external launching request %d.%d", userName.c_str(),
                        jobKey, id);
                rc = doVerify(sign, jobKey, id);
                if (rc == 0) {
                    double starttm = SysUtil::microseconds();
                    rc = -1;
                    while ((rc != 0) && ((SysUtil::microseconds() - starttm) < FIVE_MINUTES)) {
                        rc = launchReq(jobKey, id);
                        SysUtil::sleep(WAIT_INTERVAL);
                    }
                }
                break;
            default:
                break;
        }
    } catch (SocketException &e) {
        log_error("socket exception %s", e.getErrMsg().c_str());
    } catch (Exception &e) {
        log_error("exception %s, errno = %d", e.getErrMsg(), errno);
    } catch (...) {
        log_error("unknown exception");
    }

    delete stream;
    delete [] (char *)usertok.iov_base;
    setState(false);

    Locker::getLocker()->lock();
    launcherList.push_back(this);
    Locker::getLocker()->unlock();
    Locker::getLocker()->notify();
}

char * ExtLauncher::getExename(char *path)
{
    int len;
    char *exename = NULL;
    char *p1 = NULL, *p2 = NULL;

    // get the exe name
    p1 = path;
    do {
        if (((*p1)==' ') || ((*p1)=='\t')) {
            p1++;
        } else if ((*p1) == '\0') {
            return NULL;
        } else {
            break;
        }
    } while (1);
    p2 = p1;
    while (((*p2)!=' ') && ((*p2)!='\t') && ((*p2)!='\0')) {
        p2++;
    }
    len = p2 - p1;
    exename = new char[len+1]; // Need to be deleted outside
    ::strncpy(exename, p1, len);
    exename[len] = '\0';

    return exename;
}

int ExtLauncher::launchInt(int jobkey, int id, char *path, char *envStr, struct iovec &signature)
{
    pid_t pid;
    int i = 0;
    int rc = 0;
    char *exename = getExename(path); // There is a new inside

    if (::access(exename, F_OK | R_OK | X_OK) != 0) {
        retStr = string(exename) + " is not an executable file";
        log_error("%s", retStr.c_str());
        delete [] exename;
        return -1;
    }

    pid = ::fork();
    if (pid < 0) { // fork failed
        rc =  errno;
        retStr = "fork failed";
    } else if (pid == 0) { // child process
        int sfd = stream->getSocket();
        // the child process can't ignore SIGCHLD signal
        ::sigaction(SIGCHLD, &oldSa, NULL);
        dup2(sfd, STDIN_FILENO);
        for (i = STDERR_FILENO + 1; i < MAX_FD; i++) {
            ::close(i);
        }
        set_oom_adj(1000);

        try {
            rc = putSessionKey(-1, signature, jobkey, id, path, envStr, true);
            if (rc != 0) {
                exit(rc);
            }
        } catch (Exception &e) {
            exit(-1);
        }

        char *p = envStr;
        char *params[4096];
        for (i = 0; i < MAX_ENV_VAR_NUM-1; i++) {
            p = ::strchr(p, ';');
            if (NULL == p) {
                break;
            }
            *p = '\0';
            params[i] = ++p;
        }
        params[i] = NULL;
        rc = ::execle("/bin/sh", "/bin/sh", "-c", path, (char *)NULL, params); 
        if (rc < 0) {
            exit(0);
        }
    }
    delete [] exename;

    return rc;
}

int ExtLauncher::getSessionKey(int fd)
{
    int n, rc;
    struct iovec vecs[2];

    vecs[0].iov_base = &rc;
    vecs[0].iov_len = sizeof(rc);
    vecs[1].iov_base = &sessionKey;
    vecs[1].iov_len = sizeof(sessionKey);
    if ((n = readv(fd, vecs, 2)) == -1) {
        rc = -1;
    }
    ssKeyLen = n - sizeof(rc);

    return rc;
}

int ExtLauncher::putSessionKey(int fd, struct iovec &sign, int jobkey, int id, char *path, char *envStr, bool suser)
{
    int i, rc;
    struct iovec vecs[2];

    rc = verifyToken(suser);
#ifdef PSEC_OPEN_SSL
    if (sshAuth) {
        if (rc == 0) {
            rc = verifyData(sign, jobkey, id, path, envStr);
        }
        if (fd < 0)
            return rc;

        vecs[0].iov_base = &rc;
        vecs[0].iov_len = sizeof(rc);
        vecs[1].iov_base = sessionKey;
        vecs[1].iov_len = ssKeyLen;
        writev(fd, vecs, 2);
    }
#endif

    return rc;
}

int ExtLauncher::doVerify(struct iovec &sign, int jobkey, int id, char *path, char *envStr)
{
#ifndef PSEC_OPEN_SSL
    return 0;
#endif
    int rc = -1;
    pid_t pid;
    int sockfd[2];

    if (!sshAuth)
        return 0;

    if (socketpair(AF_UNIX, SOCK_STREAM, 0, sockfd) == -1) {
        log_error("Failed to create socketpair!");
        return -1;
    }

    if ((pid = ::fork()) < 0) {
        rc =  errno;
        retStr = "fork failed";
        log_error("Failed to fork child!");
    } else if (pid == 0) { // child process
        int i;
        close(sockfd[1]);
        dup2(sockfd[0], MAX_FD);
        for (i = 0; i < MAX_FD; i++) {
            ::close(i);
        }
        rc = putSessionKey(MAX_FD, sign, jobkey, id, path, envStr);
        close(MAX_FD);
        exit(0);
    } else { // parent process
        close(sockfd[0]);
        rc = getSessionKey(sockfd[1]);
        close(sockfd[1]);
    }

    return rc;
} 


int ExtLauncher::launchReg(int jobkey, int id, const char *envStr)
{
    Locker::getLocker()->lock();
    TASK_INFO &task = jobInfo[jobkey];
    task.user = userName;
    task.config[id] = envStr;
    task.timestamp = SysUtil::microseconds();
    memset(&task.token, 0, sizeof(task.token));
    if (usertok.iov_len > 0) {
        task.token.iov_len = usertok.iov_len;
        task.token.iov_base = new char [usertok.iov_len];
        memcpy(task.token.iov_base, usertok.iov_base, usertok.iov_len);
    }

    Locker::getLocker()->unlock();

    return 0;
}

int ExtLauncher::launchReq(int jobkey, int id)
{
    struct iovec sign = {0};

    Locker::getLocker()->lock();
    if (jobInfo.find(jobkey) == jobInfo.end()) {
        Locker::getLocker()->unlock();
        return -1;
    }
    TASK_INFO &task = jobInfo[jobkey];
    if (task.user != userName) {
        Locker::getLocker()->unlock();
        return -2;
    }
    TASK_CONFIG &cfg = task.config;
    if (cfg.find(id) == cfg.end()) {
        Locker::getLocker()->unlock();
        return -1;
    }

    SSHFUNC->sign_data(sessionKey, ssKeyLen, &sign, 2, cfg[id].c_str(), cfg[id].size() + 1, task.token.iov_base, task.token.iov_len);
    *stream << cfg[id] << task.token << sign << endl;

    cfg.erase(id);
    if (cfg.size() == 0)
        jobInfo.erase(jobkey);
    Locker::getLocker()->unlock();

    return 0;
}

