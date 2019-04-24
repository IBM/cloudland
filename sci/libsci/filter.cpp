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

 Classes: Filter

 Description: Filter manipulation.
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code (D153875)

****************************************************************************/
#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include <stdlib.h>
#include <math.h>
#include <assert.h>
#include <ctype.h>
#include <string.h>
#include <dlfcn.h>

#include "packer.hpp"
#include "log.hpp"

#include "message.hpp"
#include "queue.hpp"
#include "filterproc.hpp"
#include "filter.hpp"

Filter::Filter()
{
    info.filter_id = 0;
    info.so_file = NULL;
    
    handler.init_hndlr = NULL;
    handler.input_hndlr = NULL;
    handler.term_hndlr = NULL;

    param = NULL;
    file = NULL;
}

Filter::Filter(sci_filter_info_t & filter_info)
{
    info.filter_id = filter_info.filter_id;
    info.so_file = filter_info.so_file;
    
    handler.init_hndlr = NULL;
    handler.input_hndlr = NULL;
    handler.term_hndlr = NULL;

    param = NULL;
    file = NULL;
}

Filter::~Filter()
{
    if (file) {
        ::dlclose(file);
        file = NULL;
    }
}

Message * Filter::packMsg()
{
    Packer packer;
    packer.packInt(info.filter_id);
    packer.packStr(info.so_file);

    char *bufs[1];
    int sizes[1];
    
    bufs[0] = packer.getPackedMsg();
    sizes[0] = packer.getPackedMsgLen();

    Message *msg = new Message();
    msg->build(info.filter_id, SCI_GROUP_ALL, 1, bufs, sizes, Message::FILTER_LOAD);
    return msg;
}

void Filter::unpackMsg(Message &msg) 
{
    Packer packer(msg.getContentBuf());

    info.filter_id = packer.unpackInt();
    info.so_file = packer.unpackStr();
}

int Filter::load()
{
#if defined(_SCI_LINUX)
    file = ::dlopen(info.so_file, RTLD_NOW | RTLD_LOCAL);
#elif defined(__APPLE__)
    file = ::dlopen(info.so_file, RTLD_NOW | RTLD_LOCAL);
#else // aix
    file = ::dlopen(info.so_file, RTLD_NOW | RTLD_LOCAL | RTLD_MEMBER);
#endif
    if (file == NULL) {
        log_error("Loading filter failed %s", ::dlerror());
        return SCI_ERR_INVALID_FILTER;
    }
    
    handler.init_hndlr = (filter_init_hndlr *) ::dlsym(file, "filter_initialize");
    if (handler.init_hndlr == NULL) {
        log_error("Loading filter failed %s", ::dlerror());
        return SCI_ERR_INVALID_FILTER;
    }
    handler.input_hndlr = (filter_input_hndlr *) ::dlsym(file, "filter_input");
    if (handler.input_hndlr == NULL) {
        log_error("Loading filter failed %s", ::dlerror());
        return SCI_ERR_INVALID_FILTER;
    }
    handler.term_hndlr = (filter_term_hndlr *) ::dlsym(file, "filter_terminate");
    if (handler.term_hndlr == NULL) {
        log_error("Loading filter failed %s", ::dlerror());
        return SCI_ERR_INVALID_FILTER;
    }

    return handler.init_hndlr(&param);
}

int Filter::input(sci_group_t group, void *buf, int size)
{
    return handler.input_hndlr(param, group, buf, size);
}

int Filter::unload()
{
    int rc = handler.term_hndlr(param);

    // close library handle
    ::dlclose(file);
    file = NULL;

    return rc;
}

int Filter::getId()
{
    return info.filter_id;
}

