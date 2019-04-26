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
#include <assert.h>
#include "sci.h"

#define DOWN_FILTER 1
#define UP_FILTER 2
#define UP_FILTER_A 3
#define UP_FILTER_B 4

#define RST_SIZE 4096

char *result = NULL;

void handler(void *user_param, sci_group_t group, void *buffer, int size)
{
    int bytes, my_id, rc;
    FILE *fp = NULL;
    int sizes[2];
    void *bufs[2];

    rc = SCI_Query(BACKEND_ID, &my_id);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }
    bufs[0] = &my_id;
    sizes[0] = sizeof(my_id);
    
    fp = popen((const char *)buffer, "r");
    assert(result != NULL);
    memset(result, 0, RST_SIZE);
    bytes = fread(result, sizeof(char), RST_SIZE, fp);
    bufs[1] = result;
    sizes[1] = strlen(result) + 1;
    pclose(fp);
    
    rc = SCI_Upload(UP_FILTER, group, 2, bufs, sizes);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }
}

int main(int argc, char **argv)
{
    sci_info_t info;
    int rc;

    result = (char *)malloc(RST_SIZE * sizeof(char));
    bzero(&info, sizeof(info));
    info.type = SCI_BACK_END;
    info.be_info.mode = SCI_INTERRUPT;
    info.be_info.hndlr = (SCI_msg_hndlr *)&handler;
    info.be_info.param = NULL;
    
    rc = SCI_Initialize(&info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    rc = SCI_Terminate();
    free(result);
    return rc;
}

