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

 Classes: Allocator

 Description: Allocate global resources.
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   03/03/09 nieyy         Initial code (D153875)

****************************************************************************/

#include "allocator.hpp"
#include <assert.h>

#include "atomic.hpp"
#include "topology.hpp"
#include "ctrlblock.hpp"

Allocator * Allocator::instance = NULL;

Allocator::Allocator()
{
}

Allocator::~Allocator()
{
}

void Allocator::reset()
{
    nextGroupID = SCI_GROUP_ALL - 1;
    nextBEID = gCtrlBlock->getTopology()->getBENum();
}

Allocator * Allocator::getInstance()
{
    if (instance == NULL)
        instance = new Allocator();
    return instance;
}

void Allocator::allocateGroup(sci_group_t * group)
{
    assert(group);
    *group = fetch_and_add(&nextGroupID, -1);
}

void Allocator::allocateBE(int * be_id)
{
    assert(be_id);
    *be_id = fetch_and_add(&nextBEID, 1);
}

