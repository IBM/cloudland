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
    output: a. a message queue
    action: use user-defined filter handlers to process the messages
   
 Author: Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#ifndef _FILTERPROC_HPP
#define _FILTERPROC_HPP

#include "processor.hpp"

class Stream;
class MessageQueue;
class Observer;
class FilterList;

class FilterProcessor : public Processor 
{
    private:
        FilterList          *filterList;

        Observer            *observer;

        bool                filtered;
        int                 curFilterID;

    public:
        FilterProcessor(int hndl = -1, FilterList *flist = NULL);
        ~FilterProcessor();

        virtual Message * read();
        virtual void process(Message *msg);
        virtual void write(Message *msg);
        virtual void seize();
        virtual int recover();
        virtual void clean();

        void deliever(Message *msg);
        int getCurFilterID();

        void setInQueue(MessageQueue *queue);
        void setOutQueue(MessageQueue *queue);

        void setObserver(Observer *ob);
        MessageQueue * getInQueue();
        MessageQueue * getOutQueue();
};

#endif

