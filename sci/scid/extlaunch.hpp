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

****************************************************************************/

#ifndef _EXTLAUNCH_HPP
#define _EXTLAUNCH_HPP

#include <sys/uio.h>
#include <pwd.h>

#include <string>
#include <map>

#include "thread.hpp"

#define WAIT_INTERVAL 1000  // 1000 usec
#define FIVE_MINUTES 5000000 * 60

using namespace std;

class Stream;

enum LAUNCH_MODE {
    INTERNAL,
    REGISTER,
    REQUEST
};

class ExtLauncher : public Thread 
{
    private:
        Stream         *stream;
        string          retStr;
        string          userName;
        struct iovec    usertok;
        char            sessionKey[64];
        size_t          ssKeyLen;
        LAUNCH_MODE     mode;
        bool            sshAuth;

    private:
        char *getExename(char *path);

        int verifyToken(bool suser = false);
        int verifyData(struct iovec &sign, int jobkey, int id, char *path = NULL, char *envStr = NULL);
        int doVerify(struct iovec &sign, int jobkey, int id, char *path = NULL, char *envStr = NULL);
        int putSessionKey(int fd, struct iovec &sign, int jobkey, int id, char *path, char *envStr, bool suer = true);
        int getSessionKey(int fd);
    public:
        ExtLauncher(Stream *s, bool auth = false);
        virtual ~ExtLauncher();

        virtual void run();

        int launchInt(int jobkey, int id, char *path, char *envStr, struct iovec &sign);
        int launchReg(int jobkey, int id, const char *envStr);
        int launchReq(int jobkey, int id);
        int regInfo();
};

typedef map<int, string> TASK_CONFIG;
typedef struct TASK_INFO {
    string          user;
    TASK_CONFIG     config;
    double          timestamp;
    struct iovec    token;
};
typedef map<int, TASK_INFO> JOB_INFO;

extern JOB_INFO jobInfo;
extern vector<ExtLauncher *> launcherList;
extern struct sigaction oldSa;

void set_oom_adj(int s);

#endif

