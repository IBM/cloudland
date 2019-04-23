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

#ifndef _DSHPACKER_HPP
#define _DSHPACKER_HPP

#include <vector>
#include <map>
#include <string>

using namespace std;

class DshPacker {
    private:
        char *msgBuf;
        char *msgPtr;
        int     msgLen;
        
    public:
        DshPacker();

        // for message packing usage
        void packInt(int value);
        void packStr(char *value);
        char* getPackedMsg();
        int getPackedMsgLen();

        // for message unpacking usage
        void setPackedMsg(char *msg);
        int unpackInt();
        char *unpackStr();
};

#endif

