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

#include <stdlib.h>
#include <stdio.h>
#include <assert.h>
#include <string.h>

#include <vector>

#include "log.hpp"
#include "packer.hpp"
#include "group.hpp"
#include "tools.hpp"
#include "exception.hpp"
#include "stream.hpp"
#include "sshfunc.hpp"

#include "routinglist.hpp"
#include "eventntf.hpp"
#include "initializer.hpp"
#include "message.hpp"
#include "queue.hpp"
#include "ctrlblock.hpp"
#include "filterproc.hpp"
#include "readerproc.hpp"
#include "topology.hpp"
#include "writerproc.hpp"
#include "eventntf.hpp"
#include "dgroup.hpp"


using namespace std;

const int MAX_SUCCESSOR_NUM = 1024;
const int MAX_SEGMENT_SIZE = 11680;
const int MIN_SEGMENT_SIZE = 1440;  // 1500 - 40 - 20 (ethernet MTU - tcp/ip header - message header)

RoutingList::RoutingList(int hndl)
    : handle(hndl), maxSegmentSize(MAX_SEGMENT_SIZE), filterProc(NULL), myDistriGroup(NULL), topology(NULL)
{
    char *envp = ::getenv("SCI_SEGMENT_SIZE");
    if (envp != NULL) {
        maxSegmentSize = atoi(envp);
        maxSegmentSize = maxSegmentSize > MIN_SEGMENT_SIZE ? maxSegmentSize : MIN_SEGMENT_SIZE;
    }

    if (handle == -1) {
        // this is a front end, not parent
        myDistriGroup = new DistributedGroup(0);
    } else {
        int pid = -1;
        envp = ::getenv("SCI_PARENT_ID");
        if (envp) {
            pid = ::atoi(envp);
        } else {
            throw Exception(Exception::INVALID_LAUNCH);
        }
        myDistriGroup = new DistributedGroup(pid);
    }

    if (gCtrlBlock->getMyRole() != CtrlBlock::BACK_END) {
        topology = new Topology(0); // 0 is an impossible agent ID so it will be changed when it receives a real topology
    }
    successorList = new int[MAX_SUCCESSOR_NUM];
    queueInfo.clear();
    routers.clear();
    ::pthread_mutex_init(&mtx, NULL); 
}

RoutingList::~RoutingList()
{  
    delete myDistriGroup;
    delete [] successorList;
    delete topology;
    ::pthread_mutex_destroy(&mtx);
}

void RoutingList::parseCmd(Message *msg)
{
    bool notify = false;
    int rc = SCI_SUCCESS;
    if (msg->getType() == Message::GROUP_CREATE) {
        Packer packer(msg->getContentBuf());

        int num_bes = packer.unpackInt();
        int be_list[num_bes];
        for (int i=0; i<num_bes; i++) {
            be_list[i] = packer.unpackInt();
        }

        myDistriGroup->create(num_bes, be_list, msg->getGroup());
        bcast(msg->getGroup(), msg);

        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            notify = true;
        }
    } else if (msg->getType() == Message::GROUP_FREE) {
        sci_group_t group = msg->getGroup();
        
        bcast(group, msg);
        myDistriGroup->remove(group);

        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            notify = true;
        }
    } else if (msg->getType() == Message::GROUP_OPERATE) {
        Packer packer(msg->getContentBuf());

        sci_op_t op = (sci_op_t) packer.unpackInt();
        sci_group_t group1 = (sci_group_t) packer.unpackInt();
        sci_group_t group2 = (sci_group_t) packer.unpackInt();

        rc = myDistriGroup->operate(group1, group2, op, msg->getGroup());
        if (rc == SCI_SUCCESS) {
            bcast(msg->getGroup(), msg);
        }

        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            notify = true;
        }
    } else if (msg->getType() == Message::GROUP_OPERATE_EXT) {
        Packer packer(msg->getContentBuf());

        sci_op_t op = (sci_op_t) packer.unpackInt();
        sci_group_t group = (sci_group_t) packer.unpackInt();
        int num_bes = packer.unpackInt();
        int be_list[num_bes];
        for (int i=0; i<num_bes; i++) {
            be_list[i] = packer.unpackInt();
        }

        rc = myDistriGroup->operateExt(group, num_bes, be_list, op, msg->getGroup());
        if (rc == SCI_SUCCESS) {
            bcast(msg->getGroup(), msg);
        }

        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            notify = true;
        }
    } else if (msg->getType() == Message::GROUP_DROP) {
        myDistriGroup->dropSuccessor(msg->getID());
    } else if (msg->getType() == Message::GROUP_MERGE) {
        DistributedGroup subDistriGroup(-1);
        subDistriGroup.unpackMsg(*msg);

        if (subDistriGroup.getPID() == handle) {
            // if this message is from my son
            myDistriGroup->merge(msg->getID(), subDistriGroup, false);
        } else if (isSuccessorExist(subDistriGroup.getPID())){
            // if this message is from my grandson
            myDistriGroup->merge(msg->getID(), subDistriGroup, false);
        } else {
            // if this message is from my nephew
            myDistriGroup->merge(msg->getID(), subDistriGroup, true);

            // now update its parent id to me
            subDistriGroup.setPID(handle);

            // repack a message and send to my parent
            Message *newmsg = subDistriGroup.packMsg();
            filterProc->getOutQueue()->produce(newmsg);
        }
    } else {
        assert(!"should never be here");
    }

    if (notify) {
        void *ret = gNotifier->getRetVal(msg->getID());
        *((int *) ret) = rc;
        gNotifier->notify(msg->getID());
    }
}

void RoutingList::propagateGroupInfo()
{
    // propgate my group information to my parent
    Message *msg = myDistriGroup->packMsg();
    if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT) {
        filterProc->getOutQueue()->produce(msg);
    } else if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END) {
        gCtrlBlock->getUpQueue()->produce(msg);
    } else {
        assert(!"should not be here");
    }
}

int RoutingList::getSegments(Message *msg, Message ***segments, int ref)
{
    int i = 0;
    int segnum = (msg->getContentLen() + maxSegmentSize - 1) / maxSegmentSize + 1;
    int size = 0;
    char *ptr = msg->getContentBuf();
    sci_group_t gid = msg->getGroup();
    Message::Type typ = msg->getType();
    int mid = msg->getID();
    int mfid = msg->getFilterID();
    int hfid = mfid;
    int mlen = msg->getContentLen();
    *segments = (Message **)::malloc(segnum * sizeof(Message *));
    Message **segs = *segments;

    if ((mfid != SCI_FILTER_NULL) || (typ != Message::COMMAND)) {
        hfid = SCI_JOIN_SEGMENT;
    }
    ::memset(segs, 0, segnum * sizeof(Message *));
    segs[0] = new Message();
    segs[0]->build(hfid, gid, 0, NULL, NULL, Message::SEGMENT, segnum);
    segs[0]->setRefCount(ref);

    for (i = 1; i < segnum; i++) {
        segs[i] = new Message();
        size = (i < (segnum - 1)) ? maxSegmentSize : ((mlen - 1) % maxSegmentSize + 1);
        segs[i]->build(mfid, gid, 1, &ptr, &size, typ, mid);
        segs[i]->setRefCount(ref);
        ptr += size;
    }

    return segnum;
}

int RoutingList::bcast(sci_group_t group, Message *msg)
{
    if (group > SCI_GROUP_ALL) {
        int hndl = querySuccessorId((int) group);
        if (hndl == INVLIDSUCCESSORID) {
            return SCI_ERR_GROUP_NOTFOUND;
        } else if (hndl == VALIDBACKENDIDS) {
            ucast((int)group, msg);
        } else {
            ucast(hndl, msg);
        }
        return SCI_SUCCESS;
    }
    
    if (!isGroupExist(group)) {
        return SCI_ERR_GROUP_NOTFOUND;
    }

    splitBcast(group, msg);
    
    return SCI_SUCCESS;
}

void RoutingList::splitBcast(sci_group_t group, Message *msg)
{
    int numSor = numOfSuccessor(group);
    retrieveSuccessorList(group, successorList);
    mcast(msg, successorList, numSor);
}

void RoutingList::mcast(Message *msg, int *sorList, int num)
{
    int i = 0;

    if (msg->getContentLen() <= maxSegmentSize) {
        msg->setRefCount(msg->getRefCount() + num);
        for (i = 0; i < num; i++) {
            queryQueue(sorList[i])->produce(msg);
        }
        return;
    }

    Message **segments;
    int segnum = getSegments(msg, &segments, num);
    for (i = 0; i < num; i++) {
        queryQueue(sorList[i])->multiProduce(segments, segnum);
    }
    ::free(segments);
}

void RoutingList::ucast(int successor_id, Message *msg, int refInc)
{
    log_debug("Processor Router: send msg to successor %d", successor_id);
    mcast(msg, &successor_id, refInc);

    return;
}

void RoutingList::initSubGroup(int successor_id, int start_be_id, int end_be_id)
{
    char qName[64] = {0};
    MessageQueue *queue = NULL;

    if (successor_id != VALIDBACKENDIDS) {
        queue = new MessageQueue();
        ::sprintf(qName, "Agent%d_inQ", successor_id);
        queue->setName(qName);
        mapQueue(successor_id, queue);
    } else {
        int i = 0;
        for (i = start_be_id; i <= end_be_id; i++) {
            queue = new MessageQueue();
            ::sprintf(qName, "BE%d_inQ", i);
            queue->setName(qName);
            mapQueue(i, queue);
        }
    }

    myDistriGroup->initSubGroup(successor_id, start_be_id, end_be_id);
}

int RoutingList::startReading(int hndl)
{
    ROUTING_MAP::iterator it = routers.find(hndl);
    assert(it != routers.end());
    ReaderProcessor *reader = it->second.processor->getPeerProcessor();
    reader->start();

    return 0;
}

int RoutingList::setRecoverChildren()
{
    QUEUE_MAP::iterator it;
    ROUTING_MAP::iterator rit;

    for (it = queueInfo.begin(); it != queueInfo.end(); ++it) {
        int hndl = it->first;
        rit = routers.find(hndl);
        if ((rit == routers.end()) || (rit->second.stream == NULL)) {
            gCtrlBlock->setRecover(hndl);
        }
    }

    return 0;
}

int RoutingList::startReaders()
{
    ReaderProcessor *reader = NULL;
    ROUTING_MAP::iterator pit;

    for (pit = routers.begin(); pit != routers.end(); ++pit) {
        while (pit->second.processor == NULL) {
            SysUtil::sleep(WAIT_INTERVAL);
        }
        reader = pit->second.processor->getPeerProcessor();
        while (reader == NULL) {
            SysUtil::sleep(WAIT_INTERVAL);
            reader = pit->second.processor->getPeerProcessor();
        }
        reader->start();
    }

    return 0;
}

int RoutingList::numOfStreams()
{
    int size = 0;
    ROUTING_MAP::iterator it;

    lock();
    for (it = routers.begin(); it != routers.end(); ++it) {
        if (it->second.stream != NULL) {
            size++;
        }
    }
    unlock();
    return size;
}

int RoutingList::getStreamsSockfds(int *fds)
{
    int i = 0;
    ROUTING_MAP::iterator it;

    lock();
    for (it = routers.begin(); it != routers.end(); ++it) {
        if (it->second.stream == NULL) {
            continue;
        }
        fds[i] = it->second.stream->getSocket();
        i++;
    }
    unlock();

    return i;
}

int RoutingList::isActiveSockfd(int fd)
{
    int isSocket = 0;
    ROUTING_MAP::iterator it;

    lock();
    for (it = routers.begin(); it != routers.end(); ++it) {
        if (it->second.stream == NULL) {
            continue;
        }
        if (fd == it->second.stream->getSocket()) {
            if ((it->second.stream->isReadActive()) || (it->second.stream->isWriteActive()) ){
                isSocket = 1;
                break;
            }
        }
    }
    unlock();

    return isSocket;
}

bool RoutingList::allActive()
{
    bool active = true;
    ROUTING_MAP::iterator it;

    lock();
    for (it = routers.begin(); it != routers.end(); ++it) {
        if ((it->second.stream == NULL) || (!(it->second.stream->isReadActive())) || (!(it->second.stream->isWriteActive())) ){
            active = false;
            break;
        }
    }
    unlock();

    return active;
}

void RoutingList::mapRouters(int hndl, WriterProcessor *writer, Stream *stream)
{
    lock();
    if (writer != NULL) {
        routers[hndl].processor = writer;
    }
    routers[hndl].stream = stream;
    unlock();
}

int RoutingList::startRouting(int hndl, Stream *stream)
{
    char name[64] = {0};
    MessageQueue *inQ = queryQueue(hndl);
    while (inQ == NULL) {
        SysUtil::sleep(WAIT_INTERVAL);
        inQ = queryQueue(hndl);
    }

    if (stream == NULL) {
        mapRouters(hndl, NULL, NULL);
        return -1;
    }
    WriterProcessor *writer = NULL;
    if ((routers.find(hndl) != routers.end()) && (routers[hndl].stream != NULL)) {
        int count = 0;
        if (!gCtrlBlock->getRecoverMode()) {
            log_error("Duplicated client are trying to connect!!!");
            return -1;
        }
        writer = routers[hndl].processor;
        while(!(writer->getRecoverState())) {
            SysUtil::sleep(WAIT_INTERVAL);
            count++;
            if (count >= 5000) {
                log_warn("Duplicated client are trying to connect!!!");
                return -1;
            }
        }

        writer->setOutStream(stream);
        mapRouters(hndl, NULL, stream);
        return 0;
    }

    log_debug("routers[%d].stream = %p", hndl, stream);
    ReaderProcessor *reader = new ReaderProcessor(hndl);
    reader->setInStream(stream);
    reader->setOutQueue(filterProc->getInQueue());
    ::sprintf(name, "Reader%d", hndl);
    reader->setName(name);

    writer = new WriterProcessor(hndl);
    writer->setInQueue(inQ);
    writer->setOutStream(stream);
    ::sprintf(name, "Writer%d", hndl);
    writer->setName(name);

    // reader is a peer processor of writer
    writer->setPeerProcessor(reader);
    reader->setPeerProcessor(writer);
    mapRouters(hndl, writer, stream);

    log_debug("The Reader%d thread has been newed!", hndl);
    writer->start();
    reader->start(); 

    return 0;
}

routingInfo * RoutingList::getRouter(int hndl)
{
    routingInfo *rInfo = NULL;
    ROUTING_MAP::iterator pit = routers.find(hndl);

    lock();
    if (pit != routers.end()) {
        rInfo = &pit->second;
    }
    unlock();

    return rInfo;
}

int RoutingList::stopRouting(int hndl)
{
    ROUTING_MAP::iterator pit = routers.find(hndl);
    if (pit != routers.end()) {
        pit->second.processor->release();
        delete pit->second.processor;

        routers.erase(hndl);
        queueInfo.erase(hndl);
    }

    return 0;
}

int RoutingList::stopRouting()
{
    // waiting for all processor threads terminate
    ROUTING_MAP::iterator pit;
    for (pit = routers.begin(); pit != routers.end(); ++pit) {
        if (pit->second.processor == NULL) {
            continue;
        }
        pit->second.processor->release();
        delete pit->second.processor;
    }

    routers.clear();
    queueInfo.clear();

    return 0;
}

bool RoutingList::allRouted()
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT) {
        char *envp = getenv("SCI_EMBED_AGENT");
        if ((envp != NULL) && (strcasecmp(envp, "yes") == 0)) {
            return (numOfQueues() == (numOfStreams() + 1));  // queueInfo contains itself
        }
    }

    return (numOfQueues() == numOfStreams()); 
}

void RoutingList::addBE(sci_group_t group, int successor_id, int be_id, bool init)
{
    if (init) {
        char qName[64] = {0};
        int qID = 0;
        MessageQueue *queue = new MessageQueue();

        if (successor_id == VALIDBACKENDIDS) {
            qID = be_id;
            ::sprintf(qName, "BE%d_inQ", qID);
        } else {
            qID = successor_id;
            ::sprintf(qName, "Agent%d_inQ", qID);
        }
        queue->setName(qName);
        mapQueue(qID, queue);
    }

    myDistriGroup->addBE(group, successor_id, be_id);
}

void RoutingList::removeBE(int be_id)
{
    myDistriGroup->removeBE(be_id);
}

void RoutingList::removeGroup(sci_group_t group)
{
    myDistriGroup->remove(group);
}

void RoutingList::updateParentId(int pid)
{
    ::setenv("SCI_PARENT_ID", SysUtil::itoa(pid).c_str(), 1);
    myDistriGroup->setPID(pid);
}

bool RoutingList::isGroupExist(sci_group_t group)
{
    return myDistriGroup->isGroupExist(group);
}

bool RoutingList::isSuccessorExist(int successor_id)
{
    return myDistriGroup->isSuccessorExist(successor_id);
}


int RoutingList::numOfBE(sci_group_t group)
{
    return myDistriGroup->numOfBE(group);
}

int RoutingList::numOfSuccessor(sci_group_t group)
{
    return myDistriGroup->numOfSuccessor(group);
}

int RoutingList::numOfBEOfSuccessor(int successor_id)
{
    if (successor_id >= 0) {
        // if it is a back end
        return 1;
    }

    return myDistriGroup->numOfBEOfSuccessor(successor_id);
}

int RoutingList::querySuccessorId(int be_id)
{
    return myDistriGroup->querySuccessorId(be_id);
}

void RoutingList::retrieveBEList(sci_group_t group, int * ret_val)
{
    assert(ret_val);
    myDistriGroup->retrieveBEList(group, ret_val);
}

void RoutingList::retrieveSuccessorList(sci_group_t group, int * ret_val)
{
    assert(ret_val);
    myDistriGroup->retrieveSuccessorList(group, ret_val);
}

void RoutingList::retrieveBEListOfSuccessor(int successor_id, int * ret_val)
{
    assert(ret_val);
    myDistriGroup->retrieveBEListOfSuccessor(successor_id, ret_val);
}

void RoutingList::mapQueue(int hndl, MessageQueue *queue)
{
    lock();
    queueInfo[hndl] = queue;
    unlock();
}

MessageQueue * RoutingList::queryQueue(int hndl)
{       
    MessageQueue *queue = NULL;

    lock();
    QUEUE_MAP::iterator qit = queueInfo.find(hndl);
    if (qit != queueInfo.end()) {
        queue = (*qit).second;
    }
    unlock();

    return queue;
}

int RoutingList::numOfQueues()
{
    int size;
    lock();
    size = queueInfo.size();
    unlock();
    return size;
}

void RoutingList::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void RoutingList::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

Topology * RoutingList::getTopology()
{
    return topology;
}

FilterProcessor * RoutingList::getFilterProcessor()
{
    return filterProc;
}

void RoutingList::setFilterProcessor(FilterProcessor *proc)
{
    filterProc = proc;
}
