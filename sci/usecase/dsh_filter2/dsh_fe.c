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

 Description: Distributed Shell.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   03/10/08 nieyy        Initial code (D156332)

****************************************************************************/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include <unistd.h>
#include "sci.h"

#define DOWN_FILTER 1
#define UP_FILTER 2
#define UP_FILTER_A 3
#define UP_FILTER_B 4

int num_resp;

void handler(void *user_param, sci_group_t group, void *buffer, int size)
{
    int be_id = ((int *) buffer)[0];
    char *msg = (char *)((char *) buffer + sizeof(int));
    char *pos = NULL;

    while (1) {
        pos = strstr(msg, "\n");
        if (pos == NULL) {
            break;
        } else {
            pos[0] = '\0';
        }
        printf("%d: %s\n", be_id, msg);
        msg = pos + 1;
    }
    
    num_resp++;
}

int main(int argc, char **argv)
{
    sci_info_t info;
    sci_filter_info_t filter_info;

    char msg[256];
    char *s;
    int i, rc, num_bes, job_key, sizes[1];
    void *bufs[1];

    char pwd[256];
    char hfile[256], bpath[256], fpath[256];

    getcwd(pwd, 256);
    sprintf(hfile, "%s/host.list", pwd);
#ifdef __64BIT__
    sprintf(bpath, "%s/dsh_be64", pwd);
#else
    sprintf(bpath, "%s/dsh_be", pwd);
#endif

    bzero(&info, sizeof(info));
    info.type = SCI_FRONT_END;
    info.fe_info.mode = SCI_INTERRUPT;
    info.fe_info.hostfile = hfile;
    info.fe_info.bepath = bpath;
    info.fe_info.hndlr = (SCI_msg_hndlr *)&handler;
    info.fe_info.param = NULL;
    
    rc = SCI_Initialize(&info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }
    rc = SCI_Query(NUM_BACKENDS, &num_bes);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }
    
#ifdef __64BIT__
    sprintf(fpath, "%s/downfilter64.so", pwd);
#else
    sprintf(fpath, "%s/downfilter.so", pwd);
#endif
    bzero(&filter_info, sizeof(filter_info));
    filter_info.filter_id = DOWN_FILTER;
    filter_info.so_file = fpath;
    rc = SCI_Filter_load(&filter_info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

#ifdef __64BIT__
    sprintf(fpath, "%s/upfilter64.so", pwd);
#else
    sprintf(fpath, "%s/upfilter.so", pwd);
#endif
    bzero(&filter_info, sizeof(filter_info));
    filter_info.filter_id = UP_FILTER;
    filter_info.so_file = fpath;
    rc = SCI_Filter_load(&filter_info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

#ifdef __64BIT__
    sprintf(fpath, "%s/upfiltera64.so", pwd);
#else
    sprintf(fpath, "%s/upfiltera.so", pwd);
#endif
    bzero(&filter_info, sizeof(filter_info));
    filter_info.filter_id = UP_FILTER_A;
    filter_info.so_file = fpath;
    rc = SCI_Filter_load(&filter_info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

#ifdef __64BIT__
    sprintf(fpath, "%s/upfilterb64.so", pwd);
#else
    sprintf(fpath, "%s/upfilterb.so", pwd);
#endif
    bzero(&filter_info, sizeof(filter_info));
    filter_info.filter_id = UP_FILTER_B;
    filter_info.so_file = fpath;
    rc = SCI_Filter_load(&filter_info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    while (1) {
        printf(">>> ");
        fflush(stdout);
        memset(msg, 0 , sizeof(msg));
        fgets(msg, sizeof(msg), stdin);
        msg[strlen(msg) - 1] = '\0';

        if (0 == strcmp(msg, "quit"))
            break;
        
        num_resp = 0;
        
        bufs[0] = msg;
        sizes[0] = strlen(msg) + 1;
        rc = SCI_Bcast(DOWN_FILTER, SCI_GROUP_ALL, 1, bufs, sizes);
        if (rc != SCI_SUCCESS) {
            exit(1);
        }
        
        while (num_resp < num_bes) {
            usleep(500);
        } 
    }
    
    rc = SCI_Terminate();
    
    return rc;
}

