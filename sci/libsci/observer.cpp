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

#include "observer.hpp"
#include <assert.h>
#include <errno.h>
#include <unistd.h>
#include <stddef.h>
#include <fcntl.h>
#include <stdio.h>

#include "exception.hpp"

Observer::Observer()
{
    pipeFd[0] = -1;
    pipeFd[1] = -1;

    count = 0;
    hasChar = false;

    int rc = ::pipe(pipeFd);
    if (rc != 0) {
        throw Exception(Exception::SYS_CALL);
    }
    async(pipeFd[0]);
    async(pipeFd[1]);

    ::pthread_mutex_init(&mtx, NULL);
}

Observer::~Observer()
{
    ::close(pipeFd[0]);
    ::close(pipeFd[1]);

    ::pthread_mutex_destroy(&mtx);
}

void Observer::notify()
{
    lock();
    try {
        count++;
        check();
    } catch (...) {
        unlock();
        throw;
    }
    unlock();
}

void Observer::unnotify()
{
    lock();
    try {
        if (hasChar) {
            readChar();
            hasChar = false;
        }
        check();
    } catch (...) {
        unlock();
        throw;
    }
    unlock();
}

int Observer::getPollFd()
{
    return pipeFd[0];
}

int Observer::getPipeWriteFd()
{
    return pipeFd[1];
}

void Observer::async(int fd)
{
    int flags, newflags;

    flags = ::fcntl(fd, F_GETFL);
    if (flags < 0)
        throw Exception(Exception::SYS_CALL);

    newflags = flags & ~O_NONBLOCK;

    if (newflags != flags) {
        if (::fcntl(fd, F_SETFL, newflags) < 0) {
            throw Exception(Exception::SYS_CALL);
        }
    }
}

void Observer::readChar()
{
    // read a char signal from the socket
    char signal;
    while (true) {
        int bytes = ::read(pipeFd[0], &signal, sizeof(char));
        if (bytes < 0) {
            if (errno == EINTR) {
                continue;
            }
            if ((errno == EAGAIN) || (errno == EWOULDBLOCK)) {
                break;
            }
            throw Exception(Exception::SYS_CALL);
        }
        break;
    }
}

void Observer::writeChar()
{
    // write a char signal to the socket
    char signal = 'a';
    while (true) {
        int bytes = ::write(pipeFd[1], &signal, sizeof(char));
        if (bytes < 0) {
            if ((errno == EAGAIN) || (errno == EWOULDBLOCK) || (errno == EINTR)) {
                continue;
            }
            throw Exception(Exception::SYS_CALL);
        }
        break;
    }
}

void Observer::check()
{
    if (!hasChar) {
        if (count > 0) {
            hasChar = true;
            writeChar();
            count--;
        }
    }
}

void Observer::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void Observer::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

