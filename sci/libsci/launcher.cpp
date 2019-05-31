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

 Classes: Launcher

 Description: Runtime Launch the clients.
   
 Author: Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   06/21/10 tuhongj   Initial code (D153875)
   07/19/12 ronglli   Fix tree4 external launching mode support
   07/19/12 ronglli   Optimize tree launching   

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
#include <pwd.h>

#include "log.hpp"
#include "tools.hpp"
#include "packer.hpp"
#include "exception.hpp"
#include "sshfunc.hpp"
#include "ipconverter.hpp"

#include "atomic.hpp"
#include "launcher.hpp"
#include "topology.hpp"
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

#define WAIT_TIMES 60 // 60s

Launcher::Launcher(Topology &topo)
	: topology(topo), shell(""), localName(""), scidPort(SCID_PORT), mode(INTERNAL), rescue(false), embedMode(false), waitTimes(WAIT_TIMES)
{
}

Launcher::~Launcher()
{
    env.unsetAll();
    childMap.clear();
}

void Launcher::setRescue(bool res)
{
    rescue = res;
    if (res == true) {
        mode = REGISTER;
    }
}

int Launcher::initEnv()
{
    char *envp = NULL;
    string envStr;
    struct servent *serv = NULL;
    childMap.clear();

    envp = getenv("SCI_DAEMON_NAME");
    if (envp != NULL) {
        serv = getservbyname(envp, "tcp");
    } else {
        serv = getservbyname(SCID_NAME, "tcp");
    }
    if (serv != NULL) {
        scidPort = ntohs(serv->s_port);
    }
    envp = ::getenv("SCI_DEVICE_NAME");
    if (envp) {
        IPConverter converter;
        string ifname = envp;
        if (converter.getIP(ifname, true, localName) != 0) {
            localName = "";
            log_error("Launcher: invalid device name(%s). Will use the localhost", ifname.c_str());
        } else {
            env.set("SCI_DEVICE_NAME", envp);
        }
    } 
    if (localName == "") {
        char tmp[256] = {0};
        ::gethostname(tmp, sizeof(tmp));
        localName = SysUtil::get_hostname(tmp);
        if (localName == "") {
            localName = tmp;
        }
    }
    int jobKey = gCtrlBlock->getJobKey();

    env.set("SCI_JOB_KEY", jobKey);
    if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
        env.set("SCI_WORK_DIRECTORY", ::getenv("PWD"));
    } else {
        env.set("SCI_WORK_DIRECTORY", ::getenv("SCI_WORK_DIRECTORY"));
    }
    envp = ::getenv("SCI_EMBED_AGENT");
    if ((envp != NULL) && (strcasecmp(envp, "yes") == 0)) {
        embedMode = true;
        env.set("SCI_EMBED_AGENT", envp);
    }
    env.set("SCI_AGENT_PATH", topology.agentPath);
    envp = ::getenv("SCI_LIB_PATH");
    if (envp) {
        env.set("SCI_LIB_PATH", envp);
        envStr = envp;
    }
#if defined(_SCI_LINUX)
    char *library_path = "LD_LIBRARY_PATH";
#elif defined(__APPLE__)
    char *library_path = "DYLD_LIBRARY_PATH";
#else
    char *library_path = "LIBPATH";
#endif
    envp = ::getenv(library_path);
    if (envp) {
        if (envStr.length()) {
            envStr += ":";
            envStr += envp;
        } else {
            envStr = envp;
        }
    }
    if (envStr.length()) {
        env.set(library_path, envStr);
    }
    
    env.set("SCI_LOG_DIRECTORY", ::getenv("SCI_LOG_DIRECTORY"));
    env.set("SCI_LOG_LEVEL", ::getenv("SCI_LOG_LEVEL"));
    env.set("SCI_LOG_ENABLE", ::getenv("SCI_LOG_ENABLE"));
    envp = ::getenv("SCI_WAIT_TIMES");
    if (envp) {
        waitTimes = atoi(envp);
		env.set("SCI_WAIT_TIMES", waitTimes);
    }
    envp = ::getenv("SCI_REMOTE_SHELL");
    if (envp) {
        shell = envp;
		env.set("SCI_REMOTE_SHELL", shell);
    }
    env.set("SCI_USE_EXTLAUNCHER", "no");
    envp = ::getenv("SCI_USE_EXTLAUNCHER");
    if (rescue || (envp && (::strcasecmp(envp, "yes") == 0))) {
        mode = REGISTER;
        env.set("SCI_USE_EXTLAUNCHER", "yes");
    }
    env.set("SCI_ENABLE_FAILOVER", "no");
    envp = ::getenv("SCI_ENABLE_FAILOVER");
    if (envp && (::strcasecmp(envp, "yes") == 0)) {
        env.set("SCI_ENABLE_FAILOVER", "yes");
    }
    envp = ::getenv("SCI_DEBUG_TREE");
    if (envp) {
        env.set("SCI_DEBUG_TREE", envp);
    }
    envp = ::getenv("SCI_SEGMENT_SIZE");
    if (envp) {
        env.set("SCI_SEGMENT_SIZE", envp);
    }

    // add any tool specific environment variables
    if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
        char **tool_envp = NULL;
        tool_envp = gCtrlBlock->getEndInfo()->fe_info.beenvp;
        if (tool_envp) {
            while (*tool_envp) {
                // filter out SCI_ and library path.
                if (::strncmp(*tool_envp, "SCI_", 4) && 
                        ::strncmp(*tool_envp, library_path, ::strlen(library_path))) {
                    char *envstr = strdup(*tool_envp);
                    char *value = ::strchr(envstr, '=');
                    if (value) {
                        *value = '\0';
                        env.set(envstr, value+1);
                    }
                    free(envstr);
                }
                tool_envp++;
            }
        }
    } else { 
        string savedEnv = gInitializer->getEnvStr();
        int size = savedEnv.size();
        char *st =  strdup(savedEnv.c_str());
        char *key = NULL;
        char *value = NULL;
        const char *delim = ";";
        char *savePtr1 = NULL;
        key=strtok_r(st,delim,&savePtr1);
        if((key!= NULL) && (key < (st + size))) {
            if (::strncmp(key, "SCI_", 4) && 
                    ::strncmp(key, library_path, ::strlen(library_path))) {
                value = strchr(key,'=');
                if(value != NULL) {
                    (*value) = '\0';
                    if((value!= key)&&((value + 1) != NULL) && ((value+1) < (st + size))) { 
                        if((*(value+1)) == '\0')
                            env.set(key,"");
                        else
                            env.set(key,value+1);
                    }
                } else {
                    env.set(key,"");
                }
            }
            while(key = strtok_r(NULL,delim,&savePtr1)) { 
                if (::strncmp(key, "SCI_", 4) && 
                        ::strncmp(key, library_path, ::strlen(library_path))) { 
                    value = strchr(key,'=');
                    if(value != NULL) {
                        (*value) = '\0';
                        if((value != key)&&((value + 1) != NULL) && ((value+1) < (st + size))) {
                            if((*(value+1)) == '\0')
                                env.set(key,"");
                            else
                                env.set(key,value+1);
                        }
                    } else {
                        env.set(key,"");
                    }
                }
            }
        }
        free(st);
    }
    env.set("SCI_PARENT_HOSTNAME", localName);
    env.set("SCI_ENABLE_LISTENER", ::getenv("SCI_ENABLE_LISTENER"));
    env.set("SCI_PARENT_ID", topology.agentID);
    // flow control threshold
    env.set("SCI_FLOWCTL_THRESHOLD", gCtrlBlock->getFlowctlThreshold());
    
    log_debug("Launcher: env(%s)", env.getEnvString().c_str());
    return SCI_SUCCESS;
}

int Launcher::launch()
{
    int tree = 2;
    int rc = 0;

    char *envp = ::getenv("SCI_DEBUG_TREE");
    if (envp) {
        tree = ::atoi(envp);
    }

    try {
        initEnv();
        envp = getenv("SCI_ENABLE_LISTENER");
        if ((envp != NULL) && (strcasecmp(envp, "yes") == 0)) {
            gInitializer->initListener();
        }
        switch (tree) {
            case 1:
                rc = launch_tree1();
                break;
            case 2:
                rc = launch_tree2();
                break;
            case 3:
                rc = launch_tree3();
                break;
            case 4:
                rc = launch_tree4();
                break;
            default:
                return -1;
        }
        if (rc != SCI_SUCCESS) {
            return rc;
        }

        if ((shell.empty()) && (childMap.size() > 0)) {
            rc = startAll();
            if (rc != SCI_SUCCESS) {
                log_error("Launcher: startAll failed temporarily");
            }
        }
    } catch (std::bad_alloc) {
        log_error("Launcher: out of memory");
        return SCI_ERR_NO_MEM;
    } catch (SocketException &e) {
        log_error("Launcher: socket exception: %s", e.getErrMsg().c_str());
        return SCI_ERR_LAUNCH_FAILED;
    } catch (ThreadException &e) {
        log_error("Launcher: thread exception %d", e.getErrCode());
        return SCI_ERR_LAUNCH_FAILED;
    } catch (...) {
        log_error("Launcher: unknown exception");
        return SCI_ERR_LAUNCH_FAILED;
    }
    if ((mode == REGISTER) || !shell.empty()) {
        int times = 0;
        while (true) {
            if (topology.routingList->allRouted()) {
                break;
            }
            if (times >= (waitTimes * 1000000  / WAIT_INTERVAL)) {
                topology.routingList->setRecoverChildren();
                log_error("Launcher: allRouted failed temporarily");
                rc = SCI_ERR_LAUNCH_FAILED;
                break;
            }
            times++;
            SysUtil::sleep(WAIT_INTERVAL);
        }
    }
   
    /*Moved the reader start into startRouting*/ 
    /* if (rc == SCI_SUCCESS)
       rc = topology.routingList->startReaders();
    }*/
 
    return rc;
}

int Launcher::startAll()
{
    int rc = SCI_SUCCESS;
    CHILD_MAP::iterator it;
    int hndl;

    try {
        for (it = childMap.begin(); it != childMap.end(); ++it) {
            hndl = it->first;
            rc = startClient(hndl, it->second);
            if (rc != SCI_SUCCESS) {
                gCtrlBlock->setRecover(hndl);
            }
        }
    } catch (SocketException &e) {
        rc = SCI_ERR_LAUNCH_FAILED;
        log_warn("Launcher: socket exception: %s", e.getErrMsg().c_str());
    }
    childMap.clear();

    return rc;
}

int Launcher::launchBE(int beID, const char * hostname)
{
    int rc;
    char queueName[32];
    Message *flistMsg = topology.filterList->getFlistMsg();

    initEnv();
    topology.routingList->addBE(SCI_GROUP_ALL, VALIDBACKENDIDS, beID, true);
    topology.routingList->queryQueue(beID)->produce(flistMsg);

    rc = launchClient(beID, topology.bePath, hostname, mode);
    if (rc != SCI_SUCCESS) {
        topology.routingList->removeBE(beID);
    } else {
        if (mode == REGISTER) {
            int times = 0;
            while (!topology.routingList->allRouted()) {
                if (times >= (waitTimes * 1000000  / WAIT_INTERVAL))
                    return SCI_ERR_LAUNCH_FAILED;
                times++;
                SysUtil::sleep(WAIT_INTERVAL);
            }
        }
        /*Moved the reader start into startRouting*/
        //topology.routingList->startReading(beID);
    }

    return rc;
}

int Launcher::launchAgent(int beID, const char * hostname)
{
    int rc;
    
    initEnv();
    Topology *childTopo = new Topology(topology.nextAgentID--);
    childTopo->fanOut = topology.fanOut;
    childTopo->level = topology.level + 1;
    childTopo->height =  topology.height + 1;
    childTopo->bePath = topology.bePath;
    childTopo->agentPath = topology.agentPath;
    childTopo->beMap[beID] = hostname;

    topology.routingList->addBE(SCI_GROUP_ALL, childTopo->agentID, beID, true);
    MessageQueue *queue = topology.routingList->queryQueue(childTopo->agentID);

    rc = launchClient(childTopo->agentID, childTopo->agentPath, hostname);
    if (rc == SCI_SUCCESS) {
        Message *flistMsg = topology.filterList->getFlistMsg();
        Message *topoMsg = childTopo->packMsg();
        if (flistMsg != NULL) {
            incRefCount(flistMsg->getRefCount());
            queue->produce(flistMsg);
        }
        queue->produce(topoMsg);
        topology.incWeight(childTopo->agentID);
        /*Moved the reader start into startRouting*/
        //topology.routingList->startReading(childTopo->agentID);
    } else {
        topology.routingList->removeBE(beID);
    }
    delete childTopo;

    return rc;
}

int Launcher::startClient(int hndl, Stream *stream)
{
    int rc = -1;
    struct iovec sign = {0};
    psec_idbuf_desc &usertok = SSHFUNC->get_token();
    Stream *st = stream;

    if (stream != NULL) {
        env.set("SCI_CLIENT_ID", hndl);
        log_debug("Launcher: start client(%d)", hndl); 
        psec_sign_data(&sign, "%s", env.getEnvString().c_str());
        *st << usertok << env.getEnvString() << sign << endl;
        psec_free_signature(&sign);
    }
    rc = topology.routingList->startRouting(hndl, st);

    return rc;
}

int Launcher::launchClient(int ID, string &path, string host, Launcher::MODE m, bool batch, int beID)
{
    int rc = 0;
    Listener *listener = gInitializer->getListener();
    assert(!path.empty());
    if ((m == REGISTER) || !shell.empty()) {
        listener = gInitializer->initListener();
    }
    if (listener != NULL) {
        int port = 0;
        while (1) {
            port = listener->getBindPort();
            if (port > 0)
                break;
            SysUtil::sleep(WAIT_INTERVAL);
        }

        env.set("SCI_PARENT_PORT", port);
    } 
    env.set("SCI_CLIENT_ID", ID);
    log_debug("Launch client: %s: %s", host.c_str(), path.c_str());
    
    if (shell.empty()) {
        string username = gCtrlBlock->getUsername();
        try {
            int hndl = ID;
            struct iovec sign = {0};
            int jobKey = gCtrlBlock->getJobKey();
            psec_idbuf_desc &usertok = SSHFUNC->get_token();
            Stream *stream = new Stream();
            SCI_connect_hndlr *conn = NULL;

            if (gCtrlBlock->getMyRole() != CtrlBlock::AGENT)
                conn = gCtrlBlock->getEndInfo()->connect_hndlr;
            if (conn == NULL) {
                int cID = ID;
                if (embedMode && (beID >= 0)) {
                    cID = beID;
                }
                rc = psec_sign_data(&sign, "%d%d%d%s%s", m, jobKey, cID, path.c_str(), env.getEnvString().c_str());
                stream->init(host.c_str(), scidPort);
                *stream << username << usertok << sign << (int)m << jobKey << cID << path << env.getEnvString() << endl;
                psec_free_signature(&sign);
            } else {
                int sockfd = conn(host.c_str());
                if (sockfd < 0) {
                    log_error("Launcher: invalid return sockfd(%d) of the connection handler.", sockfd);
                    return SCI_ERR_LAUNCH_FAILED;
                }
                stream->init(sockfd);
            }
            if (m == REGISTER) {
                stream->stop();
                delete stream;
                childMap[hndl] = NULL;
            } else {
                if (batch) {
                    childMap[hndl] = stream;
                } else {
                    startClient(hndl, stream);
                }
            }
        } catch (SocketException &e) {
            childMap[ID] = NULL;
            rc = SCI_ERR_LAUNCH_FAILED;
            log_error("Launcher: socket exception: %s", e.getErrMsg().c_str());
        }
    } else {
        if ((strcasecmp(shell.c_str(), "rsh") == 0) || (strcasecmp(shell.c_str(), "ssh") == 0)) {
            string cmd = shell + " " + host + " -n '" + env.getExportcmd() + path + " >&- 2>&- <&- &'";
            rc = system(cmd.c_str());
        }
    }

    return rc;
}

int Launcher::launch_tree1()
{
    int rc;
    
    // this tree will have minimum agents
    int totalSize = (int) topology.beMap.size();
    char queueName[32];
    Message *flistMsg = topology.filterList->getFlistMsg();

    if (totalSize <= topology.fanOut) {
        // no need to generate agent
        BEMap::iterator it = topology.beMap.begin();

        int initID = (*it).first;
        int startID = initID;
        int endID = initID + totalSize - 1;
        topology.routingList->initSubGroup(VALIDBACKENDIDS, startID, endID);
        if (flistMsg != NULL)
            flistMsg->setRefCount(totalSize + 1);
        
        for ( ; it != topology.beMap.end(); ++it) {
            MessageQueue *queue = topology.routingList->queryQueue((*it).first);
            queue->produce(flistMsg); 
            rc = launchClient((*it).first, topology.bePath, (*it).second, mode, true);
        }

        return SCI_SUCCESS;
    }

    int stride = (int) ::ceil (::pow(double(topology.fanOut), topology.height - topology.level - 1));
    int divf;
    if ((totalSize % stride) == 0) {
        divf = totalSize / stride;
    } else {
        divf = (totalSize - totalSize%stride) / stride + 1;
    }
    int step;
    if ((totalSize % divf) == 0) {
        step = totalSize / divf;
    } else {
        step = (totalSize - (totalSize % divf)) / divf + 1;
    }
    if (flistMsg != NULL)
        flistMsg->setRefCount((totalSize + step - 1) / step + 1);
    ::srand((unsigned int) ::time(NULL));
    BEMap::iterator it = topology.beMap.begin();
    int initID = (*it).first;
    for (int i = 0; i < totalSize; i += step) {
        it = topology.beMap.begin();
        for (int j = 0; j < i; j++) {
            ++it;
        }

        // generate an agent
        Topology *childTopo = new Topology(topology.nextAgentID--);
        childTopo->fanOut  = topology.fanOut;
        childTopo->level = topology.level + 1;
        childTopo->height = topology.height;
        childTopo->bePath = topology.bePath;
        childTopo->agentPath = topology.agentPath;
        
        int min = (totalSize - i) < step ? (totalSize - i) : step;

        int startID = initID + i;
        int endID = initID + i + min - 1;
        topology.routingList->initSubGroup(childTopo->agentID, startID, endID);

        string hostname;
        
        int pos = ::rand() % min;
        for (int j = 0; j < min; j++) {
            if (pos == j) {
                hostname = (*it).second;
            }
            childTopo->beMap[(*it).first] = (*it).second;
            topology.incWeight(childTopo->agentID);
            ++it;
        }

        rc = launchClient(childTopo->agentID, topology.agentPath, hostname, INTERNAL, true);
        if (rc == SCI_SUCCESS) {
            Message *msg = childTopo->packMsg();
            MessageQueue *queue = topology.routingList->queryQueue(childTopo->agentID);
            queue->produce(flistMsg); // before topology message
            queue->produce(msg);
            delete childTopo;
        } else {
            delete childTopo;
        }
    }
   
    return SCI_SUCCESS;
}

int Launcher::launch_tree2()
{
    // this tree will have maximum agents but supposed to have better performance
    // after evaluated by HongJun
    int i, rc = SCI_SUCCESS;
    int left = 0;
    int totalSize = topology.beMap.size();
    int step;
    int size = 0;
    int out = topology.fanOut;
    Message *flistMsg = topology.filterList->getFlistMsg();
    int ref = 0;
    bool bottom = false; 

    if (totalSize == 0)
        return SCI_SUCCESS;

    bottom = ((totalSize + out - 1) / out) > 1 ? false : true;
    ref = (totalSize > out) ? out : totalSize;
    if (flistMsg != NULL)
        flistMsg->setRefCount(ref + totalSize);  // Keep it undeleted
    // launch all of my back ends
    BEMap::iterator it = topology.beMap.begin();
    int initID = it->first;
    while (size < totalSize) {
        BEMap::iterator fEnt = it; // first entry of each step
        left = totalSize - size; 
        step = (left + out - 1) / out;
        out--;
        if (step == 1) {
            topology.routingList->addBE(SCI_GROUP_ALL, VALIDBACKENDIDS, it->first, true);
            MessageQueue *queue = topology.routingList->queryQueue(it->first);
            queue->produce(flistMsg); 
            if ((gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)
                    && (fEnt == topology.beMap.begin())) {
                gCtrlBlock->getPurifierProcessor()->setInQueue(queue);
                gCtrlBlock->setMyHandle(it->first);
                gCtrlBlock->getPurifierProcessor()->start();
            } else {
                rc = launchClient(it->first, topology.bePath, it->second, mode, true);
            }
            it++;
            // do the notify when it is back agent.
            if ((size == (totalSize-1)) && bottom && (gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)) {
                int tmp_hndl = gCtrlBlock->getMyEmbedHandle();
                int tmp_initID = gCtrlBlock->getAgent(tmp_hndl)->getRoutingList()->getTopology()->getInitID();
                if ((mode == REGISTER) || !shell.empty()) {
                    int times = 0;
                    while (!gCtrlBlock->allRouted()) {
                        if (times >= (waitTimes * 1000000  / WAIT_INTERVAL)) {
                            rc = SCI_ERR_LAUNCH_FAILED;
                            *(int *)gNotifier->getRetVal(tmp_initID) = rc;
                            break;
                        }
                        times++;
                        SysUtil::sleep(WAIT_INTERVAL);
                    }
                }
                gNotifier->notify(tmp_initID);
            }
        } else {
            int auxID = it->first;
            string &hostname = it->second;
            Topology *childTopo = new Topology(topology.nextAgentID--);
            childTopo->fanOut  = topology.fanOut;
            childTopo->level = topology.level + 1;
            childTopo->height = topology.height;
            childTopo->bePath = topology.bePath;
            childTopo->agentPath = topology.agentPath;

            int startID = initID + size;
            int endID = initID + size + step - 1;
            topology.routingList->initSubGroup(childTopo->agentID, startID, endID);
            MessageQueue *queue = topology.routingList->queryQueue(childTopo->agentID);

            for (i = 0; i < step; i++) {
                childTopo->beMap[it->first] = it->second;
                topology.incWeight(childTopo->agentID);
                it++;
            }

            if ((gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)
                    && (fEnt == topology.beMap.begin())) {
                EmbedAgent *beAgent = new EmbedAgent();
                beAgent->init(childTopo->agentID, NULL, queue, gCtrlBlock->getUpQueue());
                rc = beAgent->work();
            } else {
                MODE m = INTERNAL;
                if (embedMode)
                    m = mode;
                rc = launchClient(childTopo->agentID, topology.agentPath, hostname, m, true, auxID);
            }
            if (rc == SCI_SUCCESS) {
                Message *msg = childTopo->packMsg();
                queue->produce(flistMsg); // make the filter list loaded before topology
                queue->produce(msg);
                delete childTopo;
            } else {
                delete childTopo;
            }
        }
        size += step;
    }

    return SCI_SUCCESS;
}

int Launcher::launch_tree3()
{
    // this tree will have maximum agents but supposed to have better performance
    // after evaluated by HongJun
    int i, rc = SCI_SUCCESS;
    int left = 0;
    int totalSize = topology.beMap.size();
    int step;
    int size = 0;
    int out = topology.fanOut;
    Message *flistMsg = topology.filterList->getFlistMsg();
    int ref = 0; 
    bool shift = true;
    bool bottom = false;

    if (totalSize == 0)
        return SCI_SUCCESS;

    bottom = ((totalSize + out - 1) / out) > 1 ? false : true;
    ref = (totalSize > out) ? out : totalSize;
    if (flistMsg != NULL)
        flistMsg->setRefCount(ref + totalSize);  // Keep it undeleted
    // launch all of my back ends
    BEMap::iterator it = topology.beMap.begin();
    int initID = it->first;
    while (size < totalSize) {
        BEMap::iterator fEnt = it; // first entry of each step
        left = totalSize - size; 
        step = (left + out - 1) / out;
        out--;
        if (step == 1) {
            topology.routingList->addBE(SCI_GROUP_ALL, VALIDBACKENDIDS, it->first, true);
            MessageQueue *queue = topology.routingList->queryQueue(it->first);
            queue->produce(flistMsg); 
            if ((gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)
                    && (fEnt == topology.beMap.begin())) {
                gCtrlBlock->getPurifierProcessor()->setInQueue(queue);
                gCtrlBlock->setMyHandle(it->first);
                gCtrlBlock->getPurifierProcessor()->start();
            } else {
                rc = launchClient(it->first, topology.bePath, it->second, mode, true);
            }
            it++;
            // do the notify when it is back agent.
            if ((size == (totalSize-1)) && bottom && (gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)) {
                int tmp_hndl = gCtrlBlock->getMyEmbedHandle();
                int tmp_initID = gCtrlBlock->getAgent(tmp_hndl)->getRoutingList()->getTopology()->getInitID();
                if ((mode == REGISTER) || !shell.empty()) {
                    int times = 0;
                    while (!gCtrlBlock->allRouted()) {
                        if (times >= (waitTimes * 1000000  / WAIT_INTERVAL)) {
                            rc = SCI_ERR_LAUNCH_FAILED;
                            *(int *)gNotifier->getRetVal(tmp_initID) = rc;
                            break;
                        }
                        times++;
                        SysUtil::sleep(WAIT_INTERVAL);
                    }
                }
                gNotifier->notify(tmp_initID);
            }
        } else {
            string &hostname = it->second;
            if (shift) {
                int aID = topology.agentID;
                int tmp = aID;
                BEMap::iterator fh = it;
                do {
                    tmp = aID / topology.fanOut;
                    if (tmp * topology.fanOut == aID) {
                        fh++;
                    } else {
                        break;
                    }
                    aID = tmp;
                } while (aID > 0);
                hostname = fh->second;
                shift = false;
            }
            Topology *childTopo = new Topology(topology.nextAgentID--);
            childTopo->fanOut  = topology.fanOut;
            childTopo->level = topology.level + 1;
            childTopo->height = topology.height;
            childTopo->bePath = topology.bePath;
            childTopo->agentPath = topology.agentPath;

            int startID = initID + size;
            int endID = initID + size + step - 1;
            topology.routingList->initSubGroup(childTopo->agentID, startID, endID);
            MessageQueue *queue = topology.routingList->queryQueue(childTopo->agentID);

            for (i = 0; i < step; i++) {
                childTopo->beMap[it->first] = it->second;
                topology.incWeight(childTopo->agentID);
                it++;
            }

            if ((gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)
                    && (fEnt == topology.beMap.begin())) {
                EmbedAgent *beAgent = new EmbedAgent();
                beAgent->init(childTopo->agentID, NULL, queue, gCtrlBlock->getUpQueue());
                rc = beAgent->work();
            } else {
                MODE m = INTERNAL;
                rc = launchClient(childTopo->agentID, topology.agentPath, hostname, m, true);
            }
            if (rc == SCI_SUCCESS) {
                Message *msg = childTopo->packMsg();
                queue->produce(flistMsg); // make the filter list loaded before topology
                queue->produce(msg);
                delete childTopo;
            } else {
                delete childTopo;
            }
        }
        size += step;
    }

    return SCI_SUCCESS;
}

int Launcher::launch_tree4()
{
    // this tree will have maximum agents but supposed to have better performance
    // after evaluated by HongJun
    int i, rc = SCI_SUCCESS;
    int left = 0;
    int totalSize = topology.beMap.size();
    int step = 1;
    int size = 0;
    int out = topology.fanOut;
    Message *flistMsg = topology.filterList->getFlistMsg();
    int ref = 0; 
    bool shift = true;

    if (totalSize == 0)
        return SCI_SUCCESS;

    ref = (totalSize > out) ? out : totalSize;
    if (flistMsg != NULL)
        flistMsg->setRefCount(ref + totalSize);  // Keep it undeleted
    // launch all of my back ends
    BEMap::iterator it = topology.beMap.begin();
    int initID = it->first;
    while (1) {
        BEMap::iterator fEnt = it; // first entry of each step
        if (step == 1) {
            topology.routingList->addBE(SCI_GROUP_ALL, VALIDBACKENDIDS, it->first, true);
            MessageQueue *queue = topology.routingList->queryQueue(it->first);
            queue->produce(flistMsg); 
            if ((gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)
                    && (fEnt == topology.beMap.begin())) {
                gCtrlBlock->getPurifierProcessor()->setInQueue(queue);
                gCtrlBlock->setMyHandle(it->first);
                gCtrlBlock->getPurifierProcessor()->start();
            } else {
                rc = launchClient(it->first, topology.bePath, it->second, mode, true);
            }
            it++;
        } else {
            int auxID = it->first;
            string &hostname = it->second;
            Topology *childTopo = new Topology(topology.nextAgentID--);
            childTopo->fanOut  = topology.fanOut;
            childTopo->level = topology.level + 1;
            childTopo->height = topology.height;
            childTopo->bePath = topology.bePath;
            childTopo->agentPath = topology.agentPath;

            int startID = initID + size;
            int endID = initID + size + step - 1;
            topology.routingList->initSubGroup(childTopo->agentID, startID, endID);
            MessageQueue *queue = topology.routingList->queryQueue(childTopo->agentID);

            for (i = 0; i < step; i++) {
                childTopo->beMap[it->first] = it->second;
                topology.incWeight(childTopo->agentID);
                it++;
            }

            if ((gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT)
                    && (fEnt == topology.beMap.begin())) {
                assert(!"should not come here");
            } else {
                MODE m = INTERNAL;
                if (embedMode)
                    m = mode;
                rc = launchClient(childTopo->agentID, topology.agentPath, hostname, m, true, auxID);
            }
            if (rc == SCI_SUCCESS) {
                Message *msg = childTopo->packMsg();
                queue->produce(flistMsg); // make the filter list loaded before topology
                queue->produce(msg);
                delete childTopo;
            } else {
                delete childTopo;
            }
        }
        size += step;
        if (size >= totalSize) {
            // do the notify when it is back agent.
            if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT) {
                int tmp_hndl = gCtrlBlock->getMyEmbedHandle();
                int tmp_initID = gCtrlBlock->getAgent(tmp_hndl)->getRoutingList()->getTopology()->getInitID();
                if ((mode == REGISTER) || !shell.empty()) {
                    int times = 0;
                    while (!gCtrlBlock->allRouted()) {
                        if (times >= (waitTimes * 1000000  / WAIT_INTERVAL)) {
                            rc = SCI_ERR_LAUNCH_FAILED;
                            *(int *)gNotifier->getRetVal(tmp_initID) = rc;
                            break;
                        }
                        times++;
                        SysUtil::sleep(WAIT_INTERVAL);
                    }
                }
                gNotifier->notify(tmp_initID);
            }
            break;
        }
        left = totalSize - size; 
        step = (left + out - 1) / out;
        out--;
    }

    return SCI_SUCCESS;
}

