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
    output: a. two message queues
    action: purify message, discarded useless messages
   
 Author: Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   04/28/09 nieyy      Initial code (F156654)

****************************************************************************/

#ifndef _PURIFERPROC_HPP
#define _PURIFERPROC_HPP

#include "sci.h"
#include "processor.hpp"

class Stream;
class MessageQueue;
class Observer;
class WriterProcessor;
class RoutingList;
class FilterList;

class PurifierProcessor : public Processor 
{
    private:
        Stream              *inStream;
        MessageQueue        *outErrorQueue;
        WriterProcessor     *peerProcessor;
        RoutingList         *routingList;
        FilterList          *filterList;
        Observer            *observer;
        SCI_msg_hndlr       *hndlr;
        void                *param;

        bool                isCmd;
        bool                isError;
        bool                joinSegs;

    public:
        PurifierProcessor(int hndl = -1);
        ~PurifierProcessor();

        virtual Message * read();
        virtual void process(Message *msg);
        virtual void write(Message *msg);
        virtual void seize();
        virtual int recover();
        virtual void clean();

        void setInStream(Stream *stream);
        Stream * getInStream();
        void setOutQueue(MessageQueue *queue);
        void setInQueue(MessageQueue *queue);
        MessageQueue * getInQueue();
        void setOutErrorQueue(MessageQueue *queue);
        void setPeerProcessor(WriterProcessor *processor);
        WriterProcessor * getPeerProcessor();
        void setObserver(Observer *ob);
        RoutingList * getRoutingList();
        FilterList * getFilterList();
};

#endif

