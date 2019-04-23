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

 Classes: Locker

 Description: Global lock operations.
   
 Author: Tu HongJ, Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/12/09 tuhongj      Initial code (D155101)

****************************************************************************/

#ifndef _LOCKER_HPP
#define _LOCKER_HPP

#include <pthread.h>

class Locker 
{
    private:
        static Locker         *locker;
        pthread_mutex_t       gMutex;

        pthread_mutex_t       cMutex;
        pthread_cond_t        cond;

        bool                  freezed;

        Locker();
        
    public:
        ~Locker();
        static Locker * getLocker();
        
        void lock();
        void unlock();

        void freeze();
        void notify();
};

#endif

