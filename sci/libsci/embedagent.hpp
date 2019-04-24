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

 Classes: EmbedAgent

 Description: embedded agent in front-end and back-end
   
 Author: Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code ()

****************************************************************************/

#ifndef _EMBEDAGENT_HPP
#define _EMBEDAGENT_HPP

#include <stdlib.h>

class MessageQueue;
class Stream;
class FilterProcessor;
class RouterProcessor;
class WriterProcessor;
class RoutingList;
class FilterList;
class Listener;
class PrivateData;

class EmbedAgent {
    private:
        int                 handle;
        Stream              *inStream;
        Stream              *outStream;
        MessageQueue        *filterInQ;
        MessageQueue        *filterOutQ;
        MessageQueue        *routerInQ;
        FilterProcessor     *filterProc;
        RouterProcessor     *routerProc;
        WriterProcessor     *writerProc;
        RoutingList         *routingList;
        FilterList          *filterList;

    public:
        EmbedAgent();
        ~EmbedAgent();
        int init(int hndl, Stream *stream, MessageQueue *inQ, MessageQueue *outQ = NULL);
        int work();
        int syncWait();
        void setParent(EmbedAgent *agent);
        void setChild(EmbedAgent *child);
        MessageQueue *getRouterInQ();
        MessageQueue *getUpQueue();
        int registPrivateData();
        FilterProcessor *getFilterProcessor();
        RoutingList *getRoutingList();
        PrivateData *genPrivateData();
};

#endif
