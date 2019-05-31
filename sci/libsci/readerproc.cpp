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

 Classes: ReaderProcessor

 Description: Properties of class 'ReaderProcessor':
    input: a stream
    output: two message queues
    action: relay messages from the stream to the queues, normal messages to a
            queue, error handling messages to another queue
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   05/25/09 nieyy      Initial code (F156654)
   01/16/12 ronglli    Add codes to detect SOCKET_BROKEN

****************************************************************************/

#include "readerproc.hpp"
#include <assert.h>

#include "log.hpp"
#include "exception.hpp"
#include "socket.hpp"

#include "ctrlblock.hpp"
#include "routinglist.hpp"
#include "message.hpp"
#include "stream.hpp"
#include "queue.hpp"
#include "writerproc.hpp"

#include "eventntf.hpp"
#include "tools.hpp"

ReaderProcessor::ReaderProcessor(int hndl) 
    : Processor(hndl), recoverID(-1), notifyID(-1), peerProcessor(NULL) 
{
    name = "Reader";

    inStream = NULL;
    outQueue = NULL;

    outErrorQueue = NULL;
}

ReaderProcessor::~ReaderProcessor()
{
}

Message * ReaderProcessor::read()
{
    assert(inStream);
    Message *msg = NULL;

    msg = new Message();
    *inStream >> *msg;

    return msg;
}

void ReaderProcessor::process(Message * msg)
{
    assert(msg);
    // no action
}

void ReaderProcessor::write(Message * msg)
{
    assert(outQueue);

    // normal and error messages to different queues
    switch (msg->getType()) {
        case Message::GROUP_MERGE:
        case Message::ERROR_EVENT:
            // use 'id' field to store child agent id information, and transfer this message
            // to router processor
            msg->setID(handle);
        case Message::UNCLE:
        case Message::UNCLE_LIST:
        case Message::PARENT:
        case Message::SHUTDOWN:
        case Message::KILLNODE:
            if (outErrorQueue) {
                outErrorQueue->produce(msg);
            } else {
                delete msg;
            }
            break;
        case Message::SOCKET_BROKEN:
        case Message::ERROR_DATA:
        case Message::ERROR_THREAD:
            {
                gCtrlBlock->notifyChildHealthState(msg);
            }
            break;
        default:
            outQueue->produce(msg);
            break;
    }
}

void ReaderProcessor::seize()
{    
    // exit the peer relay processor thread related to the same socket
    setState(false);  
    if (!gCtrlBlock->getTermState()) {
        gCtrlBlock->notifyChildHealthState(handle, hState); 
    }
}

void ReaderProcessor::releasePeer(WriterProcessor * writer)
{
    while(!(writer->isLaunched())) {
        SysUtil::sleep(WAIT_INTERVAL);
    }
    writer->setReleaseState(true);
    Message *msg = new Message(); 
    msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 0, NULL, NULL, Message::RELEASE);
    writer->getInQueue()->produce(msg);
}

int ReaderProcessor::recover()
{    
    // exit the peer relay processor thread related to the same socket
    if ((gCtrlBlock->getTermState()) || (!gCtrlBlock->getRecoverMode())) {
        return -1;
    }
    inStream->stopRead(); 

    WriterProcessor * writer = getPeerProcessor();
    if (!(writer->getRecoverState())) {
        releasePeer(writer);
    }
    while (!(writer->getRecoverState())) {
        SysUtil::sleep(WAIT_INTERVAL);
    }
    gCtrlBlock->setRecover(handle);

    if (recoverID == -1) {
        recoverID = gNotifier->allocate();
    }

    Stream *st;
    if (gNotifier->freeze_i(recoverID, &st) != 0) {
        log_debug("reader%d: recover error: freeze_i failed for the stream %p, recoverID = %d", handle, st, recoverID);
        return -1;
    }
    log_debug("reader%d: finish freeze for the stream %p, recoverID = %d", handle, st, recoverID);
    recoverID = gNotifier->allocate();
    log_debug("reader%d: begin to notify notifyID %d", handle, notifyID);
    if (gNotifier->notify_i(notifyID) != 0) {
        log_debug("reader%d: recover error: notify_i failed for the stream %p, recoverID = %d", handle, st, recoverID);
        return -1;
    }
    
    inStream = st;
    return 0; 
}

void ReaderProcessor::clean()
{
    inStream->stopRead();
    setState(false);    
}

void ReaderProcessor::setInStream(Stream * stream)
{
    if (inStream == NULL) {
        log_debug("reader%d: begin to set the stream. Original is NULL", handle);
        inStream = stream;
    } else {
        log_debug("reader%d: begin to notify the stream %p, recoverID = %d", handle, stream, recoverID);
        while (recoverID == -1) {
            SysUtil::sleep(WAIT_INTERVAL);
        }
        if (notifyID == -1) {
            notifyID = gNotifier->allocate();
        }
        *(Stream **)gNotifier->getRetVal(recoverID) = stream;
        gNotifier->notify(recoverID);
        log_debug("reader%d: finish notify the recoverID %d", handle, recoverID);
        log_debug("reader%d: begin to freeze the notifyID %d", handle, notifyID);
        gNotifier->freeze(notifyID, NULL);
        log_debug("reader%d: finish freeze the notifyID %d", handle, notifyID);
        notifyID = gNotifier->allocate();
    }
    gCtrlBlock->clearRecover(handle);
}

void ReaderProcessor::setOutQueue(MessageQueue * queue)
{
    outQueue = queue;
}

void ReaderProcessor::setOutErrorQueue(MessageQueue * queue)
{
    outErrorQueue = queue;
}

void ReaderProcessor::setPeerProcessor(WriterProcessor* processor)
{
    peerProcessor =  processor;
}

WriterProcessor *ReaderProcessor::getPeerProcessor()
{
    return peerProcessor;
}

