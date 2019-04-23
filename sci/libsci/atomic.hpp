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

 Classes: None

 Description: Atomic operations
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   04/01/09 nieyy        Initial code (From LAPI)

****************************************************************************/

#ifndef _ATOMIC_HPP
#define _ATOMIC_HPP

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include <assert.h>
#include <pthread.h>

#if defined(_SCI_LINUX) // Linux

/**********************************************************************
 *
 *  Atomic operations
 *
 **********************************************************************/
typedef int          *atomic_p;
typedef long long   *atomic_l;
typedef int          boolean_t;
typedef unsigned int uint;

#ifdef POWER_ARCH

/*
  For Power architecture, isync is only necessary when entering a 
  critical section to discard any instruction prefetch and possible
  execution on stale data. 

  It's handy to put it in _check_lock but not other routines.
*/

static __inline__ 
int fetch_and_add(atomic_p dest, int val)
{
    int old, sum;
    __asm__ __volatile__(
            "1: lwarx   %[old], 0, %[dest]      \n\t"
            "   add     %[sum], %[old], %[val]  \n\t"
            "   stwcx.  %[sum], 0, %[dest]      \n\t"
            "   bne-    1b                      \n\t"
            : [sum] "=&r" (sum), [old] "=&r" (old)
            : [val] "r" (val), [dest] "r" (dest)
            : "%0", "cc", "memory");
    return old;
}

#endif /* POWER_ARCH */

#ifdef INTEL_ARCH
/*
 Note: Inlining cmpxchg2 doesn't generate
       correct code!
 */
static //__inline__ 
boolean_t cmpxchg2(atomic_p dest, int comp, int exch)
{
    unsigned int old;
    __asm__ __volatile__(
        "lock; cmpxchgl %[exch], %[dest]"
        : [dest] "=m" (*dest), "=a" (old)
        : [exch] "r" (exch), "m"  (*dest), "a"  (comp)
        : "memory" );
    return (old == comp);
}

static __inline__ 
int fetch_and_add(atomic_p ptr, int val)
{
    int prev;
    do prev = *ptr;
    while (!cmpxchg2(ptr, prev, (prev+val)));
    return prev;
}

#endif /* INTEL_ARCH */

#elif defined(__APPLE__)
#include <libkern/OSAtomic.h>
typedef int *atomic_p;

/*
 * We need to implement fetch_and_add using a
 * splinlock as the OSAtomicAdd functions return
 * the new value, not the old value.
 */
static __inline__
int fetch_and_add(atomic_p ptr, int val) {
    int old_val;
    OSSpinLock lock = 0;
    OSSpinLockLock(&lock);
    old_val = *ptr;
    *ptr += val;
    OSSpinLockUnlock(&lock);
    return old_val;
}

#else // AIX

#include <sys/atomic_op.h>

#endif

static int decRefCount(int &refCount)
{
    int count = fetch_and_add(&refCount, -1);
    return (count - 1);
}

static int incRefCount(int &refCount)
{
    int count = fetch_and_add(&refCount, 1);
    return (count + 1);
}

#endif

