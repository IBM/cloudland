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

 Classes: MessageQueue

 Description: Messages manipulation.
   
 Author: Tu HongJ, Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)
   01/16/12 ronglli      Fix the issue of semaphore overflow

****************************************************************************/

#include "queue.hpp"
#include <stdlib.h>
#include <sys/time.h>
#include <time.h>
#include <errno.h>
#include <assert.h>
#ifdef __APPLE__
#include <mach/mach.h>
#endif /* __APPLE__ */

#include "exception.hpp"
#include "ctrlblock.hpp"
#include "log.hpp"

#include "atomic.hpp"
#include "message.hpp"
#include "tools.hpp"

MessageQueue::MessageQueue(bool ctl)
    : thresHold(0), flowCtl(ctl)
{
    state = true;
    ::pthread_mutex_init(&mtx, NULL);
#ifndef __APPLE__
    ::sem_init(&sem, 0, 0);
#else /* __APPLE__ */
    task = ::mach_task_self();
    ::semaphore_create(task, &sem, SYNC_POLICY_FIFO, 0);
#endif /* __APPLE__ */
}

MessageQueue::~MessageQueue()
{
    Message *msg = NULL;
    while (!queue.empty()) {
        msg = queue.front();
        queue.pop_front();
        if (decRefCount(msg->getRefCount()) == 0) {
            delete msg;
        }
    }
    queue.clear();
    
    ::pthread_mutex_destroy(&mtx);
#ifndef __APPLE__
    ::sem_destroy(&sem);
#else /* __APPLE__ */
    ::semaphore_destroy(task, sem);
#endif /* __APPLE__ */
}

int MessageQueue::flowControl(int size)
{
    long long flowctlThreshold = gCtrlBlock->getFlowctlThreshold();

    if(flowCtl) {
        if ((gCtrlBlock->getMyRole() != CtrlBlock::BACK_END) && (size > 0)) {
            while ((thresHold > flowctlThreshold) && (gCtrlBlock->getFlowctlState())) {
                SysUtil::sleep(WAIT_INTERVAL);
            }   
        }
    }

    return 0;
}

int MessageQueue::multiProduce(Message **msgs, int num)
{
    assert(msgs && (num > 0));
    int i;
    int len = 0;

    for (i = 0; i < num; i++) {
        assert(msgs[i]);
        len += msgs[i]->getContentLen();
    }
    lock();
    for (i = 0; i < num; i++) {
        queue.push_back(msgs[i]);
        
    }
    thresHold += len;
    unlock();

    release();
    flowControl(len);

    return 0;
}

void MessageQueue::release()
{
    int cnt = 0;
#ifndef __APPLE__
    while (::sem_post(&sem) != 0) {
#else /* __APPLE__ */
    while (::semaphore_signal(sem) != 0) {
#endif /* __APPLE__ */    
        if (!state)
            break;
        if (!gCtrlBlock->getFlowctlState()) {
            if (cnt > 10) {
                state = false;
                break;
            }
            cnt++;
        }
        SysUtil::sleep(WAIT_INTERVAL);
    } 
}

int MessageQueue::sem_getvalue_i()
{
    int i = -1;
#ifndef __APPLE__
    ::sem_getvalue(&sem, &i);
#endif /* __APPLE__ */
    return i;
}

void MessageQueue::produce(Message *msg)
{
    int len = 0; 

    if (!msg) {
        return;
    }
    len = msg->getContentLen();
    lock();
    queue.push_back(msg);

    thresHold += len;

    unlock();
#ifdef _SCI_DEBUG
#ifndef __APPLE__
    int val = sem_getvalue_i();
    log_debug("queue %s: produce: sem value = %ld, thresHold = %ld", name.c_str(), val, thresHold);
#endif /* __APPLE__ */
#endif
    release();
    flowControl(len);

    return;
}

void MessageQueue::insert(Message *msg)
{
    int len = 0; 

    if (!msg) {
        return;
    }
    len = msg->getContentLen();
    lock();
    queue.push_front(msg);

    thresHold += len;

    unlock();
#ifdef _SCI_DEBUG
#ifndef __APPLE__
    int val = sem_getvalue_i();
    log_debug("queue %s: produce: sem value = %ld, thresHold = %ld", name.c_str(), val, thresHold);
#endif /* __APPLE__ */
#endif
    release();
    flowControl(len);

    return;
}

int  MessageQueue::multiConsume(Message **msgs, int num)
{
    int i;
    int len = 0;

    for (i = 0; i < num; i++) {
        if (sem_wait_i(-1) != 0) {
            return -1;
        }
    }
    lock();
    for (i = 0; i < num; i++) {
        msgs[i] = queue.front();
        queue.pop_front();
        len += msgs[i]->getContentLen();
    }
    thresHold -= len;

    unlock();

    return 0;
}

Message* MessageQueue::consume(int millisecs)
{
    int len = 0;

    if (sem_wait_i(millisecs*1000) != 0) {
        return NULL;
    }

    Message *msg = NULL;

    lock();
    if (!queue.empty()) {
        msg = queue.front();
        len = msg->getContentLen();
        thresHold -= len;
    }
    unlock();

    return msg;
}

void MessageQueue::remove()
{
    Message *msg = NULL;

    lock();
    if (queue.empty()) {
        unlock();
        return;
    }

    msg = queue.front();
    queue.pop_front();
    unlock();
    if (decRefCount(msg->getRefCount()) == 0) {
        delete msg;
    }
}

int MessageQueue::getSize() 
{
    int size;

    lock();
    size = queue.size();
    unlock();

    return size;
}

bool MessageQueue::getState()
{
    return state;
}

void MessageQueue::setName(char *str)
{
    name = str;
    if (name == "filterInQ") { 
        flowCtl = true;
    }
}

string MessageQueue::getName()
{
    return name;
}

int MessageQueue::sem_wait_i(int usecs)
{
    int rc = 0;

#ifdef _SCI_DEBUG
#ifndef __APPLE__
    int val = sem_getvalue_i();
    log_debug("queue %s: sem value = %ld, thresHold = %ld", name.c_str(), val, thresHold);
#endif /* __APPLE__ */
#endif

#ifndef __APPLE__
    if (usecs < 0) {
        while (((rc = ::sem_wait(&sem)) != 0) && (errno == EINTR));
        return rc;
    } else { 
        timespec ts;
        ::clock_gettime(CLOCK_REALTIME, &ts);    // get current time
        ts.tv_nsec += (usecs % 1000000) * 1000;
        int ca = (ts.tv_nsec >= 1000000000) ? 1 : 0;
        ts.tv_nsec %= 1000000000;
        ts.tv_sec += (usecs / 1000000) + ca;
        
        while (((rc=::sem_timedwait(&sem, &ts))!=0) && (errno == EINTR));
        return rc;
    }
#else /* __APPLE__ */
    if (usecs < 0) {
        ::semaphore_wait(sem);
    } else {
        mach_timespec_t ts;
        struct timeval tv;
        ::gettimeofday(&tv, NULL);    // get current time
        ts.tv_nsec = (tv.tv_usec + (usecs % 1000000)) * 1000;
        int ca = (ts.tv_nsec >= 1000000000) ? 1 : 0;
        ts.tv_nsec %= 1000000000;
        ts.tv_sec = tv.tv_sec + (usecs / 1000000) + ca;

        ::semaphore_timedwait(sem, ts);
    }
    return 0;
#endif /* __APPLE__ */
}

void MessageQueue::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void MessageQueue::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

