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

 Description: Back End.
   
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
#define RST_SIZE 4096

void handler(void *user_param, sci_group_t group, void *buffer, int size)
{
    int seq_no = ((int *) buffer)[0];
    char *cmd = (char *)((char *) buffer + sizeof(int));

    int my_id;
    int rc = SCI_Query(BACKEND_ID, &my_id);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    char result[RST_SIZE] = {0};
    FILE *fp = popen((const char *)cmd, "r");
    memset(result, 0, RST_SIZE);
    int bytes = fread(result, sizeof(char), RST_SIZE, fp);
    pclose(fp);

    DshMessage msg;
    msg.readFromString(result, my_id);
    msg.setSeqNo(seq_no);

    int sizes[1];
    void *bufs[1];
    bufs[0] = msg.pack(&sizes[0]);

    rc = SCI_Upload(DSH_FILTER, group, 1, bufs, sizes);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    for (int i=0; i<msg.getSize(); i++)
        delete msg.getLine(i)->getLine();
}

int main(int argc, char **argv)
{
    sci_info_t info;
    int rc;

    bzero(&info, sizeof(info));
    info.type = SCI_BACK_END;
    info.be_info.hndlr = (SCI_msg_hndlr *)&handler;
    
    rc = SCI_Initialize(&info);
    if (rc != SCI_SUCCESS) {
        exit(1);
    }

    rc = SCI_Terminate();
    return rc;
}

