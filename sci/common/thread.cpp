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

#include <assert.h>
#include <string.h>
#include <signal.h>

#include "thread.hpp"

using namespace std;

pthread_key_t Thread::key = 0;
pthread_once_t Thread::once = PTHREAD_ONCE_INIT;

void makeKey()
{
    pthread_key_create(&(Thread::key), NULL);
}

void* init(void * pthis)
{
    Thread *p = (Thread *) pthis;
    void *data = p->getSpecific();
    pthread_setcancelstate(PTHREAD_CANCEL_ENABLE, NULL);
    pthread_setcanceltype(PTHREAD_CANCEL_ASYNCHRONOUS, NULL);
    int rc = pthread_once(&(Thread::once), makeKey);
    if (data != NULL) {
        rc = pthread_setspecific(Thread::key, data);
    }
    p->setState(true);
    p->run();

    return 0;
}

Thread::Thread(int hndl)
    : handle(hndl), launched(false), running(false), data(NULL)
{
}

Thread::~Thread() 
{
}

void Thread::start()
{
    if (!launched) {
        sigset_t sigs_to_block;
        sigset_t old_sigs;
        sigfillset(&sigs_to_block);
        pthread_sigmask(SIG_SETMASK, &sigs_to_block, &old_sigs);
        
        if (pthread_create(&(thread), NULL, init, this) != 0) {
            running = false;
            pthread_sigmask(SIG_SETMASK, &old_sigs, NULL);
            throw ThreadException(ThreadException::ERR_CREATE);
        }
        pthread_sigmask(SIG_SETMASK, &old_sigs, NULL);
    } else {
        throw ThreadException(ThreadException::ERR_LAUNCH);
    }
}

void Thread::join()
{
    if (!launched)
        return;
    
    pthread_join(thread, NULL);
    running = false;
}

void Thread::detach()
{
    if (launched) {
        pthread_detach(thread);
    } else {
        throw ThreadException(ThreadException::ERR_DETACH);
    }
}

void Thread::cancel()
{
    pthread_cancel(thread);
}

void Thread::setSpecific(void *dat)
{
    data = dat;
}

void * Thread::getSpecific()
{
    return data;
}

ThreadException::ThreadException(int code) throw()
    : errCode(code)
{
}

int ThreadException::getErrCode() const throw()
{
    return errCode;
}

