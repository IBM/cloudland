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

#ifndef _ALLOCATOR_HPP
#define _ALLOCATOR_HPP

#include "sci.h"

class Allocator 
{
    private:
        int             nextGroupID;
        int             nextBEID;
        
        Allocator();
        static Allocator *instance;
        
    public:
        ~Allocator();
        static Allocator *getInstance();

        void reset();

        void allocateGroup(sci_group_t *group);
        void allocateBE(int *be_id);
};

#define gAllocator Allocator::getInstance()

#endif

