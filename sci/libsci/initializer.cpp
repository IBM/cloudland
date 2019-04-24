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

 Classes: Initializer

 Description: Prepare the environment when startup, which includes:
        1) Processor threads
        2) Message queue
        3) Others like environment variables
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   02/10/09 nieyy      Initial code (D153875)
   07/19/12 ronglli    Optimize the user query

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include <assert.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <pwd.h>

#include "sci.h"

#include "log.hpp"
#include "socket.hpp"
#include "stream.hpp"
#include "exception.hpp"
#include "sshfunc.hpp"

#include "embedagent.hpp"
#include "initializer.hpp"
#include "ctrlblock.hpp"
#include "routinglist.hpp"
#include "topology.hpp"
#include "launcher.hpp"
#include "queue.hpp"
#include "message.hpp"
#include "readerproc.hpp"
#include "writerproc.hpp"
#include "filterproc.hpp"
#include "handlerproc.hpp"
#include "routerproc.hpp"
#include "purifierproc.hpp"
#include "observer.hpp"
#include "listener.hpp"
#include "eventntf.hpp"
#include "allocator.hpp"
#include "filterlist.hpp"

#include "tools.hpp"

Initializer* Initializer::instance = NULL;

Initializer::Initializer()
    : listener(NULL), inStream(NULL), handle(-1), parentAddr(""), parentPort(-1), parentID(-1), pInfoUpdated(false)
{
    notifyID = gNotifier->allocate();
}

Initializer::~Initializer()
{
    instance = NULL;
    if (listener) {
        listener->stop();
        delete listener;
    }
    // inStream will be deleted in Writer
}

int Initializer::init()
{
    int rc = SCI_SUCCESS;
    int level = Log::INFORMATION;
    int mode = Log::DISABLE;
    char dir[MAX_PATH_LEN] = "/tmp";
    char *envp = NULL; 
    int hndl = -1;

    envp = ::getenv("SCI_LOG_DIRECTORY"); 
    if (envp != NULL) {
        ::strncpy(dir, envp, sizeof(dir));
    }
    envp = ::getenv("SCI_LOG_LEVEL"); 
    if (envp != NULL)
        level = ::atoi(envp);
   
    envp = ::getenv("SCI_LOG_ENABLE");
    if ((envp != NULL) && (strcasecmp(envp, "yes") == 0)) {
        mode = Log::ENABLE;
    }

    try {
        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            Log::getInstance()->init(dir, "fe.log", level, mode);
            log_debug("I am a front end, my handle is %d", gCtrlBlock->getMyHandle());
        } else if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT) {
            Log::getInstance()->init(dir, "scia.log", level, mode);
            log_debug("I am an agent, my handle is %d", gCtrlBlock->getMyHandle());
        } else {
            Log::getInstance()->init(dir, "be.log", level, mode);
            log_debug("I am a back end, my handle is %d", gCtrlBlock->getMyHandle());
        }

        if (SSHFUNC == NULL)
            return SCI_ERR_SSHAUTH;

        if (gCtrlBlock->getMyRole() == CtrlBlock::FRONT_END) {
            rc = initFE();
        } else if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT) {
            rc = initAgent();
        } else {
            rc = initBE();
        }
    } catch (Exception &e) {
        log_error("Initializer: exception %s", e.getErrMsg());
        return SCI_ERR_INITIALIZE_FAILED;
    } catch (ThreadException &e) {
        log_error("Initializer: thread exception %d", e.getErrCode());
        return SCI_ERR_INITIALIZE_FAILED;
    } catch (SocketException &e) {
        log_error("Initializer: socket exception: %s", e.getErrMsg().c_str());
        return SCI_ERR_INITIALIZE_FAILED;
    } catch (std::bad_alloc) {
        log_error("Initializer: out of memory");
        return SCI_ERR_INITIALIZE_FAILED;
    } catch (...) {
        log_error("Initializer: unknown exception");
        return SCI_ERR_INITIALIZE_FAILED;
    }

    return rc;
}

Listener * Initializer::getListener()
{
    return listener;
}

Stream * Initializer::getInStream()
{
    return inStream;
}

void Initializer::setInStream(Stream * s)
{
    inStream = s;
}

Listener * Initializer::initListener()
{
    if (listener)
        return listener;

    listener = new Listener(-1);
    listener->init();
    listener->start();

    return listener;
}

int Initializer::initFE()
{
    char *envp = NULL;
    handle = gCtrlBlock->getMyHandle();
    EmbedAgent *feAgent = NULL;

    Topology *topo = new Topology(handle);
    int rc = topo->init();
    if (rc != SCI_SUCCESS)
        return rc;
    gCtrlBlock->enable();

    rc = gCtrlBlock->setUsername();
    if (rc != SCI_SUCCESS)
        return rc;

    feAgent = new EmbedAgent();
    feAgent->init(-1, NULL, NULL);
    HandlerProcessor *handler = NULL;
    if (gCtrlBlock->getEndInfo()->fe_info.mode == SCI_INTERRUPT) {
        // interrupt mode
        handler = new HandlerProcessor();
        handler->setInQueue(feAgent->getUpQueue());
        handler->setSpecific(feAgent->genPrivateData());
        gCtrlBlock->setHandlerProcessor(handler);
    } else {
        // polling mode
        Observer *ob = new Observer();
        gCtrlBlock->setObserver(ob);
        gCtrlBlock->setPollQueue(feAgent->getFilterProcessor()->getOutQueue());
        feAgent->getFilterProcessor()->setObserver(ob);
    }
    if (handler) {
        handler->start();
    }
    feAgent->getRoutingList()->getTopology()->setInitID();
    rc = feAgent->work();
    gAllocator->reset();

    Message *flistMsg = gCtrlBlock->getFilterList()->packMsg(gCtrlBlock->getEndInfo()->fe_info.filter_list);
    MessageQueue *routerInQ = feAgent->getRouterInQ();
    routerInQ->produce(flistMsg);
    Message *topoMsg = topo->packMsg();
    routerInQ->produce(topoMsg);
    rc = feAgent->syncWait();
    delete topo;

    return rc;
}

int Initializer::initAgent()
{ 
    EmbedAgent *agent = NULL;
    char *envp;
    int rc;

    rc = gCtrlBlock->setUsername();
    if (rc != SCI_SUCCESS)
        return rc;

    envp = ::getenv("SCI_REMOTE_SHELL");
    if (envp != NULL) {
        rc = connectBack();
        if (rc != 0)
            return rc;
    } else {
        inStream = initStream();
    }

    agent = new EmbedAgent();
    agent->init(handle, inStream, NULL);
    gCtrlBlock->enable();
    agent->getRoutingList()->getTopology()->setInitID();
    rc = agent->work();
    rc = agent->syncWait();

    return rc;
}

Stream * Initializer::initStream()
{
    int rc;
    char *envp;
    string envStr;
    Stream *stream = new Stream();  
    struct iovec token = {0};
    struct iovec sign = {0};

    stream->init(STDIN_FILENO);
    *stream >> token >> envStr >> sign >> endl;
    setEnvStr(envStr);
    SSHFUNC->set_user_token(&token);
    rc = psec_verify_data(&sign, "%s", envStr.c_str());
    delete [] (char *)sign.iov_base;
    if (rc != 0)
        throw Exception(Exception::INVALID_SIGNATURE);

    parseEnvStr(envStr);

    // get hostname and port no from environment variable.
    envp = ::getenv("SCI_WORK_DIRECTORY");
    if (envp != NULL) {
        ::chdir(envp);
        log_debug("Change working directory to %s", envp);
    }
    envp = ::getenv("SCI_PARENT_HOSTNAME");
    if (envp != NULL) {
        parentAddr = envp;
    }
    envp = ::getenv("SCI_PARENT_PORT");
    if (envp != NULL) {
        parentPort = ::atoi(envp);
    }
	envp = ::getenv("SCI_PARENT_ID");
	if (envp != NULL) {
		parentID = ::atoi(envp);
	}

    handle = gCtrlBlock->getMyHandle();
    log_debug("My parent host is %s, parent port is %d, my ID is %d", parentAddr.c_str(), parentPort, handle);

    return stream;
}

int Initializer::parseEnvStr(string &envStr)
{  
     char *envp;
     char dir[MAX_PATH_LEN];
     int level = -1;
     int mode = Log::INVALID;
     int hndl = -1;
     int jobkey;
     
     char *st = strdup(envStr.c_str());
     int st_size=envStr.size();
     char *key = NULL;
     char *value = NULL;
     char *delim=";";
     char *saveptr=NULL;
     key=strtok_r(st,delim,&saveptr);
     if( (key != NULL) && (key < st + st_size))
     {
        value = strchr(key,'=');
        if(value != NULL)
        {
            (*value) = '\0';
            if((value != key)&&((value + 1) != NULL) && ((value+1) < (st + st_size)))
            {
              if((*(value+1)) == '\0')
                 ::setenv(key,"",1);
              else
                 ::setenv(key,value+1,1);
            }
        }
        else{
           ::setenv(key,"",1);
        }

        while(key = strtok_r(NULL,delim,&saveptr))
        {
            value = strchr(key,'=');
            if(value != NULL)
            {
                (*value) = '\0';
                if((value != key)&&((value + 1) != NULL) && ((value+1) < (st + st_size)))
                {
                  if((*(value+1)) == '\0')
                     ::setenv(key,"",1);
                  else
                     ::setenv(key,value+1,1);
                }
            }
            else{
               ::setenv(key,"",1);
            }
        }
     }
   
    free(st);

    envp = getenv("SCI_CLIENT_ID");
    assert(envp != NULL);
    hndl = atoi(envp);
    gCtrlBlock->setMyHandle(hndl);
    envp = getenv("SCI_JOB_KEY");
    assert(envp != NULL);
    jobkey = atoi(envp);
    gCtrlBlock->setJobKey(jobkey);
    envp = ::getenv("SCI_EMBED_AGENT");
    if ((envp != NULL) && (strcasecmp(envp, "yes") == 0) && (hndl < 0)) {
        gCtrlBlock->setMyRole(CtrlBlock::BACK_AGENT);
    }
    envp = ::getenv("SCI_FLOWCTL_THRESHOLD");
    if (envp != NULL) {
        long long th = ::atoll(envp);
        if (th > 0) {
            gCtrlBlock->setFlowctlThreshold(th);
        }
    }
    envp = ::getenv("SCI_LOG_LEVEL"); 
    if (envp != NULL)
        level = ::atoi(envp);
    envp = ::getenv("SCI_LOG_ENABLE");
    if (envp != NULL) {
        if (strcasecmp(envp, "yes") == 0) {
            mode = Log::ENABLE;
        } else if (strcasecmp(envp, "no") == 0) {
            mode = Log::DISABLE;
        }
    }
    envp = ::getenv("SCI_LOG_DIRECTORY");
    if (envp != NULL) {
        ::strncpy(dir, envp, MAX_PATH_LEN-1);
        dir[MAX_PATH_LEN-1] = '\0';
        log_rename(dir, level, mode);
    } else {
        log_rename(NULL, level, mode);
    }

    return 0;
}

Stream * Initializer::connectParent()
{
    int count = 0;
    int rc = -1;

    while (rc < 0) {
        try {
            rc = connectBack();
        } catch (SocketException &e) {
            if (gCtrlBlock->getMyRole() == CtrlBlock::AGENT) {
                throw (e);
            }
            unsetenv("SCI_PARENT_HOSTNAME");
            unsetenv("SCI_PARENT_PORT");
        }
    }

    return inStream;
}

int Initializer::connectBack()
{
	struct iovec sign = {0};
    int hndl = -1;
	char *envp = NULL;

    handle = gCtrlBlock->getMyHandle();
    if ((!getenv("SCI_PARENT_HOSTNAME") || !getenv("SCI_PARENT_PORT") || !getenv("SCI_PARENT_ID"))
            && (::getenv("SCI_REMOTE_SHELL") == NULL)) {
        int rc = initExtBE(handle);
        if (rc != 0)
			return rc;
    } 

	envp = ::getenv("SCI_PARENT_HOSTNAME");
	if (envp != NULL) {
		parentAddr = envp;
	}
	envp = ::getenv("SCI_PARENT_PORT");
	if (envp != NULL) {
		parentPort = ::atoi(envp);
	}
	envp = ::getenv("SCI_PARENT_ID");
	if (envp != NULL) {
		parentID = ::atoi(envp);
	}

    hndl = gCtrlBlock->getMyHandle();       // hndl may change
    handle = hndl; 
	inStream = new Stream();
	inStream->init(parentAddr.c_str(), parentPort);
	psec_sign_data(&sign, "%d%d%d", gCtrlBlock->getJobKey(), hndl, parentID);
	*inStream << gCtrlBlock->getJobKey() << hndl << parentID << sign << endl;
    *inStream >> endl;
	psec_free_signature(&sign);
    log_debug("My parent host is %s, parent port is %d, parent id is %d", parentAddr.c_str(), parentPort, parentID);

	return 0;
}

int Initializer::initBE()
{
    int rc = SCI_SUCCESS;
    char *envp = ::getenv("SCI_USE_EXTLAUNCHER");
    if (((envp != NULL) && (::strcasecmp(envp, "yes") == 0))
			|| (::getenv("SCI_REMOTE_SHELL") != NULL)) {
        rc = connectBack();
        if (rc != 0)
            return rc;
        if (handle < 0) {
            gCtrlBlock->setMyRole(CtrlBlock::BACK_AGENT);
        }
    } else {
        inStream = initStream();
    }
    gCtrlBlock->enable();

    PurifierProcessor *purifier = new PurifierProcessor(handle);
    gCtrlBlock->setPurifierProcessor(purifier);

    if (gCtrlBlock->getEndInfo()->be_info.mode == SCI_POLLING) {
        // polling mode
        MessageQueue *sysQ = new MessageQueue();
        sysQ->setName("sysQ");

        Observer *ob = new Observer();
        gCtrlBlock->setObserver(ob);
        gCtrlBlock->setPollQueue(sysQ);
        purifier->setObserver(ob);
        purifier->setOutQueue(sysQ);
    }

    if (gCtrlBlock->getMyRole() == CtrlBlock::BACK_AGENT) {
        rc = gCtrlBlock->setUsername();
        if (rc != SCI_SUCCESS)
            return rc;

        EmbedAgent *beAgent = new EmbedAgent();
        beAgent->init(handle, inStream, NULL);
        gCtrlBlock->setMyEmbedHandle(handle);
        beAgent->getRoutingList()->getTopology()->setInitID();
        rc = beAgent->work();
        rc = beAgent->syncWait();
    } else {
        MessageQueue *userQ = new MessageQueue();
        userQ->setName("userQ");
        gCtrlBlock->setUpQueue(userQ);

        purifier->setInStream(inStream);
        WriterProcessor *writer = new WriterProcessor(handle);
        // writer is a peer processor of purifier
        purifier->setPeerProcessor(writer);

        writer->setInQueue(userQ);
        writer->setOutStream(inStream);
        purifier->start();
        writer->start();
    }

    return rc;
}

int Initializer::initExtBE(int hndl)
{
    string envStr;
    char hostname[256];

    Stream stream;
    string username;
    psec_idbuf_desc &usertok = SSHFUNC->get_token();
    struct iovec sign = {0};
    struct iovec token = {0};
    int rc, tmp0, tmp1, tmp2;
    int port = SCID_PORT;
    Launcher::MODE mode = Launcher::REQUEST;
    int jobKey = gCtrlBlock->getJobKey();
    struct servent *serv = NULL;
    char *envp = getenv("SCI_DAEMON_NAME");
    char fmt[32] = {0};

    rc = gCtrlBlock->setUsername();
    if (rc != SCI_SUCCESS)
        return rc;
    username = gCtrlBlock->getUsername();

    if (envp != NULL) {
        serv = getservbyname(envp, "tcp");
    } else {
        serv = getservbyname(SCID_NAME, "tcp");
    }
    if (serv != NULL) {
        port = ntohs(serv->s_port);
    }
    rc = psec_sign_data(&sign, "%d%d%d", mode, jobKey, hndl);
    ::gethostname(hostname, sizeof(hostname));
    stream.init(hostname, port);
    stream << username.c_str() << usertok << sign << (int)mode << jobKey << hndl << endl;
    psec_free_signature(&sign);
    stream >> envStr >> token >> sign >> endl;
    setEnvStr(envStr);
    stream.stop();
    sprintf(fmt, "%%s%%%ds", token.iov_len);
    rc = psec_verify_data(&sign, fmt, envStr.c_str(), token.iov_base);
    SSHFUNC->set_user_token(&token);
    delete [] (char *)sign.iov_base;
    if (rc != 0)
        return -1; 
    parseEnvStr(envStr);
    return 0;
}

void Initializer::setParentAddr(char * addr)
{
    parentAddr = addr;
}

string & Initializer::getParentAddr()
{
    return parentAddr;
}

int Initializer::updateParentInfo(char * addr, int port)
{
    while (pInfoUpdated == true) {
        if ((gCtrlBlock->getTermState()) || (!gCtrlBlock->getRecoverMode()) || (!gCtrlBlock->getParentInfoWaitState())) {
            return SCI_ERR_INVALID_CALLER;
        }
        SysUtil::sleep(WAIT_INTERVAL);
    }
    parentAddr = addr;
    parentPort = port;
    pInfoUpdated = true;
    gNotifier->freeze(notifyID);
    notifyID = gNotifier->allocate();

    return SCI_SUCCESS;
}

void Initializer::setParentPort(int port)
{
    parentPort = port;
}

int Initializer::getParentPort()
{
    return parentPort;
}

int Initializer::getParentID()
{
    return parentID;
}

int Initializer::getOrgHandle()
{
    return handle;
}


void Initializer::setEnvStr(string env)
{
    initEnv = env;
    return;
}

string Initializer::getEnvStr()
{
   return initEnv;
}
