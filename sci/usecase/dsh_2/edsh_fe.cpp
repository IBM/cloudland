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

 Description: Enhanced Distributed Shell.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/22/08 nieyy        Initial code (D154050)

****************************************************************************/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <assert.h>
#include <strings.h>
#include "sci.h"

#include <string>

using namespace std;

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
    int i, rc, num_bes, sizes[1];
    void *bufs[1];

    sci_group_t odd_group, even_group;
    int odd_size, even_size;
    int *odd_list = NULL, *even_list = NULL;

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
    info.fe_info.mode = SCI_POLLING;
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

    odd_size = (num_bes - num_bes%2) / 2;
    odd_list = (int *)malloc(sizeof(int) * odd_size);
    for (i=0; i<odd_size; i++) {
        odd_list[i] = i*2 + 1;
    }
    rc = SCI_Group_create(odd_size, odd_list, &odd_group);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    even_size = (num_bes + num_bes%2) / 2;
    even_list = (int *)malloc(sizeof(int) * even_size);
    for (i=0; i<even_size; i++) {
        even_list[i] = i*2;
    }
    rc = SCI_Group_create(even_size, even_list, &even_group);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    string cur_grp = "all";
    int exp_resp = num_bes;
    while (1) {
        printf("(%s) ", cur_grp.c_str());
        fflush(stdout);
        memset(msg, 0 , sizeof(msg));
        fgets(msg, sizeof(msg), stdin);
        msg[strlen(msg) - 1] = '\0';

        if (strcmp(msg, "help") == 0) {
            printf("Commands:\n");
            printf("\t\thelp -- show help topics\n");
            printf("\t\tall -- send commands to all back ends\n");
            printf("\t\teven -- send commands to back ends with even id\n");
            printf("\t\todd -- send commands to back ends with odd id\n");
            printf("\t\tquit -- exit this program\n");
            continue;
        } else if (strcmp(msg, "all") == 0) {
            cur_grp = "all";
            exp_resp = num_bes;
            continue;
        } else if (strcmp(msg, "odd") == 0) {
            cur_grp = "odd";
            exp_resp = odd_size;
            continue;
        } else if (strcmp(msg, "even") == 0) {
            cur_grp = "even";
            exp_resp = even_size;
            continue;
        } else if (strcmp(msg, "quit") == 0) {
            break;
        }
        
        num_resp = 0;
        
        bufs[0] = msg;
        sizes[0] = strlen(msg) + 1;
        if (cur_grp == "all") {
            rc = SCI_Bcast(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes);
        } else if (cur_grp == "odd") {
            rc = SCI_Bcast(SCI_FILTER_NULL, odd_group, 1, bufs, sizes);
        } else if (cur_grp == "even") {
            rc = SCI_Bcast(SCI_FILTER_NULL, even_group, 1, bufs, sizes);
        } else {
            assert(!"Unknown group");
        }
        if (rc != SCI_SUCCESS) {
            exit(1);
        }

        do {
            rc = SCI_Poll(-1);
            if (num_resp >= exp_resp) {
                break;
            }
        } while (rc == SCI_SUCCESS);
    }

    rc = SCI_Group_free(odd_group);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    rc = SCI_Group_free(even_group);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    free(odd_list);
    free(even_list);
    
    rc = SCI_Terminate();
    return rc;
}

