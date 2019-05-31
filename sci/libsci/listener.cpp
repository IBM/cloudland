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

 Classes: Listener

 Description: Listener Thread.
   
 Author: Tu HongJ, Liu Wei, Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)

****************************************************************************/

#include <assert.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/wait.h>

#include "log.hpp"
#include "stream.hpp"
#include "sshfunc.hpp"
#include "exception.hpp"

#include "atomic.hpp"
#include "listener.hpp"
#include "topology.hpp"
#include "filterlist.hpp"
#include "launcher.hpp"
#include "ctrlblock.hpp"
#include "embedagent.hpp"
#include "socket.hpp"
#include "readerproc.hpp"
#include "writerproc.hpp"
#include "routinglist.hpp"
#include "queue.hpp"
#include "tools.hpp"
#include "ipconverter.hpp"

Listener:: Listener(int hndl)
        : Thread(hndl), bindPort(-1)
{
	char tmp[256] = {0};
    socket = new Socket();
	::gethostname(tmp, sizeof(tmp));
	bindName = SysUtil::get_hostname(tmp);
    if (bindName == "") {
        bindName = tmp;
    }
}

Listener::~Listener()
{
    delete socket;
}

int Listener::init()
{
    char *envp = NULL;
    bindPort = 0;
    if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
        envp = ::getenv("SCI_LISTENER_PORT");
        if (envp) {
            bindPort = atoi(envp);
        }
    }
    envp = ::getenv("SCI_DEVICE_NAME");
    if (envp) {
        IPConverter converter;
        string ifname = envp;
        if (converter.getIP(ifname, true, bindName) == 0) {
            socket->iflisten(bindPort, ifname);
        } else {
            log_error("Listener: invalid device name(%s). Will use the localhost", ifname.c_str());
            socket->listen(bindPort, NULL);
        }
    } else {
        socket->listen(bindPort, NULL);
    }
    
    log_debug("listener binded to port %d", bindPort);

    return bindPort;
}

int Listener::stop()
{
    setState(false);
	socket->stopAccept();
    join();

    return 0;
}

void Listener::clearStream(Stream *stream)
{
    stream->stop();
    delete stream;
}

void Listener::rescue()
{
    int i, j, rc;
    int beNum = 0;
    int *beList = NULL;
    vector<int> successors;

    gCtrlBlock->getRecoverChildren(successors);

    if (successors.size() == 0) {
        return;
    }
    for (i = 0; i < successors.size(); i++) {
        RoutingList *rtList = gCtrlBlock->findAgent(successors[i])->getRoutingList();
        Topology *topo = rtList->getTopology();
        Launcher launcher(*topo);
        launcher.initEnv();
        int sID = successors[i];
        string hostname;
        if (sID >= 0) {
            string & hostname = topo->beMap[sID];
            rc = launcher.launchClient(sID, topo->bePath, hostname, Launcher::REGISTER);
            continue;
        }
        Topology *childTopo = new Topology(sID);
        childTopo->fanOut  = topo->fanOut;
        childTopo->level = topo->level + 1;
        childTopo->height = topo->height;
        childTopo->bePath = topo->bePath;
        childTopo->agentPath = topo->agentPath;
        beNum = rtList->numOfBEOfSuccessor(sID);
        beList = new int[beNum];
        rtList->retrieveBEListOfSuccessor(sID, beList);
        for (j = 0; j < beNum; j++) {
            childTopo->beMap[beList[j]] = topo->beMap[beList[j]];
        }
        for (j = beNum - 1; j >= 0; j--) {
            hostname = topo->beMap[beList[j]];
            rc = launcher.launchClient(sID, childTopo->agentPath, hostname, Launcher::INTERNAL, true);
            if (rc == SCI_SUCCESS) {
                MessageQueue *queue = rtList->queryQueue(sID);
                Message *flistMsg = topo->filterList->getFlistMsg();
                Message *topoMsg = childTopo->packMsg(Message::RESCUE);
                queue->insert(topoMsg);
                if (flistMsg != NULL) {
                    incRefCount(flistMsg->getRefCount());
                    queue->insert(flistMsg);
                }
                launcher.startClient(sID, NULL);
                break;
            }
        }
        delete [] beList;
    }
   
    return;
}

void Listener::run()
{
    int child = -1;
    int hndl = -1;
    int pID = 0;
    int key;
    int rc;
    struct iovec sign = {0};
    bool state = true;

    while (getState()) {
        child = -1;
        try {
            if (!state) {
                init();
                state = true;
            }
            child = socket->accept();
        } catch (SocketException &e) {
            log_warn("Listener: socket broken: %s", e.getErrMsg().c_str());
            if (child >= 0) {
                shutdown(child, SHUT_RDWR);
                close(child);
            }
            state = false;
            SysUtil::sleep(WAIT_INTERVAL);
            continue;
        } catch (...) {
            log_warn("Listener: unknown exception: %s");
            break;
        }
        if (child < 0) {
            rescue();
            continue;
        }
        if (!gCtrlBlock->isEnabled()) {
            log_debug("Listener: uninitialized, rejected this connection");
            break;
        }

        log_debug("Listener: accepted a child agent sockfd %d", child);

        if (gCtrlBlock->getRecoverMode()) {
            rc = gCtrlBlock->isActiveSockfd(child);
            if (rc != 0) {
                log_warn("Listener: the fd %d is already used", child);
                shutdown(child, SHUT_RDWR);
                close(child);
                log_warn("Listener: closed the fd %d", child);
                continue;
            }
        }

        Stream *stream = NULL;
        try {
            stream = new Stream();
            stream->init(child);

            *stream >> key >> hndl;
            if (key != gCtrlBlock->getJobKey()) {
                log_warn("Listener: client with invalid credential is trying to connect. key = %d, JobKey = %d, hndl = %d",
                        key, gCtrlBlock->getJobKey(), hndl);
                clearStream(stream);
                continue;
            }
            if (hndl >= 0) {
                log_debug("Listener: back end %d is connected. Parent ID is %d", hndl, pID); 
            } else {
                log_debug("Listener: agent %d is connected. Parent ID is %d", hndl, pID); 
            }

            log_debug("Listener: begin to get pID and sign");
            *stream >> pID >> sign >> endl;

            rc = psec_verify_data(&sign, "%d%d%d", key, hndl, pID);
            delete [] (char *)sign.iov_base;
            if (rc != 0) {
                log_warn("Misleading message comes");
                clearStream(stream);
                continue;
            }
            log_debug("Listener: begin to send back endl");
            *stream << endl;
            rc = gCtrlBlock->getAgent(pID)->getRoutingList()->startRouting(hndl, stream);
            if (rc < 0) {
                log_warn("Wrong client %d trying to recover", hndl);
                clearStream(stream);
                continue;
            }
        } catch (Exception &e) {
            log_error("Listener: exception %s", e.getErrMsg());
            break;
        } catch (ThreadException &e) {
            log_error("Listener: thread exception %d", e.getErrCode());
            continue; // sometimes, the writer thread is starting
        } catch (SocketException &e) {
            log_error("Listener: socket exception: %s", e.getErrMsg().c_str());
            if (stream != NULL) {
                stream->stop();
                delete stream;
                stream = NULL;
            }
            continue;
        } catch (std::bad_alloc) {
            log_error("Listener: out of memory");
            break;
        } catch (...) {
            log_error("Listener: unknown exception");
            break;
        }
    }

    setState(false);
}

