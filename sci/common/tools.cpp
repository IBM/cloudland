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

 Classes: None

 Description: Tool functions.
   
 Author: Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code (D153875)
   11/26/10 ronglli      To add config file reading functions

****************************************************************************/

#include "tools.hpp"
#include <ctype.h>
#include <fcntl.h>
#include <signal.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <netdb.h>                                                              
#include <netinet/in.h>                                                         
#include <sys/socket.h>   
#include <string.h>
#include <sys/param.h>
#include <sys/time.h> 
#include <dlfcn.h>

#include <fstream>
#include "exception.hpp"
#include "log.hpp"

string SysUtil::itoa(int value)
{
    static char buffer[32];
    sprintf(buffer, "%d", value);
    return string(buffer);
}

string SysUtil::lltoa(long long value)
{
    static char buffer[128];
    sprintf(buffer, "%lld", value);
    return string(buffer);
}

double SysUtil::microseconds()
{
    struct timeval time_v;
    ::gettimeofday(&time_v, NULL);
    return time_v.tv_sec * 1e6 + time_v.tv_usec;
}

void SysUtil::sleep(int usecs)
{
    struct timespec req;
    req.tv_sec = usecs / 1000000;
    req.tv_nsec = (usecs % 1000000) * 1000;
    ::nanosleep (&req, NULL);
}

string SysUtil::get_hostname(const char * name)
{
    string uniquestring;
    
    struct addrinfo hints, *host = NULL;
    memset(&hints, 0, sizeof(struct addrinfo));
    hints.ai_flags = AI_CANONNAME | AI_NUMERICHOST;
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    int rc = ::getaddrinfo(name, NULL, &hints, &host);
    if (rc == EAI_NONAME) {
        hints.ai_flags = AI_CANONNAME;
        rc = ::getaddrinfo(name, NULL, &hints, &host);
    }
    if (rc < 0)
        throw Exception(Exception::GET_ADDR_INFO);
    uniquestring = host->ai_canonname;
    ::freeaddrinfo(host);

    return uniquestring;
}

char* SysUtil::get_path_name(const char *program)
{
    static char path[MAXPATHLEN];
    sprintf(path, "which %s", program);
    FILE *fp = popen(path, "r");
    if (!fp)
        return NULL;
    path[0] = '\0';
    fscanf(fp, "%s", path);
    pclose(fp);
    if (!path[0]){
        return NULL;
    }
    if (path[0] == '.' && path[1] == '/') {
        char save_path[MAXPATHLEN];
        strcpy(save_path, path);
        strcpy(path, getenv("PWD"));
        strcat(path, save_path + 1);
    }
    return path;
}

int SysUtil::read_config(const char* var, string & out_val)
{
#define FILE_PATH "/etc/sci.conf"
#define VAR_PATTERN "^([^= ][^= ]*=[^= ][^= ]*)$"
    int rc = -1;
    ifstream fs;

    string line;
    size_t pos = 0;
    string word;
    bool found = false;
    string::iterator it;

    if (var == NULL) {
        return -1;
    }

    fs.open(FILE_PATH);
    if (!fs) {
        return -1;
    }

    while(fs) {
        size_t tmpp = 0;
        getline(fs,line);
        for (it = line.begin(); it != line.end(); ){
            if ((*it == ' ') || *it == '\t')
                it = line.erase(it);
            else
                it++;
        }
        if (line.length() == 0){
            continue;
        } else if (line[0] == '#') {
            continue;         //The comments line will start with '#'
        }

        tmpp = line.find_first_of('#',0);
        pos = line.find_first_of('=',0);
        if ((pos == 0) || (pos == string::npos) || (pos >= tmpp)) { //Skip unused lines
            continue;
        }

        word = line.substr(0, pos);
        if (word.compare(var) == 0) {
            pos += 1;
            out_val = line.substr(pos, tmpp-pos); //The contents after '#' is regarded as comments
            found = true;
            break;
        } else {
            continue;
        }
    }

    fs.close();

    return (found ? 0 : -1);
}

