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

 Classes: FilterProcessor

 Description: Properties of class 'FilterProcessor':
    input: a. a message queue
    output: a. a stream
            b. a message queue
    action: use user-defined filter handlers to process the messages
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#include "filterproc.hpp"
#include <assert.h>

#include "log.hpp"
#include "exception.hpp"
#include "socket.hpp"

#include "atomic.hpp"
#include "ctrlblock.hpp"
#include "message.hpp"
#include "stream.hpp"
#include "filter.hpp"
#include "filterlist.hpp"
#include "queue.hpp"
#include "eventntf.hpp"
#include "observer.hpp"

FilterProcessor::FilterProcessor(int hndl, FilterList *flist)
    : Processor(hndl), filterList(flist), filtered(false), curFilterID(SCI_FILTER_NULL)
{
    name = "UpstreamFilter";

    inQueue = NULL;
    outQueue = NULL;

    observer = NULL;
}

FilterProcessor::~FilterProcessor()
{
    delete inQueue;
}

Message * FilterProcessor::read()
{
    assert(inQueue);

    Message *msg = NULL;

    filtered = false;
    msg = inQueue->consume();
    
    return msg;
}

void FilterProcessor::process(Message * msg)
{
    int id = msg->getFilterID();
    if (id != SCI_FILTER_NULL) {
        Filter *filter = filterList->getFilter(id);
        // call user's filter handler
        if (filter != NULL) {
            curFilterID = id;
            
            filtered = true;
            filter->input(msg->getGroup(), msg->getContentBuf(), msg->getContentLen());
        }
    }
}

void FilterProcessor::write(Message * msg)
{
    assert(outQueue);

    if (filtered) {
        inQueue->remove();
        return;
    }

    if (observer) {
        observer->notify();
    }

    incRefCount(msg->getRefCount());
    outQueue->produce(msg);

    inQueue->remove();
}

void FilterProcessor::seize()
{
    setState(false);
}

int FilterProcessor::recover()
{
    // TODO
    return -1;
}

void FilterProcessor::clean()
{
}

void FilterProcessor::deliever(Message * msg)
{
    if (observer) {
        observer->notify();
    }
    outQueue->produce(msg);
}

int FilterProcessor::getCurFilterID()
{
    return curFilterID;
}

void FilterProcessor::setInQueue(MessageQueue * queue)
{
    inQueue = queue;
}

void FilterProcessor::setOutQueue(MessageQueue * queue)
{
    outQueue = queue;
}

void FilterProcessor::setObserver(Observer * ob)
{
    observer = ob;
}

MessageQueue * FilterProcessor::getInQueue()
{
    return inQueue;
}

MessageQueue * FilterProcessor::getOutQueue()
{
    return outQueue;
}
