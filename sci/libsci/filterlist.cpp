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
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   05/24/08 nieyy        Initial code (F156654)

****************************************************************************/

#include "filterlist.hpp"
#include <assert.h>

#include "filter.hpp"
#include "packer.hpp"


FilterList::FilterList()
    : flistMsg(NULL)
{
    filterInfo.clear();

    ::pthread_mutex_init(&mtx, NULL);
}

FilterList::~FilterList()
{
    // delete all loaded filters
    FILTER_MAP::iterator fit = filterInfo.begin();
    for (; fit != filterInfo.end(); fit++) {
        fit->second->unload();
        delete fit->second;
    }
    filterInfo.clear();

    ::pthread_mutex_destroy(&mtx);
}

int FilterList::loadFilter(int filter_id, Filter * filter, bool invoke)
{
    int rc = SCI_SUCCESS;
    if (invoke) {
        // call init func
        rc = filter->load();
    }
    
    if (rc == SCI_SUCCESS) {
        lock();
        filterInfo[filter_id] = filter;
        unlock();
    }

    return rc;
}

Filter * FilterList::getFilter(int filter_id)
{
    Filter *filter = NULL;

    lock();
    FILTER_MAP::iterator fit = filterInfo.find(filter_id);
    if (fit != filterInfo.end()) {
        filter = (*fit).second;
    }
    unlock();

    return filter;
}

Message * FilterList::packMsg(sci_filter_list_t &flist)
{
    int i = 0;
    char *bufs[1];
    int sizes[1];
    Packer packer;

    if (flist.num == 0)
        return NULL;

    packer.packInt(flist.num);
    for (i = 0; i < flist.num; i++) {
        packer.packInt(flist.filters[i].filter_id);
        packer.packStr(flist.filters[i].so_file);
    }
    
    bufs[0] = packer.getPackedMsg();
    sizes[0] = packer.getPackedMsgLen();

    Message *msg = new Message();
    msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes, Message::FILTER_LIST);

    return msg;
}

void FilterList::loadFilterList(Message &msg, bool invoke) 
{
    int i = 0;
    Packer packer(msg.getContentBuf());
    int num = packer.unpackInt();
    sci_filter_info_t finfo;
    Filter *filter = NULL;
    char *bufs[1];
    int sizes[1];

    for (i = 0; i < num; i++) {
        finfo.filter_id = packer.unpackInt();
        finfo.so_file = packer.unpackStr();
        filter = new Filter(finfo);
        loadFilter(finfo.filter_id, filter, invoke);
    }

    bufs[0] = msg.getContentBuf();
    sizes[0] = msg.getContentLen();
    flistMsg = new Message();
    flistMsg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes, msg.getType());
}

int FilterList::unloadFilter(int filter_id, bool invoke)
{
    Filter *filter = NULL;

    lock();
    FILTER_MAP::iterator fit = filterInfo.find(filter_id);
    if (fit != filterInfo.end()) {
        filter = (*fit).second;
    } else {
        unlock();
        return SCI_ERR_FILTER_NOTFOUND;
    }

    filterInfo.erase(filter_id);
    unlock();
    
    int rc = SCI_SUCCESS;
    if (invoke) {
        // call term_func
        rc = filter->unload();
    }
    delete filter;
    
    return rc;
}

int FilterList::numOfFilters()
{
    int size;

    lock();
    size = filterInfo.size();
    unlock();

    return size;
}

void FilterList::retrieveFilterList(int * ret_val)
{
    int i = 0;

    lock();
    FILTER_MAP::iterator it = filterInfo.begin();
    for (; it!=filterInfo.end(); ++it) {
        ret_val[i++] = (*it).first;
    }
    unlock();
}

void FilterList::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void FilterList::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

