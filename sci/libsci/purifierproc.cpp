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

 Classes: PurifierProcessor

 Description: Properties of class 'PurifierProcessor':
    input: a. a stream 
    output: a. a message queue
    action: purify message, discarded useless messages
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/25/09 nieyy      Initial code (D153875)
   01/16/12 ronglli    Add codes to detect SOCKET_BROKEN

****************************************************************************/

#include "purifierproc.hpp"
#include <assert.h>

#include "log.hpp"
#include "exception.hpp"
#include "socket.hpp"

#include "atomic.hpp"
#include "ctrlblock.hpp"
#include "routinglist.hpp"
#include "message.hpp"
#include "stream.hpp"
#include "privatedata.hpp"
#include "queue.hpp"
#include "observer.hpp"
#include "filter.hpp"
#include "filterlist.hpp"
#include "writerproc.hpp"
#include "initializer.hpp"
#include "eventntf.hpp"
#include "tools.hpp"
#include "sshfunc.hpp"

PurifierProcessor::PurifierProcessor(int hndl) 
    : Processor(hndl), inStream(NULL), outErrorQueue(NULL), peerProcessor(NULL), observer(NULL), joinSegs(false)
{
    name = "Purifier";
    hndlr = gCtrlBlock->getEndInfo()->be_info.hndlr;
    param = gCtrlBlock->getEndInfo()->be_info.param;
    routingList = new RoutingList(hndl);
    routingList->addBE(SCI_GROUP_ALL, VALIDBACKENDIDS, hndl);
    filterList = new FilterList();
    PrivateData *pData = new PrivateData(routingList, filterList, NULL);
    setSpecific(pData);
}

PurifierProcessor::~PurifierProcessor()
{
    if (inQueue)
        delete inQueue;
    if (routingList)
        delete routingList;
    if (filterList)
        delete filterList;
}

RoutingList * PurifierProcessor::getRoutingList()
{
    return routingList;
}

FilterList * PurifierProcessor::getFilterList()
{
    return filterList;
}

Message * PurifierProcessor::read()
{
    Message *msg = NULL;
    assert(inStream || inQueue);

    if (inStream != NULL) {
        msg = new Message();
        *inStream >> *msg;
    } else {
        msg = inQueue->consume();
    }

    if (msg && (msg->getType() == Message::SEGMENT)) {
        joinSegs = true;
        msg = Message::joinSegments(msg, inStream, inQueue);
    }

    return msg;
}

void PurifierProcessor::process(Message * msg)
{
    Filter *filter = NULL;
    switch(msg->getType()) {
        case Message::SEGMENT:
        case Message::COMMAND:
            if (observer) {
                observer->notify();
                incRefCount(msg->getRefCount()); // inQueue and outQueue
                outQueue->produce(msg);
            } else {
                hndlr(param, msg->getGroup(), msg->getContentBuf(), msg->getContentLen());
            }
            break;
        case Message::UNCLE:
        case Message::UNCLE_LIST:
        case Message::PARENT:
        case Message::ERROR_EVENT:
        case Message::SHUTDOWN:
        case Message::KILLNODE:
            isError = true;
            msg->setID(handle);
            break;
        case Message::GROUP_CREATE:
        case Message::GROUP_OPERATE:
        case Message::GROUP_OPERATE_EXT:
            routingList->addBE(msg->getGroup(), VALIDBACKENDIDS, gCtrlBlock->getMyHandle());
            break;
        case Message::GROUP_FREE:
            routingList->removeGroup(msg->getGroup());
            break;
        case Message::FILTER_LOAD:
            filter = new Filter();
            filter->unpackMsg(*msg);
            filterList->loadFilter(filter->getId(), filter, false);
            break;
        case Message::FILTER_UNLOAD:
            filterList->unloadFilter(msg->getFilterID(), false);
            break;
        case Message::FILTER_LIST:
            filterList->loadFilterList(*msg, false);
            break;
        case Message::BE_REMOVE:
        case Message::QUIT:
            gCtrlBlock->setTermState(true);
            gCtrlBlock->setRecoverMode(0);
            setState(false);
            break;
        default:
            break;
    }
}

void PurifierProcessor::write(Message * msg)
{
    if (joinSegs || inStream) {
        joinSegs = false;
        if (decRefCount(msg->getRefCount()) == 0)
            delete msg;
        return;
    }
    inQueue->remove();
}

int PurifierProcessor::recover()
{
    int rc = -1;

    if ((gCtrlBlock->getTermState()) || (!gCtrlBlock->getRecoverMode())) {
        return rc;
    }

    log_debug("Purifier: begin to do the recover.");
    if (gCtrlBlock->getParentInfoWaitState()) {
        while (gInitializer->pInfoUpdated == false) {
            if (gCtrlBlock->getTermState()) {
                log_debug("Purifier: incorrect state");
                return rc;
            }
            SysUtil::sleep(WAIT_INTERVAL);
        }
    }

    log_debug("Purifier: begin to do the reconnect...");
    try {
        struct iovec sign = {0};
        int hndl = gInitializer->getOrgHandle();
        int pID = gInitializer->getParentID();
        string pAddr = gInitializer->getParentAddr();
        int pPort = gInitializer->getParentPort();

        inStream->stopRead();

        WriterProcessor * writer = getPeerProcessor();
        while(!(writer->isLaunched())) {
            SysUtil::sleep(WAIT_INTERVAL);
        }
        if (!writer->getRecoverState()) {
            Message *msg = new Message(); 
            // The writer thread may be in consume, which will not enter into recover. Need to send a notification msg to it
            msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 0, NULL, NULL, Message::RELEASE);

            log_debug("Purifier: begin to set the writer release state to false, and produce rel msg to writer");
            writer->setReleaseState(true);
            writer->getInQueue()->produce(msg);
        }
        while(!(writer->getRecoverState())) {
            SysUtil::sleep(WAIT_INTERVAL);
        }

        inStream = gInitializer->connectParent();
        writer->setOutStream(inStream);

        if (gCtrlBlock->getParentInfoWaitState()) {           
            gInitializer->pInfoUpdated = false; 
            gCtrlBlock->setParentInfoWaitState(false); 
            gNotifier->notify(gInitializer->notifyID);
        }
        rc = 0;
    } catch (SocketException &e) {
        rc = -1;
        log_error("Purifier: recover exception: socket exception: %s", e.getErrMsg().c_str());
        SysUtil::sleep(WAIT_INTERVAL);
    }

    return rc;
}

void PurifierProcessor::seize()
{
    setState(false);
}

void PurifierProcessor::clean()
{
    if (inStream)
        inStream->stopRead();
    if (observer) {
        try {
            gCtrlBlock->releasePollQueue();
        } catch (std::bad_alloc) {
            log_error("Processor Purifier: out of memory");
            // To do; add correct error handling
        }
    }
    gCtrlBlock->setFlowctlState(false);

    gCtrlBlock->disable();
    if (peerProcessor) {
        peerProcessor->release();
        delete peerProcessor;
    }
}

void PurifierProcessor::setInStream(Stream * stream)
{
    inStream = stream;
}

Stream * PurifierProcessor::getInStream()
{
    return inStream;
}

void PurifierProcessor::setInQueue(MessageQueue * queue)
{
    inQueue = queue;
}

MessageQueue * PurifierProcessor::getInQueue()
{
    return inQueue;
}

void PurifierProcessor::setOutQueue(MessageQueue * queue)
{
    outQueue = queue;
}

void PurifierProcessor::setOutErrorQueue(MessageQueue * queue)
{
    outErrorQueue = queue;
}

void PurifierProcessor::setPeerProcessor(WriterProcessor * processor)
{
    peerProcessor =  processor;
}

WriterProcessor * PurifierProcessor::getPeerProcessor()
{
    return peerProcessor;
}

void PurifierProcessor::setObserver(Observer * ob)
{
    observer = ob;
}
