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

 Classes: Thread, ThreadException

 Description: Thread manipulation.
   
 Author: Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)

****************************************************************************/

#ifndef _THREAD_HPP
#define _THREAD_HPP

#include <pthread.h>

#include <string>
#include <vector>

using namespace std;

class Thread 
{
    private:

    protected:
        int          handle;
        bool         launched;
        bool         running;
        pthread_t    thread;
        void        *data;

    public:
        static pthread_key_t key;
        static pthread_once_t once;
        Thread(int hndl = 0);
        virtual ~Thread();

        void start();
        void join();
        void detach();
        void cancel();
        virtual void run() = 0;
        
        bool isLaunched() { return launched; }
        bool getState() { return running; }
        void setState(bool state) { running = state; launched = true; }
        void setSpecific(void *data);
        void *getSpecific();
};

class ThreadException
{
    public:
        enum CODE
        {
            ERR_CREATE,
            ERR_LAUNCH,
            ERR_END,
            ERR_PRIO,
            ERR_LOCK,
            ERR_UNLOCK,
            ERR_SLEEP,
            ERR_DETACH
        };
        
    private:
        int         errCode;

    public:
        ThreadException(int code) throw();

        int getErrCode() const throw();
};

#endif

