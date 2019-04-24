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

****************************************************************************/

#ifndef _READERPROC_HPP
#define _READERPROC_HPP

#include "processor.hpp"

class Stream;
class MessageQueue;

class WriterProcessor;

class ReaderProcessor : public Processor 
{
    private:
        Stream              *inStream;
        MessageQueue        *outErrorQueue;
        int                 recoverID;
        int                 notifyID;
        WriterProcessor     *peerProcessor;
        void releasePeer(WriterProcessor * writer);

    public:
        ReaderProcessor(int hndl = -1);
        ~ReaderProcessor();

        virtual Message * read();
        virtual void process(Message *msg);
        virtual void write(Message *msg);
        virtual void seize();
        virtual int recover();
        virtual void clean();

        void setInStream(Stream *stream);
        void setOutQueue(MessageQueue *queue);

        void setOutErrorQueue(MessageQueue *queue);

        void setPeerProcessor(WriterProcessor* processor);
        WriterProcessor *getPeerProcessor();

};

#endif

