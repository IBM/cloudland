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

 Classes: DshPacker

 Description: Wrapper for various kind of information.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#include "dsh_packer.hpp"
#include <stdlib.h>
#include <arpa/inet.h>
#include <string.h>

DshPacker::DshPacker()
{
    msgBuf = (char *)malloc(1);
    msgPtr = msgBuf;
    msgLen = 0;
}

char* DshPacker::getPackedMsg()
{
    return msgBuf;
}

int DshPacker::getPackedMsgLen()
{
    return msgLen;
}

void DshPacker::packInt(int value)
{
    int size = htonl(value);

    int oldLen = msgLen;
    msgLen += sizeof(size);
    msgBuf = (char *)realloc(msgBuf, msgLen);
    msgPtr = msgBuf + oldLen;

    memcpy(msgPtr, &size, sizeof(size));
    msgPtr += sizeof(size);
}

void DshPacker::packStr(char *value)
{
    int len = strlen(value) + 1;
    packInt(len);

    int oldLen = msgLen;
    msgLen += len;
    msgBuf = (char *)realloc(msgBuf, msgLen);
    msgPtr = msgBuf + oldLen;

    memcpy(msgPtr, value, len);
    msgPtr += len;
}

void DshPacker::setPackedMsg(char * msg)
{
    msgBuf = msg;
    msgPtr = msgBuf;
}

int DshPacker::unpackInt()
{
    int size, value;
    memcpy(&size, msgPtr, sizeof(size));

    value = ntohl(size);
    msgPtr += sizeof(size);

    return value;
}

char* DshPacker::unpackStr()
{
    int len;
    char *value;

    len = unpackInt();
    value = msgPtr;
    msgPtr += len;

    return value;
}

