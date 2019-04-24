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

#ifndef _FILTER_HPP
#define _FILTER_HPP

#include "sci.h"
#include "general.hpp"

class Message;

class Filter
{
    public:
        struct Handler {
            filter_init_hndlr    *init_hndlr;
            filter_input_hndlr   *input_hndlr;
            filter_term_hndlr    *term_hndlr;
        };
        
    private:
        sci_filter_info_t        info;
        Handler                  handler;
        void                     *param;
        void                     *file;

    public:
        Filter();
        Filter(sci_filter_info_t &filter_info);
        ~Filter();

        Message * packMsg();
        void unpackMsg(Message &msg);
        
        int load();
        int input(sci_group_t group, void *buf, int size);
        int unload();

        int getId();
        sci_filter_info_t & getInfo() { return info; }
};

#endif

