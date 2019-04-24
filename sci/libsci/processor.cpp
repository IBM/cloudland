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

 Classes: Processor

 Description: Properties of class 'Processor':
    input: a. a stream 
           b. a message queue
    output: a. none
            b. a stream
            c. one or multiple message queues
    action: any kind message processing actions
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#include "processor.hpp"
#include <assert.h>

#include "ctrlblock.hpp"
#include "log.hpp"
#include "tools.hpp"
#include "exception.hpp"
#include "socket.hpp"

#include "message.hpp"
#include "queue.hpp"

Processor::Processor(int hndl) 
    : Thread(hndl), inQueue(NULL), outQueue(NULL), hState(CtrlBlock::HEALTH)
{
    name = "Processor";

    totalCount = 0;
    totalSize = 0;
}

void Processor::run()
{
    log_debug("Processor %s: started", name.c_str());
    
    Message *msg = NULL;
    while (getState() || isActive()) {
        try {
            // read a message
            msg = read();
            if (msg == NULL) {
                log_debug("Processor %s: read a NULL message", name.c_str());
                continue;
            }

            totalCount++;
            totalSize += msg->getContentLen();
            log_debug("Processor %s: processing a message, type=%d, filter ID=%d, group=%d, size=%d", 
                    name.c_str(), msg->getType(), msg->getFilterID(), msg->getGroup(), msg->getContentLen());

            // process the message
            process(msg);

            // write the message
            write(msg);

            log_debug("Processor %s: finished", name.c_str());
        } catch (Exception &e) {
            if (e.getErrCode() == Exception::INVALID_SIGNATURE) {
                log_warn("Receives a misleading message");
                continue;
            }
            hState = CtrlBlock::ERROR_THREAD;
            seize();
            log_error("Processor %s: exception %s", name.c_str(), e.getErrMsg());
            break;
        } catch (SocketException &e) {
            log_error("Processor %s: socket exception %s", name.c_str(), e.getErrMsg().c_str());
            if (e.getErrCode() == SocketException::NET_ERR_DATA) {
                hState = CtrlBlock::ERROR_DATA;
            } else {
                if (recover() == 0) {
                    continue;
                }
                hState = CtrlBlock::ERROR_CHILD_BROKEN;
            }
            seize();
            if (e.getErrCode() == SocketException::NET_ERR_CLOSED) {
                log_warn("Processor %s: socket exception %s", name.c_str(), e.getErrMsg().c_str());
            }
            break;
        } catch (ThreadException &e) {
            hState = CtrlBlock::ERROR_THREAD;
            seize();
            log_error("Processor %s: thread exception %d", name.c_str(), e.getErrCode());
            break;
        } catch (std::bad_alloc) {
            hState = CtrlBlock::ERROR_THREAD;
            seize();
            log_error("Processor %s: out of memory", name.c_str());
            break;
        } catch (...) {
            hState = CtrlBlock::ERROR_THREAD;
            seize();
            log_error("Processor %s: unknown exception", name.c_str());
            break;
        }
    }

    // do cleanup works
    clean();
    
    log_debug("Processor %s: exited", name.c_str());
}

void Processor::release()
{
    while (!isLaunched()) {
        // before join, this thread should have been launched
        SysUtil::sleep(WAIT_INTERVAL);
    } 
    setState(false);
    if (inQueue)
        inQueue->release();
    join();
}

bool Processor::isActive()
{
    if ((inQueue) && (inQueue->getState()))
        return (inQueue->getSize() > 0);
    return false;
}

void Processor::dump()
{
    log_perf("Until now, processor %s has processed %d messages, total size is %d bytes",
        name.c_str(), totalCount, totalSize);
}

