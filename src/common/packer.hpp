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

 Classes: Packer

 Description: Wrapper for various kind of information.
   
 Author: Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code (D153875)

****************************************************************************/

#ifndef _PACKER_HPP
#define _PACKER_HPP

#include <vector>
#include <map>
#include <string>

using namespace std;

class Packer 
{
    private:
        char        *msgBuf;
        char        *msgPtr;
        int         msgLen;
        int         bufSize;
        
    public:
        Packer();
        Packer(char *msg);
        ~Packer();

        // for message packing usage
        void packInt(int value);
        void packStr(const char *value);
        void packStr(const string &value);
        void packStr(const string &value, int len);
        char * getPackedMsg();
        int getPackedMsgLen();

        // for message unpacking usage
        void setPackedMsg(const void *msg);
        int unpackInt();
        char * unpackStr();
        char * unpackStr(int *length);

    public:
        void checkBuffer(int size);
};

#endif

