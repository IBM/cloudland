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

 Classes: FilterList

 Description: Filter management (Note: STL does not guarantee the safety of 
              several readers & one writer cowork together, and user threads
              can query filter information at runtime, so it's necessary 
              to add a lock to protect these read & write operations).
   
 Author: Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   05/24/08 nieyy        Initial code (F156654)

****************************************************************************/

#ifndef _FILTERLIST_HPP
#define _FILTERLIST_HPP

#include <pthread.h>
#include <map>

using namespace std;

#include "general.hpp"
#include "message.hpp"

class Filter;

class FilterList
{
    public:
        typedef map<int, Filter*> FILTER_MAP;
        
    private:
        FILTER_MAP           filterInfo;
        pthread_mutex_t      mtx;
        Message              *flistMsg;

    public:
        FilterList();
        ~FilterList();

        int loadFilter(int filter_id, Filter *filter, bool invoke = true);
        int unloadFilter(int filter_id, bool invoke = true);
        
        Filter * getFilter(int filter_id);
        int numOfFilters();
        void retrieveFilterList(int *ret_val);
        Message * packMsg(sci_filter_list_t &flist);
        Message * getFlistMsg() { return flistMsg; }
        void loadFilterList(Message &msg, bool invoke = true);

    private:
        void lock();
        void unlock();
};

#endif

