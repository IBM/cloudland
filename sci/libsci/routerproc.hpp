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

 Classes: RouterProcessor

 Description: Properties of class 'RouterProcessor':
    input: a. a stream 
           b. a message queue
    output: a set of message queues
    action: route the message to the designated destination
   
 Author: Nicole Nie     Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#ifndef _ROUTERPROC_HPP
#define _ROUTERPROC_HPP

#include "sci.h"

#include "processor.hpp"

class MessageQueue;
class Stream;
class RoutingList;
class FilterList;

class WriterProcessor;

class RouterProcessor : public Processor 
{
    private:
        Stream              *inStream;
        RoutingList         *routingList;
        FilterList          *filterList;

        int                 curFilterID;
        sci_group_t         curGroup;
        bool                joinSegs;

        WriterProcessor     *peerProcessor; 

    public:
        RouterProcessor(int hndl, RoutingList *rlist, FilterList *flist);
        ~RouterProcessor();

        virtual Message * read();
        virtual void process(Message *msg);
        virtual void write(Message *msg);
        virtual void seize();
        virtual int recover();
        virtual void clean();

        int getCurFilterID();
        sci_group_t getCurGroup();

        void setInQueue(MessageQueue *queue);
        MessageQueue * getInQueue();
        void setInStream(Stream * stream);
        RoutingList * getRoutingList();

        void setPeerProcessor(WriterProcessor * processor);
        WriterProcessor * getPeerProcessor();
};

#endif

