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

#ifndef _EXCEPTION_HPP
#define _EXCEPTION_HPP

class Exception
{
    public:
        enum CODE {
            MEM_BAD_ALLOC,
            GET_ADDR_INFO,
            INVALID_USER,
            SYS_CALL,
            INVALID_LAUNCH,
            INVALID_SIGNATURE
        };
        
    private:
        int        errCode;
        
    public:
        Exception(int code) throw();
        
        const char * getErrMsg() const throw();
        int getErrCode() const throw();
};

#endif

