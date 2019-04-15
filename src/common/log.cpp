/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <time.h>
#include <stdarg.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>

#include "log.hpp"

const char *logHeader[] = {
    "[CRIT]",
    "[ERROR]",
    "[WARN]",
    "[INFO]",
    "[DEBUG]",
    "[PERF]",
    "[OTHER]",
};

Loger *Loger::logger = NULL;

Loger::Loger()
{
}

Loger::~Loger()
{
}

void Loger::init(const char *directory, const char * filename, int level, int m)
{
    assert(filename);
    assert(directory);
    
    char node[256] = {0};
    gethostname(node, sizeof(node));

    logFile = filename;
    logDir = directory;
    sprintf(logPath, "%s/%s.%d.%s" , directory, node, (int)getpid(), filename);
    permitLevel = level;
    mode = m;
    unlink(logPath);
}

void Loger::rename(const char *directory, int level, int m)
{
    int rc = -1;
    char new_logPath[2 * MAX_PATH_LEN]; 
    char node[256] = {0};

    if ((level >= 0) && (permitLevel != level)) {
        permitLevel = level;
    }
    if (m != INVALID)
        mode = m;

    if (directory == NULL) 
        return; 

    if (logDir == string(directory))
        return;

    gethostname(node, sizeof(node));
    sprintf(new_logPath, "%s/%s.%s.%d" , directory, node, logFile.c_str(), (int)getpid());
    if (::access(logPath, F_OK) == 0) {
        rc = ::rename(logPath, new_logPath);
        if (rc != 0) {
            log_error("Unable to rename log file from %s to %s, rc is %d, errno=%d(%s)", 
                    logPath, new_logPath, rc, errno, strerror(errno));
        } else {
            sprintf(logPath, "%s", new_logPath);
            log_warn("Move log file from %s to %s", logDir.c_str(), directory);
            logDir = directory;
        }
    } else { 
        sprintf(logPath, "%s", new_logPath);
        log_warn("Move log file from %s to %s", logDir.c_str(), directory);
        logDir = directory;
    }
}

void Loger::print(int level, char *srcFile, int srcLine, const char *format, ...)
{
    if (mode != ENABLE)
        return;

    if(level > permitLevel)
        return;
    
    char tmMsg[MAX_LOG_LEN];
    time_t time1;
    struct tm tm1;
    va_list args;
    
    va_start(args, format);
    memset(tmMsg, 0, MAX_LOG_LEN);
    time(&time1);
    localtime_r(&time1, &tm1);
    strftime(tmMsg, MAX_LOG_LEN, "%y%m%d-%H:%M:%S", &tm1);
    
    FILE *fp = fopen(logPath, "a");
    if (fp) {
        fprintf(fp, "%s", tmMsg);
        fprintf(fp, " %s ", (char *)logHeader[level]);
        vfprintf(fp, format, args);
        fprintf(fp, " (%s:%d|%lu)\n", srcFile, srcLine, pthread_self());
        fclose(fp);
    }
    
    va_end(args);
}

