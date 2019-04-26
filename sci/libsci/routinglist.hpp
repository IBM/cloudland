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

 Classes: RoutingList

 Description: Provide routing services for all threads.
   
 Author: Nicole Nie, Liu Wei, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   05/08/09 nieyy        Initial code (D156654)
   01/16/12 ronglli      Add codes to retrieve BE list

****************************************************************************/

#ifndef _ROUTINGLIST_HPP
#define _ROUTINGLIST_HPP

#include <pthread.h>
#include "sci.h"
#include "general.hpp"

#include <map>

class Message;
class DistributedGroup;
class Stream;
class MessageQueue;
class WriterProcessor;
class FilterProcessor;
class Topology;

#define SCI_JOIN_SEGMENT -1001

using namespace std;

struct routingInfo {
    Stream *stream;
    WriterProcessor *processor;
};

class RoutingList
{      
    public:
        typedef map<int, MessageQueue*> QUEUE_MAP;
        typedef map<int, struct routingInfo> ROUTING_MAP;

    private:
        int                    handle;
        DistributedGroup      *myDistriGroup;
        int                   *successorList;
        int                    maxSegmentSize;
        pthread_mutex_t        mtx;
        QUEUE_MAP              queueInfo;
        ROUTING_MAP            routers;
        FilterProcessor       *filterProc;
        Topology              *topology;
        
    public:
        RoutingList(int hndl);
        ~RoutingList();
        
        void lock();
        void unlock();
        void parseCmd(Message *msg);
        void propagateGroupInfo();
        void mapQueue(int hndl, MessageQueue *queue);
        MessageQueue * queryQueue(int hndl);
        int numOfQueues();
        void mapRouters(int hndl, WriterProcessor *writer, Stream *stream);

        int startRouting(int hndl, Stream *stream);
        int startReading(int hndl);
        int startReaders();
        int stopRouting(int hndl);
        int stopRouting();
        bool allRouted();
        routingInfo * getRouter(int hndl);

        int setRecoverChildren();
        void setFilterProcessor(FilterProcessor *proc);
        FilterProcessor * getFilterProcessor();
        Topology * getTopology();

        int getSegments(Message *msg, Message ***segments, int ref);
        int bcast(sci_group_t group, Message *msg);
        void ucast(int successor_id, Message *msg, int refInc = 1);
        void mcast(Message *msg, int *sorList, int num);
        void splitBcast(sci_group_t group, Message *msg);

        void initSubGroup(int successor_id, int start_be_id, int end_be_id);
        void addBE(sci_group_t group, int successor_id, int be_id, bool init = false);
        void removeBE(int be_id);
        void removeGroup(sci_group_t group);

        void updateParentId(int pid);

        bool isGroupExist(sci_group_t group);
        bool isSuccessorExist(int successor_id);

        int numOfBE(sci_group_t group);
        int numOfSuccessor(sci_group_t group);
        int numOfBEOfSuccessor(int successor_id);
        
        int querySuccessorId(int be_id);
        void retrieveBEList(sci_group_t group, int *ret_val);
        void retrieveSuccessorList(sci_group_t, int *ret_val);
        void retrieveBEListOfSuccessor(int successor_id, int * ret_val);
        int getStreamsSockfds(int *fds);
        int numOfStreams();

        int isActiveSockfd(int fd);
        bool allActive();
};

#endif

