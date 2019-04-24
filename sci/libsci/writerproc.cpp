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

 Classes: WriterProcessor

 Description: Properties of class 'WriterProcessor':
    input: a message queue
    output: a stream
    action: relay messages from the queue to the stream.
   
 Author: Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   05/25/09 nieyy      Initial code (F156654)
   01/16/12 ronglli    Add codes to detect SOCKET_BROKEN

****************************************************************************/

#include "writerproc.hpp"
#include <assert.h>

#include "log.hpp"
#include "exception.hpp"
#include "socket.hpp"

#include "ctrlblock.hpp"
#include "tools.hpp"
#include "message.hpp"
#include "stream.hpp"
#include "queue.hpp"
#include "readerproc.hpp"

#include "eventntf.hpp"
#include "tools.hpp"

WriterProcessor::WriterProcessor(int hndl) 
    : Processor(hndl), peerProcessor(NULL), recoverID(-1), notifyID(-1), recoverState(false), releaseState(false)
{
    name = "Writer";

    inQueue = NULL;
    outStream = NULL;
}

WriterProcessor::~WriterProcessor()
{
    if (outStream)
        delete outStream;
    if (inQueue)
        delete inQueue;
}

Message * WriterProcessor::read()
{
    assert(inQueue);

    Message *msg = NULL;

    msg = inQueue->consume();

    return msg;
}

void WriterProcessor::process(Message * msg)
{
    // no action
}

void WriterProcessor::write(Message * msg)
{
    assert(outStream);

    switch (msg->getType()) {
        case Message::RELEASE:
            inQueue->remove();
            if (getReleaseState()) {
                throw (SocketException(SocketException::NET_ERR_CLOSED));
            }
            break;
        default:
            try {    
                *outStream << *msg;
            } catch (SocketException &e) {
                inQueue->release();
                throw;
            } catch (...) {
                inQueue->release();
                throw;
            }
            inQueue->remove();
            break;
    }
}

void WriterProcessor::seize()
{
    setState(false);
}

int WriterProcessor::recover()
{
    if ((gCtrlBlock->getTermState()) || (!gCtrlBlock->getRecoverMode())) {
        return -1;
    }

    outStream->stopWrite(); 

    if (recoverID == -1) {
        recoverID = gNotifier->allocate();
    }
    setRecoverState(true); 
    Stream *st;
    if (gNotifier->freeze_i(recoverID, &st) != 0) {
        return -1;
    }
    if (outStream != st) {
        delete outStream;
    }
    log_debug("writer%d: have set the outStream to st %p, recoverID %d", handle, st, recoverID);
    recoverID = gNotifier->allocate();
    outStream = st;
    setReleaseState(false);
    setRecoverState(false);
    log_debug("writer%d: begin to notify notifyID %d", handle, notifyID);
    if (gNotifier->notify_i(notifyID) != 0) {
        return -1;
    }

    return 0;
}

void WriterProcessor::clean()
{
    outStream->stopWrite();
    gCtrlBlock->setFlowctlState(false);
    if (peerProcessor) {
        while (!peerProcessor->isLaunched()) {
            SysUtil::sleep(WAIT_INTERVAL);
        }  
        peerProcessor->join(); // ReaderProcessor
        delete peerProcessor;
    }
}

void WriterProcessor::setInQueue(MessageQueue * queue)
{
    inQueue = queue;
}

void WriterProcessor::setOutStream(Stream * stream)
{
    if (outStream == NULL) {
        outStream = stream;
    } else {
        log_debug("writer%d: begin to notify the stream %p, recoverID = %d", handle, stream, recoverID);
        if (peerProcessor) {
            peerProcessor->setInStream(stream);
        }
        while (recoverID == -1) {
            SysUtil::sleep(WAIT_INTERVAL);
        }
        if (notifyID == -1) {
            notifyID = gNotifier->allocate();
        }
        *(Stream **)gNotifier->getRetVal(recoverID) = stream;
        gNotifier->notify(recoverID);
        log_debug("writer%d: finish notify the recoverID %d", handle, recoverID);
        log_debug("writer%d: begin to freeze the notifyID %d", handle, notifyID);
        gNotifier->freeze(notifyID, NULL);
        log_debug("writer%d: finish freeze the notifyID %d", handle, notifyID);
        notifyID = gNotifier->allocate();
    }
}

MessageQueue * WriterProcessor::getInQueue()
{
    return inQueue;
}

void WriterProcessor::setPeerProcessor(ReaderProcessor * processor)
{
    peerProcessor =  processor;
}

ReaderProcessor *WriterProcessor::getPeerProcessor()
{
    return peerProcessor;
}
