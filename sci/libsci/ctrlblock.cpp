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

 Classes: CtrlBlock

 Description: Internal running information management (Note: STL does not 
              guarantee the safety of several readers & one writer cowork 
              together, and user threads can query group information at 
              runtime, so it's necessary to add a lock to protect these 
              read & write operations).
   
 Author: Tu HongJ, Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)
   11/27/10 ronglli      Add SCI Version
   01/16/12 ronglli      Add codes to detect SOCKET_BROKEN
   07/19/12 ronglli      Optimize the user query 

****************************************************************************/

#include <assert.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pwd.h>
#include <errno.h>

#include "ctrlblock.hpp"
#include "stream.hpp"
#include "exception.hpp"
#include "group.hpp"
#include "log.hpp"
#include "tools.hpp"
#include "packer.hpp"

#include "eventntf.hpp"
#include "handlerproc.hpp"
#include "embedagent.hpp"
#include "purifierproc.hpp"
#include "topology.hpp"
#include "routinglist.hpp"
#include "privatedata.hpp"
#include "message.hpp"
#include "queue.hpp"
#include "listener.hpp"
#include "processor.hpp"
#include "routerproc.hpp"
#include "filterproc.hpp"
#include "observer.hpp"
#include "initializer.hpp"

const long long FLOWCTL_THRESHOLD = 1024 * 1024 * 128LL;

CtrlBlock * CtrlBlock::instance = NULL;
extern SCI_msg_hndlr *gHndlr;
extern void *gParam;

CtrlBlock::CtrlBlock()
    : role(INVALID)
{
    char *envp = NULL;
    version = SCI_VERSION;
    userName = "";
    flowctlState = true;
    childHealthState = HEALTH;
    errChildren.clear();
    cnt_disable = 0;

    endInfo = NULL;
    
    routerProc = NULL;
    filterProc = NULL;
    purifierProc = NULL;
    handlerProc = NULL;
    observer = NULL;

    routerInQueue = NULL;
    filterInQueue = NULL;
    filterOutQueue = NULL;
    purifierOutQueue = NULL;
    upQueue = NULL;
    pollQueue = NULL;
    monitorInQueue = NULL;
    errorQueue = NULL;
    termState = false; // enter into term state
    recoverMode = 1; 
    waitParentInfo = false; //whether to wait for parent info updating

    parentStream = NULL;
    embedAgents.clear();
    enableID = gNotifier->allocate();

    // flow control threshold
    thresHold = FLOWCTL_THRESHOLD;
    envp = getenv("SCI_FLOWCTL_THRESHOLD");
    if(envp != NULL) {
        thresHold = ::atoll(envp);
    } 

    envp = ::getenv("SCI_DISABLE_IPV6");
    if (envp && (::strcasecmp(envp, "yes") == 0)) {
        Socket::setDisableIPv6(1);
    }
    envp = ::getenv("SCI_CONNECT_TIMES");
    if (envp != NULL) {
        int cnt = ::atoi(envp);
        if (cnt > 0) {
            Socket::setConnTimes(cnt);
        }
    }

    ::pthread_mutex_init(&mtx, NULL); 
}

CtrlBlock::~CtrlBlock()
{
    instance = NULL;
    ::pthread_mutex_destroy(&mtx);
}

void CtrlBlock::setRecoverMode(int mo)
{
    recoverMode = mo;
}

int CtrlBlock::getRecoverMode()
{
    return recoverMode;
}

void CtrlBlock::setTermState(bool mo)
{
    termState = mo;
}

bool CtrlBlock::getTermState()
{
    return termState;
}

void CtrlBlock::setParentInfoWaitState(bool mo)
{
    waitParentInfo = mo;
}

bool CtrlBlock::getParentInfoWaitState()
{
    return waitParentInfo;
}

CtrlBlock::ROLE CtrlBlock::getMyRole() 
{
    return role; 
}

void CtrlBlock::setMyRole(CtrlBlock::ROLE ro) 
{
    role = ro; 
}

int CtrlBlock::getMyHandle() 
{ 
    return handle; 
}

void CtrlBlock::setMyHandle(int hndl) 
{ 
    handle = hndl;
}

int CtrlBlock::getMyEmbedHandle() 
{ 
    return embed_handle; 
}

void CtrlBlock::setMyEmbedHandle(int hndl) 
{ 
    embed_handle = hndl;
}

sci_info_t * CtrlBlock::getEndInfo() 
{ 
    return endInfo; 
}

int CtrlBlock::getJobKey() 
{ 
    return jobKey; 
}

void CtrlBlock::setJobKey(int key) 
{ 
    jobKey = key;
}

int CtrlBlock::initClient(ROLE ro)
{
	char *envp = ::getenv("SCI_JOB_KEY");
	if (envp != NULL)
		jobKey = ::atoi(envp);
	envp = ::getenv("SCI_CLIENT_ID");
	if (envp != NULL)
		handle = ::atoi(envp);
	role = ro;

	return 0;
}

int CtrlBlock::init(sci_info_t * info)
{
    char *envp = NULL;

    if (info == NULL) {
        initClient(AGENT);
        return SCI_SUCCESS;
    } 
    if ((info->sci_version != 0) && (info->sci_version != version)) {
        return SCI_ERR_VERSION;
    }

    if (info->disable_sshauth == 1) { 
        ::setenv("SCI_ENABLE_SSHAUTH", "no", 1);
    }

    recoverMode = info->enable_recover;

    endInfo = (sci_info_t *) ::malloc(sizeof(sci_info_t));
    if (NULL == endInfo) {
        return SCI_ERR_NO_MEM;
    }
    ::memset(endInfo, 0, sizeof(sci_info_t));
    ::memcpy(endInfo, info, sizeof(sci_info_t));
    gHndlr = info->be_info.hndlr;
    gParam = info->be_info.param;

    switch (info->type) {
        case SCI_FRONT_END:
            handle = -1;
            role = FRONT_END;
            envp = ::getenv("SCI_JOB_KEY");
            if (envp) {
                // use user's job key
                jobKey = ::atoi(envp);
            } else {
                // generate a random job key
                ::srand((unsigned int) ::time(NULL));
                jobKey = ::rand();
            }
            break;
        case SCI_BACK_END:
			initClient(BACK_END);
            break;
        default:
            return SCI_ERR_INVALID_ENDTYPE;
    }

    return SCI_SUCCESS;
}

int CtrlBlock::numOfChildrenFds()
{
    int num = 0;
    RoutingList *rtList = NULL;
/*
    if (purifierProc) {
        while (!purifierProc->isLaunched()) {
            // before join, this thread should have been launched
            SysUtil::sleep(WAIT_INTERVAL);
        } 
    } */
    lock();
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        rtList = it->second->getRoutingList();
        num += rtList->numOfStreams();
    }
    unlock();

    return num;
}

int CtrlBlock::getChildrenSockfds(int *fds)
{
    int pos = 0;
    RoutingList *rtList = NULL;
/*
    if (purifierProc) {
        while (!purifierProc->isLaunched()) {
            // before join, this thread should have been launched
            SysUtil::sleep(WAIT_INTERVAL);
        } 
    } */
    lock();
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        rtList = it->second->getRoutingList();
        pos += rtList->getStreamsSockfds(&fds[pos]);
    }
    unlock();

    return pos;
}

bool CtrlBlock::allRouted()
{
    bool flag = false;
    int streams = 0;
    int queues = 0;

    RoutingList *rtList = NULL;

    lock();
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        rtList = it->second->getRoutingList();
        streams += rtList->numOfStreams();
        queues += rtList->numOfQueues();
    }

    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT) {
        flag = (queues == (streams + embedAgents.size())); // queueInfo contains the embed agent itself
    } else {
        flag = (queues == streams);
    }
    unlock();

    return flag;
}

int CtrlBlock::isActiveSockfd(int fd)
{
    int isSocket = 0;
    RoutingList *rtList = NULL;
    
    lock();
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        rtList = it->second->getRoutingList();
        isSocket = rtList->isActiveSockfd(fd);
        if (isSocket)
            break;
    }
    Stream * s = gInitializer->getInStream();
    if ((s != NULL) && (s->getSocket() == fd)) {
        if ((s->isReadActive()) || (s->isWriteActive())) {
            isSocket = 1;
        }
    }
    unlock();

    return isSocket;
}

bool CtrlBlock::allActive()
{
    bool active = true;
    RoutingList *rtList = NULL;
    
    lock();
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        rtList = it->second->getRoutingList();
        active = rtList->allActive();
        if (!active)
            break;
    }
    unlock();

    return active;
}

void CtrlBlock::term()
{
    gNotifier->freeze(enableID, NULL);
    termState = true;
    if (purifierProc) {
        purifierProc->release();
        delete purifierProc;
    }
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        delete it->second;
    }
    lock();
    embedAgents.clear();
    errChildren.clear();
    recoverChildren.clear();
    unlock();
    if (handlerProc) {
        handlerProc->release();
        delete handlerProc;
    }
    clean();
}

void CtrlBlock::setRecover(int hndl)
{
    AGENT_MAP::iterator it;

    lock();
    recoverChildren[hndl] = 0;
    unlock();
}

void CtrlBlock::clearRecover()
{
    lock();
    recoverChildren.clear();
    unlock();
}

void CtrlBlock::clearRecover(int hndl)
{
    lock();
    recoverChildren.erase(hndl);
    unlock();
}

int CtrlBlock::getRecoverNum()
{
    int num = 0;

    lock();
    num = recoverChildren.size();
    unlock();

    return num;
}

int CtrlBlock::getRecoverChildren(int *children)
{
    int i = 0;
    RECOVER_MAP::iterator it;

    lock();
    for (it = recoverChildren.begin(); it != recoverChildren.end(); it++) {
        children[i] = it->first;
    }
    unlock();

    return 0;
}

int CtrlBlock::getRecoverChildren(vector<int> & children)
{
    RECOVER_MAP::iterator it;
    int times = 30;
    char *envp = getenv("SCI_RECONN_TIME");

    if (envp != NULL) {
        int tmp = atoi(envp);
        if (tmp > 0) {
            times = tmp * 10;
        }
    }

    lock();
    for (it = recoverChildren.begin(); it != recoverChildren.end(); it++) {
        it->second++;
        if (it->second > times) {
            children.push_back(it->first);
        }
    }
    unlock();

    return 0;
}

bool CtrlBlock::checkRouting(int hndl)
{
    bool rt = false;
    routingInfo *router = NULL;
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        router = it->second->getRoutingList()->getRouter(hndl);
        if ((router != NULL) && (router->stream != NULL)) {
            rt = true;
            break;
        }
    }
    
    return rt;
}

EmbedAgent * CtrlBlock::findAgent(int hndl)
{
    EmbedAgent *agent = NULL;

    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        if (it->second->getRoutingList()->getRouter(hndl) != NULL) {
            agent = it->second;
        }
    }

    return agent;
}

EmbedAgent * CtrlBlock::getAgent(int hndl)
{
    EmbedAgent *agent;
    lock();
    assert(embedAgents.find(hndl) != embedAgents.end());
    agent = embedAgents[hndl];
    unlock();

    return agent;
}

void CtrlBlock::clean()
{
    routerProc = NULL;
    filterProc = NULL;
    purifierProc = NULL;

    routerInQueue = NULL;
    filterInQueue = NULL;
    filterOutQueue = NULL;
    purifierOutQueue = NULL;
    upQueue = NULL;
    pollQueue = NULL;
    monitorInQueue = NULL;
    errorQueue = NULL;

    parentStream = NULL;

    if (observer != NULL) {
        delete observer;
        observer = NULL;
    }

    role = INVALID;
    if (endInfo) {
        ::free(endInfo);
        endInfo = NULL;
    }
}

void CtrlBlock::enable()
{
}

void CtrlBlock::disable()
{
    if (!isEnabled())
        return;

    lock();
    if (getMyRole() == BACK_AGENT) {
        cnt_disable++;
        if (cnt_disable < (embedAgents.size() + 1)) {
            unlock();
            return;
        }
    }
    unlock();
    gNotifier->notify(enableID);
}

bool CtrlBlock::isEnabled() 
{ 
    return gNotifier->getState(enableID);
}

void CtrlBlock::releasePollQueue()
{
    // so far, valid for polling mode only
    assert(role != AGENT);
    try {
        if(observer != NULL) {
            observer->notify();
        } else {
            log_error("CtrlBlock: releasePollQueue: observer is NULL");
        }
        if(pollQueue != NULL) {
            Message *msg = new Message(Message::INVALID_POLL);
            pollQueue->produce(msg);
        } else {
            log_error("CtrlBlock: releasePollQueue: pollQueue is NULL");
        }
    } catch (Exception &e) {
        log_error("releasePollQueue: exception %s", e.getErrMsg());
    } catch (std::bad_alloc) {
        log_error("releasePollQueue: out of memory");
    } catch (...) {
        log_error("releasePollQueue: unknown exception");
    }
}

void CtrlBlock::notifyChildHealthState(int hndl, int hState)
{
    int num = 0;
    int *cList = NULL;
    bool found = false;
    Message::Type typ = getErrMsgType(hState);
    if (typ == Message::UNKNOWN)
        return;

    lock();
    RoutingList *rtList = NULL;
    AGENT_MAP::iterator it;
    for (it = embedAgents.begin(); it != embedAgents.end(); it++) {
        rtList = it->second->getRoutingList();
        if (rtList->isSuccessorExist(hndl)) {
            if (hndl < 0) {
                num = rtList->numOfBEOfSuccessor(hndl);
                assert(num);
                cList = (int *) malloc(num * sizeof(int));
                rtList->retrieveBEListOfSuccessor(hndl, cList);
            } else {
                num = 1;
                cList = (int *) malloc(sizeof(int));
                cList[0] = hndl;
            }
            found = true;
            break;
        }
    }
    if (!found) {
        unlock();
        return;
    }
    assert(cList != NULL);

    try {
        for (int i = 0; i < num; i++) {
            errChildren.insert(cList[i]);
        }

        // if not fe, it should forward the broken msg to its parent
        if (getMyRole() != FRONT_END) { 
            Message *msg = new Message();
            Packer packer;
            packer.packInt(num);
            for (int i = 0; i < num; i++) {
                packer.packInt(cList[i]);
            }

            char *bufs[1];
            int sizes[1];
            bufs[0] = packer.getPackedMsg();
            sizes[0] = packer.getPackedMsgLen();
            msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes, typ);
            getUpQueue()->produce(msg);
        }

        // so far, valid for polling mode only
        if (getMyRole() != AGENT) {
            sci_mode_t mode;
            if (getMyRole() == FRONT_END)
                mode = getEndInfo()->fe_info.mode;
            else
                mode = getEndInfo()->be_info.mode;
            if (mode == SCI_POLLING) {
                observer->notify();
                Message *msg = new Message(typ);
                pollQueue->produce(msg);
            }
        }
    } catch (Exception &e) {
        log_error("notifyChildHealthState: exception %s", e.getErrMsg());
    } catch (std::bad_alloc) {
        log_error("notifyChildHealthState: out of memory");
    } catch (...) {
        log_error("notifyChildHealthState: unknown exception");
    }
    unlock();
    setChildHealthState(hState);
    free(cList);
}

void CtrlBlock::notifyChildHealthState(Message * msg)
{
    int num = 0;
    int *cList = NULL;
    Message::Type typ = msg->getType();
    int hState = getErrState(typ);
    if (hState == UNKNOWN) {
        delete msg;
        return;
    }

    lock();
    // upqueue can be deleted when it is terminating
    if (getTermState()) {
        delete msg;
        unlock();
        return;
    }

    Packer packer(msg->getContentBuf());
    num = packer.unpackInt();
    cList = (int *) malloc(num * sizeof(int));
    assert(cList != NULL);

    try {
        for (int i = 0; i < num; i++) {
            cList[i] = packer.unpackInt();
            errChildren.insert(cList[i]);
        }

        // if not fe, it should forward the broken msg to its parent
        if (getMyRole() != FRONT_END) { 
            getUpQueue()->produce(msg);
        } else {
            delete msg;
        }

        // so far, valid for polling mode only
        if (getMyRole() != AGENT) {
            sci_mode_t mode;
            if (getMyRole() == FRONT_END)
                mode = getEndInfo()->fe_info.mode;
            else
                mode = getEndInfo()->be_info.mode;
            if (mode == SCI_POLLING) {
                observer->notify();
                Message *tmpmsg = new Message(typ);
                pollQueue->produce(tmpmsg);
            }
        }
    } catch (Exception &e) {
        log_error("notifyChildHealthState: exception %s", e.getErrMsg());
    } catch (std::bad_alloc) {
        log_error("notifyChildHealthState: out of memory");
    } catch (...) {
        log_error("notifyChildHealthState: unknown exception");
    }
    unlock();
    setChildHealthState(hState);
    free(cList);
}

int CtrlBlock::getErrChildren(int * num, int **list)
{
    lock();
    ERRORCHILDREN_LIST tmpErrChildren = errChildren;
    unlock();

    *num = tmpErrChildren.size();
    *list = (int *) malloc(sizeof(int) * (*num));
    ::memset(*list, 0, sizeof(int) * (*num));
    log_debug("getErrChildren: err Children: size = %d", *num);

    ERRORCHILDREN_LIST::iterator it;
    int i = 0;
    for (it = tmpErrChildren.begin(); it != tmpErrChildren.end(); it++) {
        (*list)[i] = *it;
        log_debug("getErrChildren: err Children: list[%d] = %d", i, (*list)[i]);
        i++;
    }
    return 0;
}

void CtrlBlock::setObserver(Observer *ob) 
{
    observer = ob;
}

extern void makeKey();
PrivateData * CtrlBlock::getPrivateData()
{
    PrivateData *pData = (PrivateData *)pthread_getspecific(Thread::key);
    if (!pData)
    {
        if (!purifierProc) {
            EmbedAgent* agent = getAgent(handle);
            if (!agent)
                return NULL;
            agent->registPrivateData();
        } else {
            pData = new PrivateData(purifierProc->getRoutingList(), purifierProc->getFilterList(), NULL);
            int rc = pthread_once(&(Thread::once), makeKey);
            rc = pthread_setspecific(Thread::key, pData);
        }
        pData = (PrivateData *)pthread_getspecific(Thread::key);
    }
    return pData;
}

Topology * CtrlBlock::getTopology() 
{ 
    PrivateData *pData = getPrivateData();
    return pData->getRoutingList()->getTopology();
}

Observer * CtrlBlock::getObserver() {
    return observer;
}

void CtrlBlock::setRouterInQueue(MessageQueue * queue)
{
    routerInQueue = queue;
}

void CtrlBlock::setFilterInQueue(MessageQueue *queue) 
{
    filterInQueue = queue;
}

void CtrlBlock::setPollQueue(MessageQueue *queue) 
{
    pollQueue = queue;
}

void CtrlBlock::setMonitorInQueue(MessageQueue *queue) 
{
    monitorInQueue = queue;
}

void CtrlBlock::setErrorQueue(MessageQueue *queue) 
{
    errorQueue = queue;
}

MessageQueue * CtrlBlock::getRouterInQueue()
{
    return routerInQueue;
}

MessageQueue * CtrlBlock::getFilterInQueue() 
{
    return filterInQueue;
}

MessageQueue * CtrlBlock::getPollQueue() 
{
    return pollQueue;
}

MessageQueue * CtrlBlock::getErrorQueue() 
{
    return errorQueue;
}

MessageQueue * CtrlBlock::getMonitorInQueue() 
{
    return monitorInQueue;
}

void CtrlBlock::setRouterProcessor(RouterProcessor *proc) 
{
    routerProc = proc;
}

void CtrlBlock::setFilterProcessor(FilterProcessor *proc) 
{
    filterProc = proc;
}
        
void CtrlBlock::setHandlerProcessor(HandlerProcessor *proc) 
{
    handlerProc = proc;
}
        
void CtrlBlock::setPurifierProcessor(PurifierProcessor *proc) 
{
    purifierProc = proc;
}
        
void CtrlBlock::setUpQueue(MessageQueue * queue)
{
    upQueue = queue;
}

MessageQueue * CtrlBlock::getUpQueue()
{
        return upQueue;
}

RouterProcessor * CtrlBlock::getRouterProcessor() 
{
    PrivateData *pData = getPrivateData();
    return pData->getRouterProcessor();
}
        
FilterProcessor * CtrlBlock::getFilterProcessor() 
{
    PrivateData *pData = getPrivateData();
    return pData->getFilterProcessor();
}

FilterList * CtrlBlock::getFilterList() 
{
    PrivateData *pData = getPrivateData();
    return pData->getFilterList();
}

PurifierProcessor * CtrlBlock::getPurifierProcessor() 
{
    return purifierProc;
}

void CtrlBlock::setFlowctlThreshold(long long th)
{
    thresHold = th;
}

long long CtrlBlock::getFlowctlThreshold()
{
    return thresHold;
}

int CtrlBlock::getVersion()
{
    return version;
}

int CtrlBlock::setUsername()
{
    if (userName == "") {
        int rc = 0;
        long size = sysconf(_SC_GETPW_R_SIZE_MAX);
        struct passwd pwd;
        struct passwd *result = NULL;
        char *pwdBuf = new char[size];
        while(1) {
            rc = getpwuid_r(::getuid(), &pwd, pwdBuf, size, &result);
            if ((rc == EINTR) || (rc == EMFILE) || (rc == ENFILE)) {
                SysUtil::sleep(WAIT_INTERVAL);
                continue;
            }
            if (NULL == result) {
                delete []pwdBuf;
                log_error("CtrlBlock: fail to get the user info! errno = %d.", errno);
                return SCI_ERR_INVALID_USER;
            } else {
                break;
            }
        }
        userName = pwd.pw_name; 
        delete []pwdBuf;
    }
    return SCI_SUCCESS;
}

string & CtrlBlock::getUsername()
{
    return userName;
}

RoutingList * CtrlBlock::getRoutingList()
{
    PrivateData *pData = getPrivateData();
    return pData->getRoutingList();
}

void CtrlBlock::addEmbedAgent(int hndl, EmbedAgent *agent)
{
    lock();
    embedAgents[hndl] = agent;
    unlock();
}

void CtrlBlock::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void CtrlBlock::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

void CtrlBlock::setFlowctlState(bool state)
{
    flowctlState = state;
}

bool CtrlBlock::getFlowctlState()
{
    return flowctlState;
}

void CtrlBlock::setChildHealthState(int state)
{
    childHealthState = state;
}

int CtrlBlock::checkChildHealthState()
{
    int rc = SCI_SUCCESS;
    switch (childHealthState) {
        case HEALTH:
            rc = SCI_SUCCESS;
            break;
        case ERROR_CHILD_BROKEN:
            rc = SCI_ERR_CHILD_BROKEN; 
            break;
        case ERROR_DATA:
            rc = SCI_ERR_DATA; 
            break;
        case ERROR_THREAD:
            rc = SCI_ERR_THREAD; 
            break;
        default:
            rc = SCI_ERR_THREAD; 
            break;
    }
    return rc;
}

Message::Type CtrlBlock::getErrMsgType(int hState)
{
    Message::Type typ;
    switch (hState) {
        case HEALTH:
        case UNKNOWN:
            // If it is in health/unknown state, should not produce notify msg
            typ = Message::UNKNOWN;
            break;
        case ERROR_CHILD_BROKEN:
            typ = Message::SOCKET_BROKEN;
            break;
        case ERROR_DATA:
            typ = Message::ERROR_DATA;
            break;
        case ERROR_THREAD:
            typ = Message::ERROR_THREAD;
            break;
        default:
            typ = Message::ERROR_THREAD;
            break;
    }
    return typ;
}

int CtrlBlock::getErrState(Message::Type typ)
{
    int hState;
    switch (typ) {
        case Message::SOCKET_BROKEN:
            hState = ERROR_CHILD_BROKEN;
            break;
        case Message::ERROR_DATA:
            hState = ERROR_DATA;
            break;
        case Message::ERROR_THREAD:
            hState = ERROR_THREAD;
            break;
        default:
            // If it is incorrect msg type
            hState = UNKNOWN;
            break;
    }
    return hState;
}

