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

 Description: Front End.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include <unistd.h>
#include "sci.h"

#include "dsh_header.hpp"

#define DSH_FILTER 1

bool finished;
bool expand;

void handler(void *user_param, sci_group_t group, void *buffer, int size)
{
    DshMessage msg;
    msg.unpack(buffer);

    msg.print(expand);
    
    finished = true;
}

int main(int argc, char **argv)
{
    sci_info_t info;
    sci_filter_info_t filter_info;

    char msg[256];
    char *s;
    int i, rc, num_bes, seq, sizes[2];
    void *bufs[2];

    char pwd[256];
    char hfile[256], bpath[256], apath[256], fpath[256];

    getcwd(pwd, 256);
    sprintf(hfile, "%s/host.list", pwd);
#ifdef __64BIT__
    sprintf(bpath, "%s/dsh_be64", pwd);
#else
    sprintf(bpath, "%s/dsh_be", pwd);
#endif

    bzero(&info, sizeof(info));
    info.type = SCI_FRONT_END;
    info.fe_info.hostfile = hfile;
    info.fe_info.bepath = bpath;
    info.fe_info.hndlr = (SCI_msg_hndlr *)&handler;
    
    rc = SCI_Initialize(&info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }
    
    rc = SCI_Query(NUM_BACKENDS, &num_bes);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

#ifdef __64BIT__
    sprintf(fpath, "%s/dsh_filter64.so", pwd);
#else
    sprintf(fpath, "%s/dsh_filter.so", pwd);
#endif
    bzero(&filter_info, sizeof(filter_info));
    filter_info.filter_id = DSH_FILTER;
    filter_info.so_file = fpath;
    rc = SCI_Filter_load(&filter_info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    expand = false;

    seq = 0;
    while (1) {
        printf("(all) ");
        fflush(stdout);
        memset(msg, 0 , sizeof(msg));
        fgets(msg, sizeof(msg), stdin);
        msg[strlen(msg) - 1] = '\0';

        if (strcmp(msg, "quit") == 0)
            break;
        else if (strcmp(msg, "expand") == 0) {
            expand = !expand;
            if (expand)
                printf("Expand the output.\n");
            else
                printf("Do not expand the output.\n");
            continue;
        }
        
        finished = false;

        bufs[0] = &seq;
        sizes[0] = sizeof(seq);
        bufs[1] = msg;
        sizes[1] = strlen(msg) + 1;
        rc = SCI_Bcast(SCI_FILTER_NULL, SCI_GROUP_ALL, 2, bufs, sizes);
        if (rc != SCI_SUCCESS) {
            exit(1);
        }
        seq++;
       
        while (!finished) {
            usleep(500);
        } 
    }
    
    rc = SCI_Terminate();
    
    return rc;
}

