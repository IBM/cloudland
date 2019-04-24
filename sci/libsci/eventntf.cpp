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

 Classes: EventNotify

 Description: Synchronization between threads
   
 Author: Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   12/05/08 tuhongj      Initial code (D154660)

****************************************************************************/

#include "eventntf.hpp"
#include <unistd.h>
#include <string.h>
#include <assert.h>
#include <sys/time.h>

#include "tools.hpp"
#include "ctrlblock.hpp"
#include "log.hpp"

const struct serialNtfTest INIT_VAL_NTF = {0, 0, 0, 0};

EventNotify * EventNotify::notifier = NULL;

EventNotify::EventNotify()
    : serialNum(0), serialSize(0)
{
    ::pthread_mutex_init(&mtx, NULL);
    ::pthread_cond_init(&cond, NULL);
    serialTest.assign(MAX_SERIAL_NUM, INIT_VAL_NTF);
}

EventNotify::~EventNotify()
{
    ::pthread_cond_destroy(&cond);
    ::pthread_mutex_destroy(&mtx);

    notifier = NULL;
}

int EventNotify::allocate()
{
    int num;

    lock();
    do {
        if (serialSize >= serialTest.size()) {
            log_debug("EventNotify: resize the serialTest, from original size %d, to new size %d",
                    serialTest.size(), serialTest.size()*2);
            serialTest.resize(serialTest.size() * 2, INIT_VAL_NTF);
        }
        serialNum = (serialNum + 1) % serialTest.size();
    } while (serialTest[serialNum].used == true);
    num = serialNum;
    serialTest[serialNum].used = true;
    serialTest[serialNum].notified = false; 
    serialTest[serialNum].freezed = false;
    serialSize++;
    unlock();

    return num;
}

void EventNotify::freeze(int id, void *ret_val)
{
    lock();
    serialTest[id].ret = ret_val;
    serialTest[id].notified = false;
    serialTest[id].freezed = true;
    while (serialTest[id].notified == false) {
        ::pthread_cond_wait(&cond, &mtx);
    }
    serialTest[id].freezed = false;
    serialTest[id].used = false;
    serialSize--;
    unlock();
}

void EventNotify::notify(int id)
{
    test(id);
    lock();
    serialTest[id].used = false;
    serialTest[id].notified = true;
    ::pthread_cond_broadcast(&cond); 
    unlock();
}

timespec SetTime(int usecs) {
    struct timeval now;
    struct timespec to;

    gettimeofday(&now, NULL);
    to.tv_sec = now.tv_sec + usecs / 1000000;
    to.tv_nsec = now.tv_usec * 1000 + (usecs % 1000000) * 1000;

    return to;
}

int EventNotify::freeze_i(int id, void *ret_val, int usecs)
{
    struct timespec to;
    bool tmpNotified;
    bool tmpFreezed;
    int rc = -1;
    int count = 0;

    lock();
    tmpNotified = serialTest[id].notified;
    tmpFreezed = serialTest[id].freezed;

    serialTest[id].ret = ret_val;
    serialTest[id].notified = false;
    serialTest[id].freezed = true;
    while ((serialTest[id].notified == false) && (!gCtrlBlock->getTermState())) { 
        to = SetTime(usecs);
        ::pthread_cond_timedwait(&cond, &mtx, &to);
        count++;
    }
    if (serialTest[id].notified == true) { // The notify is set to true correctly
        serialTest[id].freezed = false;
        serialTest[id].used = false;
        serialSize--;
        rc = 0;
    } else { // The freeze is not set correctly. Restore the original value
        serialTest[id].notified = tmpNotified;
        serialTest[id].freezed = tmpFreezed;
        rc = -1;
    }
    unlock();

    return rc;
}

int EventNotify::notify_i(int id, int usecs)
{
    if (!test_i(id)) {
        return -1;
    }
    lock();
    serialTest[id].used = false;
    serialTest[id].notified = true;
    ::pthread_cond_broadcast(&cond); 
    unlock();

    return 0;
}

void * EventNotify::getRetVal(int id)
{
    test(id);
    return serialTest[id].ret;
}

bool EventNotify::getState(int id)
{
    bool state;

    assert((id >= 0) && (id < serialTest.size()));
    lock();
    state = serialTest[id].used;
    unlock();

    return state;
}

bool EventNotify::test(int id)
{
    assert((id >= 0) && (id < serialTest.size()));
    while (serialTest[id].freezed == false) {
        /* Almost impossible running into here */
        SysUtil::sleep(WAIT_INTERVAL);
    }
    assert(serialTest[id].used = true);
    
    return true;
}

bool EventNotify::test_i(int id)
{
    assert((id >= 0) && (id < serialTest.size()));
    while (serialTest[id].freezed == false) {
        if (gCtrlBlock->getTermState())
            return false;
        /* Almost impossible running into here */
        SysUtil::sleep(WAIT_INTERVAL);
    }
    assert(serialTest[id].used = true);
    
    return true;
}

void EventNotify::tryFreeze()
{
    lock();
    while(serialTest[serialNum].freezed == true) {
        ::pthread_cond_wait(&cond, &mtx);
    }
    unlock();
}

void EventNotify::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void EventNotify::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

