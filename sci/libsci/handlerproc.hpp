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

 Classes: HandlerProcessor

 Description: Properties of class 'HandlerProcessor':
    input: a. a message queue
    output: none
    action: use handler in sci_info_t to process the messages
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#ifndef _HANDLERPROC_HPP
#define _HANDLERPROC_HPP

#include "sci.h"
#include "processor.hpp"

class Stream;
class MessageQueue;

class HandlerProcessor : public Processor 
{
    private:
        SCI_msg_hndlr       *hndlr;
        void                *param;

    public:
        HandlerProcessor(int hndl = -1);
        ~HandlerProcessor();

        virtual Message * read();
        virtual void process(Message *msg);
        virtual void write(Message *msg);
        virtual void seize();
        virtual int recover();
        virtual void clean();

        void setInQueue(MessageQueue *queue);
};

#endif

