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

 Classes: None

 Description: SCI APIs.
   
 Author: Liu Wei, Nicole Nie, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 lwbjcdl      Initial code (D153875)
   01/16/12 ronglli      Add codes to detect SOCKET_BROKEN

****************************************************************************/

#include <assert.h>
#include <stdlib.h>
#include <math.h>
#include <string.h>
#include <unistd.h>
#include <vector>
#include <sys/socket.h>

#include "sci.h"

#include "tools.hpp"
#include "log.hpp"
#include "envvar.hpp"
#include "packer.hpp"

#include "general.hpp"

#include "message.hpp"
#include "queue.hpp"
#include "topology.hpp"
#include "ctrlblock.hpp"
#include "routinglist.hpp"
#include "eventntf.hpp"
#include "initializer.hpp"
#include "observer.hpp"
#include "filter.hpp"
#include "listener.hpp"
#include "filterlist.hpp"
#include "filterproc.hpp"
#include "routerproc.hpp"
#include "allocator.hpp"
#include "sshfunc.hpp"
#include "exception.hpp"

const char * ErrRetMsg[] = {
    "Succeeded.",
    "Invalid host file.",
    "Invalid end type, can only be SCI_FRONT_END or SCI_BACK_END.",
    "Error occurred when doing the initialization.",
    "The API is called by invalid end type.",
    "Error occurred when searching the group.",
    "Error occurred when searching the filter.",
    "Invalid filter.",
    "Error occurred when searching the backend.",
    "Unknown information.",
    "Uninitialized SCI execution environment.",
    "Can't free predefined group.",
    "The group is an empty group.",
    "Invalid group operator specified.",
    "Can't unload predefined filter.",
    "A polling timeout occurred after timeout milliseconds elapsed",
    "Invalid job key specified by SCI_JOB_KEY.",
    "Can only be used in polling mode.",
    "Invalid filter id",
    "The successor id list contains non-existed successor id.",
    "The back id already existed.",
    "Out of memory.",
    "Failed to launch client(s).",
    "Invalid polling calls.",
    "Invalid user.",
    "Invalid mode.",
    "Error occurred when searching the agent.",
    "Invalid version number.",              
    "Error occurred when doing the SSH-based authentication.",
    "Exception occurred",    
    "Invalid error message number.",                  

    "The parent is broken.",
    "The child is broken.",
    "Error occurred when doing the recovery.",
    "Recover failed.",
    "Received incorrect data.",
    "Error occurred in the threads.",
};

SCI_msg_hndlr *gHndlr = NULL;
void *gParam = NULL;
// Initialization & Termination

int SCI_Initialize(sci_info_t *info)
{
    int rc;
    if (gCtrlBlock->getMyRole() != CtrlBlock::INVALID) {
        log_warn("Has already been initialized");
        return SCI_SUCCESS;
    }

    rc = gCtrlBlock->init(info);
    if (rc != SCI_SUCCESS)
        return rc;

    return gInitializer->init();
}

int SCI_Terminate()
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;

    try {
        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            Message *msg = new Message();
            msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 0, NULL, NULL, Message::QUIT);
            gCtrlBlock->getRouterInQueue()->produce(msg);
        }
        gCtrlBlock->term();

        delete gNotifier;
        delete gInitializer;
        delete gCtrlBlock;
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    return SCI_SUCCESS;
}

int SCI_Release()
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID) {
        return SCI_ERR_UNINTIALIZED;
    }
    try {
        int role = gCtrlBlock->getMyRole();
        
        int num_fds = gCtrlBlock->numOfChildrenFds();
        log_debug("there are total %d children", num_fds);
        if (num_fds > 0) {
            int * fd_list = (int *)malloc(num_fds * sizeof(int));
            gCtrlBlock->getChildrenSockfds(fd_list);
            for(int i=0; i < num_fds; i++) {
                log_debug("close child fd %d", fd_list[i]);
                ::shutdown(fd_list[i], SHUT_RDWR);
                ::close(fd_list[i]);
            }
            free(fd_list);
        }
        if (role != CtrlBlock::FRONT_END) {
            if (gInitializer->getInStream() != NULL) {
                int p_fd = gInitializer->getInStream()->getSocket();
                log_debug("close parent fd %d", p_fd);
                ::shutdown(p_fd, SHUT_RDWR);
                ::close(p_fd);
            }
        }
        if (role == CtrlBlock::FRONT_END) { 
            Message *msg = new Message();
            msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 0, NULL, NULL, Message::RELEASE); 
            gCtrlBlock->getRouterInQueue()->produce(msg); 
        }
        gCtrlBlock->term();

        delete gNotifier;
        delete gInitializer;
        delete gCtrlBlock;
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    return SCI_SUCCESS;
}

int SCI_Parentinfo_update(char * parentAddr, int port)
{
    int rc = -1;
    if ((NULL == parentAddr) || (port <= 0))
        return SCI_ERR_UNKNOWN_INFO;

    if ((gCtrlBlock->getTermState()) || (!gCtrlBlock->getRecoverMode()) 
            || (!gCtrlBlock->getParentInfoWaitState()))
        return SCI_ERR_INVALID_CALLER;

    log_debug("Parentinfo_update: addr = %s, port = %d", parentAddr, port);
    rc = gInitializer->updateParentInfo(parentAddr, port);
    if (rc != SCI_SUCCESS) {
        log_debug("Parentinfo_update: failed to update info, rc = %d", rc);
        return rc;
    }

    return SCI_SUCCESS;
}

int SCI_Parentinfo_wait()
{
    if ((gCtrlBlock->getTermState()) || (!gCtrlBlock->getRecoverMode()))
        return SCI_ERR_INVALID_CALLER;

    gCtrlBlock->setParentInfoWaitState(true);
    log_debug("Parentinfo_wait: set the parentInfoWait state to true");

    return SCI_SUCCESS;
}

int SCI_Recover_setmode(int mode)
{
    gCtrlBlock->setRecoverMode(mode);
    log_debug("Recover_setmode: set the recover mode to %d", mode);
    return SCI_SUCCESS;
}

// Information Query


int SCI_Query_host(int hndl, char *hbuf, int len)
{
    routingInfo *rInfo = NULL;
    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
        return SCI_ERR_INVALID_CALLER;

    rInfo = gCtrlBlock->getRoutingList()->getRouter(hndl);
    if ((rInfo != NULL) && (rInfo->stream != NULL)) {
        string &peerHost =  rInfo->stream->getPeerHost();
        memset(hbuf, '\0', len);
        strncpy(hbuf, peerHost.c_str(), len - 1);
    }

    return 0;
}

int SCI_Query(sci_query_t query, void *ret_val)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if(NULL == ret_val) 
        return SCI_ERR_UNKNOWN_INFO;

    int *p = (int *) ret_val;
    switch (query) 
    {
        case CURRENT_VERSION:
            *p = gCtrlBlock->getVersion();
            break;
        case JOB_KEY:
            *p = gCtrlBlock->getJobKey();
            break;
        case NUM_BACKENDS:
            if(gCtrlBlock->getMyRole() == CtrlBlock::BACK_END) 
                return SCI_ERR_INVALID_CALLER;
            *p = gCtrlBlock->getTopology()->getBENum();
            break;
        case BACKEND_ID:
            if((gCtrlBlock->getMyRole() != CtrlBlock::BACK_END)
                    && (gCtrlBlock->getMyRole() != CtrlBlock::BACK_AGENT)) 
                return SCI_ERR_INVALID_CALLER;
            *p = gCtrlBlock->getMyHandle();
            break;
        case POLLING_FD:
            if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT)
                return SCI_ERR_INVALID_CALLER;
            if (!gCtrlBlock->getObserver())
                return SCI_ERR_MODE;
            else
                *p = gCtrlBlock->getObserver()->getPollFd();
            break;
        case PIPEWRITE_FD:
            if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT)
                return SCI_ERR_INVALID_CALLER;
            if (!gCtrlBlock->getObserver())
                return SCI_ERR_MODE;
            else
                *p = gCtrlBlock->getObserver()->getPipeWriteFd();
            break;
        case NUM_FILTERS:
            *p = gCtrlBlock->getFilterList()->numOfFilters();
            break;
        case FILTER_IDLIST:
            gCtrlBlock->getFilterList()->retrieveFilterList(p);
            break;
        case AGENT_ID:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END) 
                return SCI_ERR_INVALID_CALLER;
            *p = gCtrlBlock->getMyHandle();
            break;
        case NUM_SUCCESSORS:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            *p = gCtrlBlock->getRoutingList()->numOfSuccessor(SCI_GROUP_ALL);
            break;
        case SUCCESSOR_IDLIST:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            gCtrlBlock->getRoutingList()->retrieveSuccessorList(SCI_GROUP_ALL, p);
            break;
        case HEALTH_STATUS:
            if (gCtrlBlock->getTermState()) {
                *p = 2;
            } else if (gCtrlBlock->isEnabled()) {
                *p = 0; // 0 - normal
            } else {
                *p = 1; // 1 - exited
            }
            break;
        case AGENT_LEVEL:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            *p = gCtrlBlock->getTopology()->getLevel();
            break;
        case LISTENER_PORT:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            if (gInitializer->getListener() == NULL)
                return SCI_ERR_INVALID_CALLER;
            *p = gInitializer->getListener()->getBindPort();
            break;
        case NUM_LISTENER_FDS:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            if (gInitializer->getListener() == NULL)
                return SCI_ERR_INVALID_CALLER;
            *p = gInitializer->getListener()->numOfSockFds();
            break;
        case LISTENER_FDLIST:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            if (gInitializer->getListener() == NULL)
                return SCI_ERR_INVALID_CALLER;
            gInitializer->getListener()->getSockFds(p);
            break;
        case PARENT_SOCKFD:
            if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END)
                return SCI_ERR_INVALID_CALLER;
            *p = gInitializer->getInStream()->getSocket();
            break;
        case NUM_CHILDREN_FDS:
            *p = gCtrlBlock->numOfChildrenFds();
            break;
        case CHILDREN_FDLIST:
            gCtrlBlock->getChildrenSockfds(p);
            break;
        case RECOVER_STATUS:
            if (gCtrlBlock->allActive()) {
                *p = 1; 
            } else {
                *p = 0; 
            }
            break;
        case NUM_RECOVERY:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            *p = gCtrlBlock->getRecoverNum();
            break;
        case RECOVERY_LIST:
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
                return SCI_ERR_INVALID_CALLER;
            gCtrlBlock->getRecoverChildren(p);
            break;
        default:
            return SCI_ERR_UNKNOWN_INFO;
    }

    return SCI_SUCCESS;
}

int SCI_Error(int err_code, char *err_msg, int msg_size)
{
    if ((err_msg == NULL) || (msg_size <= 0)) {
        return SCI_ERR_NO_MEM;
    }
    memset(err_msg, 0, msg_size);

    if (err_code == 0) {
        strncpy(err_msg, ErrRetMsg[err_code], msg_size);
        return SCI_SUCCESS;
    }

    if ((err_code <= -2001) && (err_code >= -2030)){
        int index = err_code * (-1) % 2000;
        strncpy(err_msg, ErrRetMsg[index], msg_size);
        return SCI_SUCCESS;
    }

    if ((err_code <= -5000) && (err_code >= -5005)){
        int index = err_code * (-1) % 5000;
        strncpy(err_msg, ErrRetMsg[index + 31], msg_size);
        return SCI_SUCCESS;
    }

    return SCI_ERR_MSG;

}

int SCI_Query_errchildren(int *num, int **id_list)
{
    return gCtrlBlock->getErrChildren(num, id_list);
}

// Communication

int SCI_Bcast(int filter_id, sci_group_t group, int num_bufs, void *bufs[], int sizes[])
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    if (group > SCI_GROUP_ALL) {
        if (!gCtrlBlock->getTopology()->hasBE((int)group))
            return SCI_ERR_GROUP_NOTFOUND;
    } else {
        if (!gCtrlBlock->getRoutingList()->isGroupExist(group))
            return SCI_ERR_GROUP_NOTFOUND;
    }

    int rc = SCI_SUCCESS;
    rc = gCtrlBlock->checkChildHealthState();
    if (rc != SCI_SUCCESS)
        return rc;

    try {
        Message *msg = new Message();
        msg->build(filter_id, group, num_bufs, (char **)bufs, sizes, Message::COMMAND);
        log_debug("Produced a message bcast command, message group=%d, message size=%d",
            (int) group, msg->getContentLen());
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        rc = SCI_ERR_NO_MEM;
    }
  
    return rc;
}

int SCI_Upload(int filter_id, sci_group_t group, int num_bufs, void *bufs[], int sizes[])
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if ((gCtrlBlock->getMyRole() != CtrlBlock::BACK_END)
            && (gCtrlBlock->getMyRole() != CtrlBlock::BACK_AGENT))
        return SCI_ERR_INVALID_CALLER;

    try {
        Message *msg = new Message();
        msg->build(filter_id, group, num_bufs, (char **)bufs, sizes, Message::DATA);
        gCtrlBlock->getUpQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    return SCI_SUCCESS;
}

int SCI_Poll(int timeout)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT)
        return SCI_ERR_INVALID_CALLER;

    sci_mode_t mode;
    if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END)
        mode = gCtrlBlock->getEndInfo()->fe_info.mode;
    else
        mode = gCtrlBlock->getEndInfo()->be_info.mode;
    if (mode != SCI_POLLING)
        return SCI_ERR_MODE;
/*
    if (gCtrlBlock->getPollQueue()->getSize() == 0) {
        // no messages in the polling queue
        // TODO
        return SCI_ERR_POLL_INVALID;
    }
*/
    int rc = SCI_SUCCESS;
    rc = gCtrlBlock->checkChildHealthState();
    if (rc != SCI_SUCCESS)
        return rc;

    Message *msg = gCtrlBlock->getPollQueue()->consume(timeout);
    if (msg) {
        switch(msg->getType()) {
            case Message::COMMAND:
            case Message::DATA:
                try {
                    gHndlr(gParam, msg->getGroup(), msg->getContentBuf(), msg->getContentLen());
                } catch (...) {
                    // TODO
                }

                gCtrlBlock->getObserver()->unnotify();
                break;
            case Message::INVALID_POLL:
                rc = SCI_ERR_POLL_INVALID;
                gCtrlBlock->getObserver()->unnotify();
                break;
            case Message::SOCKET_BROKEN:
                rc = SCI_ERR_CHILD_BROKEN;
                log_debug("SCI_Poll: received msg SOCKET_BROKEN");
                gCtrlBlock->getObserver()->unnotify();
                break;
            case Message::ERROR_DATA:
                rc = SCI_ERR_DATA;
                log_debug("SCI_Poll: received msg ERROR_DATA");
                gCtrlBlock->getObserver()->unnotify();
                break;
            case Message::ERROR_THREAD:
                rc = SCI_ERR_THREAD;
                log_debug("SCI_Poll: received msg ERROR_THREAD");
                gCtrlBlock->getObserver()->unnotify();
                break;
            default:
                rc = SCI_ERR_UNKNOWN_INFO;
                log_error("SCI_Poll: received unknown command");
                gCtrlBlock->getObserver()->unnotify();
                break;
        }

        gCtrlBlock->getPollQueue()->remove();
    } else {
        rc = SCI_ERR_POLL_TIMEOUT;
    }

    return rc;
}

// Group

int SCI_Group_create(int num_bes, int * be_list, sci_group_t * group)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    assert(be_list);
    for (int i=0; i<num_bes; i++) {
        if (!gCtrlBlock->getTopology()->hasBE(be_list[i]))
            return SCI_ERR_BACKEND_NOTFOUND;
    }

    int msgID;
    try {
        Packer packer;
        packer.packInt(num_bes);
        for (int i=0; i<num_bes; i++) {
            packer.packInt(be_list[i]);
        }

        char *bufs[1];
        int sizes[1];

        bufs[0] = packer.getPackedMsg();
        sizes[0] = packer.getPackedMsgLen();
        msgID = gNotifier->allocate();

        Message *msg = new Message();
        gAllocator->allocateGroup(group);
        msg->build(SCI_FILTER_NULL, *group, 1, bufs, sizes, Message::GROUP_CREATE, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }
    
    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_Group_free(sci_group_t group)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    if (group>=SCI_GROUP_ALL)
        return SCI_ERR_GROUP_PREDEFINED;

    if (!gCtrlBlock->getRoutingList()->isGroupExist(group))
        return SCI_ERR_GROUP_NOTFOUND;

    int msgID;
    try {
        Message *msg = new Message();
        msgID = gNotifier->allocate();
        msg->build(SCI_FILTER_NULL, group, 0, NULL, NULL, Message::GROUP_FREE, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }
    
    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_Group_operate(sci_group_t group1, sci_group_t group2,
        sci_op_t op, sci_group_t *newgroup)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    if (!gCtrlBlock->getRoutingList()->isGroupExist(group1))
        return SCI_ERR_GROUP_NOTFOUND;

    if (!gCtrlBlock->getRoutingList()->isGroupExist(group2))
        return SCI_ERR_GROUP_NOTFOUND;

    if ((op!=SCI_UNION) && (op!=SCI_INTERSECTION) && (op!=SCI_DIFFERENCE))
        return SCI_ERR_INVALID_OPERATOR;

    int msgID;
    try {
        Packer packer;
        packer.packInt((int) op);
        packer.packInt((int) group1);
        packer.packInt((int) group2);

        char *bufs[1];
        int sizes[1];

        bufs[0] = packer.getPackedMsg();
        sizes[0] = packer.getPackedMsgLen();
        msgID = gNotifier->allocate();

        Message *msg = new Message();
        gAllocator->allocateGroup(newgroup);
        msg->build(SCI_FILTER_NULL, *newgroup, 1, bufs, sizes, Message::GROUP_OPERATE, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }
    
    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_Group_operate_ext(sci_group_t group, int num_bes, int *be_list, 
        sci_op_t op, sci_group_t *newgroup)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    if (!gCtrlBlock->getRoutingList()->isGroupExist(group))
        return SCI_ERR_GROUP_NOTFOUND;

    assert(be_list);
    for (int i=0; i<num_bes; i++) {
        if (!gCtrlBlock->getTopology()->hasBE(be_list[i]))
            return SCI_ERR_BACKEND_NOTFOUND;
    }

    int msgID;
    try {
        Packer packer;
        packer.packInt((int) op);
        packer.packInt((int) group);
        packer.packInt(num_bes);
        for (int i=0; i<num_bes; i++) {
            packer.packInt(be_list[i]);
        }

        char *bufs[1];
        int sizes[1];

        bufs[0] = packer.getPackedMsg();
        sizes[0] = packer.getPackedMsgLen();
        msgID = gNotifier->allocate();

        Message *msg = new Message();
        gAllocator->allocateGroup(newgroup);
        msg->build(SCI_FILTER_NULL, *newgroup, 1, bufs, sizes, Message::GROUP_OPERATE_EXT, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_Group_query(sci_group_t group, sci_group_query_t query, void *ret_val)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END)
        return SCI_ERR_INVALID_CALLER;

    if (!gCtrlBlock->getRoutingList()->isGroupExist(group))
        return SCI_ERR_GROUP_NOTFOUND;

    switch(query) 
    {
        case GROUP_MEMBER_NUM:
            *((int *) ret_val) = gCtrlBlock->getRoutingList()->numOfBE(group);
            break;
        case GROUP_MEMBER:
            gCtrlBlock->getRoutingList()->retrieveBEList(group, (int *) ret_val);
            break;
        case GROUP_SUCCESSOR_NUM:
            *((int *) ret_val) = gCtrlBlock->getRoutingList()->numOfSuccessor(group);
            break;
        case GROUP_SUCCESSOR:
            gCtrlBlock->getRoutingList()->retrieveSuccessorList(group, (int *) ret_val);
            break;
        default:
            return SCI_ERR_UNKNOWN_INFO;
    }
    
    return SCI_SUCCESS;
}


// Filter

int SCI_Filter_load(sci_filter_info_t *filter_info) 
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    assert(filter_info);
    if (filter_info->filter_id == SCI_FILTER_NULL)
        return SCI_ERR_FILTER_PREDEFINED;

    if (filter_info->filter_id < SCI_FILTER_NULL)
        return SCI_ERR_FILTER_ID;

    int msgID;
    try {
        Filter *filter = new Filter(*filter_info);
        Message *msg = filter->packMsg();
        msgID = gNotifier->allocate();
        msg->setID(msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_Filter_unload(int filter_id)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;
    
    if (filter_id == SCI_FILTER_NULL)
        return SCI_ERR_FILTER_PREDEFINED;

    if (filter_id < SCI_FILTER_NULL)
        return SCI_ERR_FILTER_ID;

    int msgID;
    try {
        Message *msg = new Message();
        msgID = gNotifier->allocate();
        msg->build(filter_id, SCI_GROUP_ALL, 0, NULL, NULL, Message::FILTER_UNLOAD, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_Filter_bcast(int filter_id, int num_successors, int * successor_list, int num_bufs, 
        void *bufs[], int sizes[])
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END) {
        return SCI_ERR_INVALID_CALLER;
    }

    for (int i=0; i<num_successors; i++) {
        if (!gCtrlBlock->getRoutingList()->isSuccessorExist(successor_list[i]))
            return SCI_ERR_INVALID_SUCCESSOR;
    }

    try {
        Message *msg = new Message();
        int nextFilterID = filter_id;
        if (filter_id == SCI_FILTER_NULL) {
            nextFilterID = gCtrlBlock->getRouterProcessor()->getCurFilterID();
        }
        sci_group_t curGroup = gCtrlBlock->getRouterProcessor()->getCurGroup();
        msg->build(nextFilterID, curGroup, num_bufs, (char **)bufs, sizes, Message::COMMAND);
        msg->setRefCount(num_successors);
        gCtrlBlock->getRoutingList()->mcast(msg, successor_list, num_successors);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    return SCI_SUCCESS;
}

int SCI_Filter_upload(int filter_id, sci_group_t group, int num_bufs, void *bufs[], int sizes[])
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_END) {
        return SCI_ERR_INVALID_CALLER;
    }

    try {
        Filter *filter = NULL;
        if (filter_id != SCI_FILTER_NULL) {
            filter = gCtrlBlock->getFilterList()->getFilter(filter_id);
        }
        int curFilterID = gCtrlBlock->getFilterProcessor()->getCurFilterID();
        Message *msg = new Message();
        msg->build(curFilterID, group, num_bufs, (char **)bufs, sizes, Message::DATA);
        if (filter) {
            filter->input(group, msg->getContentBuf(), msg->getContentLen());
            delete msg;
        } else {    
            gCtrlBlock->getFilterProcessor()->deliever(msg);
        }
    } catch (Exception &e) {
        return SCI_ERR_EXCEPTION;
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    return SCI_SUCCESS;
}

// Dynamic Back End

int SCI_BE_add(sci_be_t *be)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;

    if (be->id >= 0) { // user-assigned back end id
        if (gCtrlBlock->getTopology()->hasBE(be->id)) 
            return SCI_ERR_BACKEND_EXISTED;
    } else { // SCI allocated back end id
        gAllocator->allocateBE(&(be->id));
    }

    int msgID;
    try {
        Packer packer; 
        packer.packStr(be->hostname);
        packer.packInt(be->level);

        char *bufs[1];
        int sizes[1];

        bufs[0] = packer.getPackedMsg();
        sizes[0] = packer.getPackedMsgLen();

        Message *msg = new Message();
        msgID = gNotifier->allocate();
        msg->build(SCI_FILTER_NULL, (sci_group_t) (be->id), 1, bufs, sizes, Message::BE_ADD, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}

int SCI_BE_remove(int be_id)
{
    if (gCtrlBlock->getMyRole() == CtrlBlock::INVALID)
        return SCI_ERR_UNINTIALIZED;
    
    if (gCtrlBlock->getMyRole() != CtrlBlock::FRONT_END)
        return SCI_ERR_INVALID_CALLER;
    
    if (!gCtrlBlock->getTopology()->hasBE(be_id)) 
        return SCI_ERR_BACKEND_NOTFOUND;

    int msgID;
    try {
        Message *msg = new Message();
        msgID = gNotifier->allocate();
        msg->build(SCI_FILTER_NULL, (sci_group_t) be_id, 0, NULL, NULL, Message::BE_REMOVE, msgID);
        gCtrlBlock->getRouterInQueue()->produce(msg);
    } catch (std::bad_alloc) {
        return SCI_ERR_NO_MEM;
    }

    int rc;
    gNotifier->freeze(msgID, &rc);
    return rc;
}
