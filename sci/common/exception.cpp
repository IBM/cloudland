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

 Classes: Exception

 Description: Wrapper for SCI's exceptions.
   
 Author: Liu Wei, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 lwbjcdl      Initial code (D153875)

****************************************************************************/

#include "exception.hpp"
#include <assert.h>

const char * ErrMsg[] = {
    "Memory allocation failed.",
    "Error occur when call getaddrinfo.",
    "Invalid user credential.",
    "Error occur when call some system call.",
    "Invalid launch action.",
    "Invalid message signature."
};

Exception::Exception(int code) throw()
        : errCode(code)
{
}

const char * Exception::getErrMsg() const throw()
{
    return ErrMsg[errCode];
}

int Exception::getErrCode() const throw()
{
    return errCode;
}

