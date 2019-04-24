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
#include <unistd.h>
#include "sci.h"

#define DOWN_FILTER 1
#define UP_FILTER 2
#define UP_FILTER_A 3
#define UP_FILTER_B 4

int filter_initialize(void **user_param)
{
    // do nothing
    return SCI_SUCCESS;
}

int filter_terminate(void *user_param) 
{
    // do nothing
    return SCI_SUCCESS;
}

int filter_input(void *user_param, sci_group_t group, void *buf, int size) 
{
    void *bufs[1];
    int sizes[1];
    int rc;
    
    bufs[0] = buf;
    sizes[0] = size;

    rc = SCI_Filter_upload(UP_FILTER_B, group, 1, bufs, sizes);
    if (rc != SCI_SUCCESS) {
        // do something
    } 

    return SCI_SUCCESS;
}
