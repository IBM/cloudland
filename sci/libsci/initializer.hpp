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

 Classes: Initializer

 Description: Prepare the environment when startup, which includes:
        1) Processor threads
        2) Message queue
        3) Others like environment variables
   
 Author: Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#ifndef _INITIALIZER_HPP
#define _INITIALIZER_HPP

#include <string>

#include "ctrlblock.hpp" 

using namespace std;

class Stream;
class Listener;

#define SCID_PORT 6188
#define SCID_NAME "sciv1" 

class Initializer
{
    public:
        bool        pInfoUpdated;
        int         notifyID;
    private:
        Initializer();
        static Initializer *instance;
        Listener    *listener;
        Stream      *inStream;
    
        int         handle;
        string      parentAddr;
        int         parentPort;
        int         parentID;

        string      initEnv;

    public:
        ~Initializer();
        static Initializer* getInstance() {
            if (instance == NULL)
                instance = new Initializer();
            return instance;
        }

        int init();
        string getEnvStr();
        Listener * initListener();
        Listener * getListener();
        Stream * getInStream();
        void setInStream(Stream * s);

        int updateParentInfo(char * addr, int port);
        void setParentAddr(char * addr);
        string & getParentAddr();
        void setParentPort(int port);
        int getParentPort();
        int getParentID();
        int getOrgHandle();
        Stream * connectParent();
		
    private:
        int initFE();
        int initAgent();
        Stream *initStream();
        int initBE();
        int initExtBE(int hndl);
        int getIntToken();
        int parseEnvStr(string &envStr);
        int connectBack();
        void setEnvStr(string env);
};

#define gInitializer Initializer::getInstance()

#endif

