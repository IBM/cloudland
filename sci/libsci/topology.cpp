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

 Classes: BEMap, Topology, Launcher

 Description: Runtime topology manipulation.
   
 Author: Nicole Nie, Liu Wei, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code (D153875)

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include <stdlib.h>
#include <stdio.h>
#include <assert.h>
#include <math.h>
#include <ctype.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/wait.h>

#include "log.hpp"
#include "tools.hpp"
#include "packer.hpp"
#include "exception.hpp"
#include "ipconverter.hpp"

#include "topology.hpp"
#include "launcher.hpp"
#include "ctrlblock.hpp"
#include "initializer.hpp"
#include "message.hpp"
#include "queue.hpp"
#include "routinglist.hpp"
#include "filterlist.hpp"
#include "processor.hpp"
#include "eventntf.hpp"
#include "listener.hpp"
#include "sshfunc.hpp"
#include "purifierproc.hpp"
#include "embedagent.hpp"

#ifdef __APPLE__
extern char **environ;
#endif

Topology::Topology(int id)
    : agentID(id), initID(-1)
{
    beMap.clear();
    weightMap.clear();
}

Topology::~Topology()
{
    beMap.clear();
    weightMap.clear();
}

void Topology::setInitID()
{
    initID = gNotifier->allocate();
}

int Topology::getInitID()
{
    return initID;
}

Message * Topology::packMsg(Message::Type type)
{   
    Packer packer;
    packer.packInt(agentID);
    packer.packInt(fanOut);
    packer.packInt(level);
    packer.packInt(height);
    packer.packStr(bePath);
    packer.packStr(agentPath);

    BEMap::iterator it;
    packer.packInt(beMap.size());    
    for (it = beMap.begin(); it != beMap.end(); ++it) {
        packer.packInt((*it).first);
        packer.packStr((*it).second);
    }

    char *bufs[1];
    int sizes[1];

    bufs[0] = packer.getPackedMsg();
    sizes[0] = packer.getPackedMsgLen();

    Message *msg = new Message();
    msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes, type);

    return msg;
}

Topology & Topology::unpackMsg(Message &msg) 
{
    int i, id, size;
    Packer packer(msg.getContentBuf());

    agentID = packer.unpackInt();
    fanOut = packer.unpackInt();
    level = packer.unpackInt();
    height = packer.unpackInt();
    bePath = packer.unpackStr();
    agentPath = packer.unpackStr();

    size = packer.unpackInt();
    for (i = 0; i < size; i++) {
        id = packer.unpackInt();
        beMap[id] = packer.unpackStr();
    }

    return *this;
}

int Topology::init()
{
    int rc;
    char *envp = NULL;
    int numItem = -1;
    char **hostlist = gCtrlBlock->getEndInfo()->fe_info.host_list;

    if ((envp = ::getenv("SCI_BACKEND_NUM")) != NULL) {
        int n = ::atoi(envp);
        if (n > 0) {
            numItem = n;
        } else {
            log_warn("Ignore invalid SCI_BACKEND_NUM value(%d)", n);
        }
    }
    if (hostlist != NULL) {
        rc = beMap.input((const char **)hostlist, numItem);
    } else {
        // check host file & num of be
        char *hostfile = gCtrlBlock->getEndInfo()->fe_info.hostfile;
        if ((envp = ::getenv("SCI_HOST_FILE")) != NULL) {
            hostfile = envp;
        }
        if (hostfile == NULL) {
            hostfile = "host.list";
        }

        rc = beMap.input(hostfile, numItem);
    }
    if (rc != SCI_SUCCESS) {
        return rc;
    }

    // check fanout
    fanOut = 32;
    if ((envp = ::getenv("SCI_DEBUG_FANOUT")) != NULL) {
        fanOut = ::atoi(envp);
    }
    
    level = 0;
    height = (int) ::ceil(::log((double)beMap.size()) / ::log((double)fanOut));

    // check be path
    if ((envp = ::getenv("SCI_BACKEND_PATH")) != NULL) {
        bePath = envp;
    } else {
        if (gCtrlBlock->getEndInfo()->fe_info.bepath != NULL) {
            bePath = gCtrlBlock->getEndInfo()->fe_info.bepath;
        } else {
            return SCI_ERR_UNKNOWN_INFO;
        }
    }

#ifdef __64BIT__
    const char *agentName = "scia64";
#else
    const char *agentName = "scia";
#endif

    envp = ::getenv("SCI_EMBED_AGENT");
    if ((envp != NULL) && (strcasecmp(envp, "yes") == 0)) {
        agentPath = bePath;
    } else if ((envp = ::getenv("SCI_AGENT_PATH")) != NULL) {
        agentPath = envp;
        agentPath += "/";
        agentPath += agentName;
    } else {
        char * tmpp = SysUtil::get_path_name(agentName);
        if (tmpp == NULL) {
            return SCI_ERR_AGENT_NOTFOUND;
        }
        agentPath = tmpp;
    }

    return SCI_SUCCESS;
}

int Topology::deploy(bool rescue)
{
    int rc;
    Launcher launcher(*this);
    char *envp = getenv("SCI_ENABLE_FAILOVER");
    nextAgentID = (agentID + 1) * fanOut - 2; // A formular to calculate the agentID of the first child

    if (rescue || 
            ((envp != NULL) && (strcasecmp(envp, "yes") == 0))) {
        launcher.setRescue(true);
    }
    rc = launcher.launch();
    if ((initID != -1) && ((gCtrlBlock->getMyRole() != CtrlBlock::BACK_AGENT))){ 
        *(int *)gNotifier->getRetVal(initID) = rc;
        gNotifier->notify(initID);   
    }

    return rc;
}

int Topology::addBE(Message *msg)
{
    assert(msg);

    Packer packer(msg->getContentBuf());
    char *host = packer.unpackStr();
    int lev = packer.unpackInt();
    int id = (int) msg->getGroup();

    // find the first child agent with weight < fanOut
    int aID = INVLIDSUCCESSORID;
    map<int, int>::iterator it = weightMap.begin();
    for (; it!=weightMap.end(); ++it) {
        int weight = (*it).second;
        if (!isFullTree(weight)) {
            aID = (*it).first;
            break;
        }
    }

    int rc = SCI_SUCCESS;
    if ((aID == INVLIDSUCCESSORID) && ((lev <= level) || (level == height))) {
        // if do not find
        Launcher launcher(*this);
        if (weightMap.size() == 0) { // if this agent does not have any child agents, launch a back end
            rc = launcher.launchBE(id, host);
        } else { // if this agent has child agent(s), launch an agent
            rc = launcher.launchAgent(id, host);
        }
    } else {
        if (aID == INVLIDSUCCESSORID)
            aID = weightMap.begin()->first;
        // otherwise delegate this command
        routingList->addBE(SCI_GROUP_ALL, aID, id);
        routingList->ucast(aID, msg);
        incWeight(aID);
    }

    if (rc == SCI_SUCCESS) {
        beMap[id] = host;
    }

    return rc;
}

int Topology::removeBE(Message *msg)
{
    assert(msg);

    int id = (int) msg->getGroup();
    if (!hasBE(id)) {
        return SCI_ERR_BACKEND_NOTFOUND;
    }

    int aID = routingList->querySuccessorId(id);
    assert(aID != INVLIDSUCCESSORID);

    routingList->removeBE(id);
    if (aID == VALIDBACKENDIDS) {
        routingList->ucast(id, msg);
        routingList->stopRouting(id);
    } else {
        routingList->ucast(aID, msg);
        decWeight(aID);
    }
    
    beMap.erase(id);
    return SCI_SUCCESS;
}

int Topology::getBENum()
{
    return beMap.size();
}

int Topology::getLevel()
{
    return level;
}

bool Topology::hasBE(int beID)
{
    if (beMap.find(beID) != beMap.end())
        return true;
    else
        return false;
}

void Topology::incWeight(int id)
{
    if (weightMap.find(id) == weightMap.end()) {
        weightMap[id] = 1;
    } else {
        weightMap[id] = weightMap[id] + 1;
    }
}

void Topology::decWeight(int id)
{
    assert(weightMap.find(id) != weightMap.end());

    weightMap[id] = weightMap[id] - 1;
    if (weightMap[id] == 0) {
        weightMap.erase(id);
    }
}

bool Topology::isFullTree(int beNum)
{ 
    if (beNum >= fanOut)
        return true;

    return false;
}

RoutingList * Topology::getRoutingList()
{
    return routingList;
}

void Topology::setRoutingList(RoutingList *rlist)
{
    routingList = rlist;
}

FilterList * Topology::getFilterList()
{
    return filterList;
}

void Topology::setFilterList(FilterList *flist)
{
    filterList = flist;
}

