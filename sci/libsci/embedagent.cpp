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

 Classes: EmbedAgent

 Description: embedded agent in front-end back-end
   
 Author: Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code ()

****************************************************************************/

#include "tools.hpp"

#include "embedagent.hpp"
#include "ctrlblock.hpp"
#include "writerproc.hpp"
#include "filterproc.hpp"
#include "handlerproc.hpp"
#include "routerproc.hpp"
#include "topology.hpp"
#include "message.hpp"
#include "eventntf.hpp"
#include "queue.hpp"
#include "stream.hpp"
#include "routinglist.hpp"
#include "privatedata.hpp"
#include "filterlist.hpp"


EmbedAgent::EmbedAgent()
    : handle(-1), inStream(NULL), outStream(NULL), filterInQ(NULL), filterOutQ(NULL), routerInQ(NULL), filterProc(NULL), routerProc(NULL), writerProc(NULL), routingList(NULL), filterList(NULL)
{}

EmbedAgent::~EmbedAgent()
{
    if (routerProc) {
        routerProc->release();
        delete routerProc;
    }
    if (filterProc) {
        filterProc->release();
        delete filterProc;
    }
    if (writerProc) {
        writerProc->release();
        delete writerProc;
    }
    if (routingList)
        delete routingList;
    if (filterList)
        delete filterList;
}

int EmbedAgent::init(int hndl, Stream *stream, MessageQueue *inQ, MessageQueue *outQ)
{
    handle = hndl;

    filterList = new FilterList();
    routingList = new RoutingList(hndl);
    routerProc = new RouterProcessor(hndl, routingList, filterList);
    filterInQ = new MessageQueue();
    filterInQ->setName("filterInQ");
    filterProc = new FilterProcessor(hndl, filterList);
    filterProc->setInQueue(filterInQ);
    if (outQ != NULL) {
        filterProc->setOutQueue(outQ);
    } else {
        filterOutQ = new MessageQueue();
        filterOutQ->setName("filterOutQ");
        filterProc->setOutQueue(filterOutQ);
    }
    PrivateData *pDataFilter = new PrivateData(routingList, filterList, filterProc, routerProc);
    filterProc->setSpecific(pDataFilter);
    gCtrlBlock->setUpQueue(filterInQ);
    PrivateData *pDataRouter = new PrivateData(routingList, filterList, filterProc, routerProc);
    routerProc->setSpecific(pDataRouter);
    routingList->setFilterProcessor(filterProc);

    if (stream) {
        inStream = stream;
        routerProc->setInStream(stream);
        writerProc = new WriterProcessor(hndl);
        writerProc->setName("WriterP");
        writerProc->setInQueue(filterOutQ);
        writerProc->setOutStream(stream);

        routerProc->setPeerProcessor(writerProc);
    } else if (inQ) {
        routerProc->setInQueue(inQ);
    } else {
        routerInQ = new MessageQueue();
        routerInQ->setName("routerInQ");
        routerProc->setInQueue(routerInQ);
        gCtrlBlock->setRouterInQueue(routerInQ);
        gCtrlBlock->setRouterProcessor(routerProc);
    }
    gCtrlBlock->addEmbedAgent(handle, this);

    return 0;
}

RoutingList * EmbedAgent::getRoutingList()
{
    return routingList;
}

MessageQueue * EmbedAgent::getRouterInQ()
{
    return routerInQ;
}

MessageQueue * EmbedAgent::getUpQueue()
{
    return filterOutQ;
}

int EmbedAgent::work()
{
    int rc = 0;

    routerProc->start();
    filterProc->start();
    if (writerProc)
        writerProc->start();
    if (gCtrlBlock->getMyRole() != CtrlBlock::BACK_AGENT) {
        rc = registPrivateData();
    }

    return rc;
}

int EmbedAgent::syncWait()
{
    int rc = 0;
    gNotifier->freeze(routingList->getTopology()->getInitID(), &rc);

    return rc;
}

FilterProcessor * EmbedAgent::getFilterProcessor()
{
    return filterProc;
}

extern void makeKey();
int EmbedAgent::registPrivateData()
{
    PrivateData *pDataMain = new PrivateData(routingList, filterList, filterProc, routerProc);
    int rc = pthread_once(&(Thread::once), makeKey);
    rc = pthread_setspecific(Thread::key, pDataMain);

    return rc;
}

PrivateData * EmbedAgent::genPrivateData()
{
    PrivateData *pData = new PrivateData(routingList, filterList, filterProc, routerProc);

    return pData;
}
