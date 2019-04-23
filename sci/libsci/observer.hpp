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

 Classes: Observer

 Description: For external notification usage.
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/12/09 nieyy         Initial code (D153875)

****************************************************************************/

#ifndef _OBSERVER_HPP
#define _OBSERVER_HPP

#include <pthread.h>

class Observer 
{
    private:
        int             pipeFd[2];
        long long       count;
        bool            hasChar;

        pthread_mutex_t mtx;
        
    public:
        Observer();
        ~Observer();
        
        void notify();
        void unnotify();

        // access 
        int getPollFd();
        int getPipeWriteFd();

    private:
        void async(int fd);
        void readChar();
        void writeChar();

        void check();

        void lock();
        void unlock();
};

#endif

