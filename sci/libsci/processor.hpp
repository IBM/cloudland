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
   
 Author: Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)

****************************************************************************/

#ifndef _PROCESSOR_HPP
#define _PROCESSOR_HPP

#include <string>

using namespace std;

#include "thread.hpp"

class Message;
class MessageQueue;

class Processor : public Thread 
{      
    protected:
        string              name;

        // for performance counting
        int                 totalCount;
        int                 totalSize;
        MessageQueue        *inQueue;
        MessageQueue        *outQueue;
        int                 hState;

    public:
        Processor(int hndl = -1);

        virtual void run();

        virtual Message * read() = 0;
        virtual void process(Message *msg) = 0;
        virtual void write(Message *msg) = 0;
        virtual void seize() = 0;
        virtual int recover() = 0;
        virtual void clean() = 0;

        virtual bool isActive();
        
        void dump();
        void setName(char *str) { name = str; }
        string getName() { return name; }
        virtual void release();
};

#endif

