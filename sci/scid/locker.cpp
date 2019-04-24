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

 Description: Lock Operations.
   
 Author: Tu HongJ, Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/12/09 tuhongj      Initial code (D155101)

****************************************************************************/

#include "locker.hpp"
#include <stdio.h>
#include <assert.h>
#include <sched.h>

Locker * Locker::locker = NULL;

Locker::Locker()
{
    ::pthread_mutex_init(&gMutex, NULL);

    ::pthread_mutex_init(&cMutex, NULL);
    ::pthread_cond_init(&cond, NULL);

    freezed = false;
}

Locker::~Locker()
{
    ::pthread_mutex_destroy(&gMutex);

    ::pthread_cond_destroy(&cond);
    ::pthread_mutex_destroy(&cMutex);
}

Locker * Locker::getLocker()
{
    if (locker == NULL) {
        locker = new Locker();
    }
    
    return locker;
}

void Locker::lock()
{
    ::pthread_mutex_lock(&gMutex);
}

void Locker::unlock()
{
    ::pthread_mutex_unlock(&gMutex);
}

/*

Warning: do not try to use freeze() & notify() function in a same thread, the style is like:
    thread a: freeze()
    thread b: notify()
 The freeze & notification times are unlimited.

*/

void Locker::freeze()
{   
    ::pthread_mutex_lock(&cMutex);
    freezed= true;
    while (freezed) {
        ::pthread_cond_wait(&cond, &cMutex);
    }
    ::pthread_mutex_unlock(&cMutex);
}

void Locker::notify()
{
    ::pthread_mutex_lock(&cMutex);
    freezed= false;
    ::pthread_cond_broadcast(&cond);
    ::pthread_mutex_unlock(&cMutex);
}

