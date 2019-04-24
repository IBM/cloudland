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

 Classes: Message

 Description: SCI internal message
   
 Author: Tu HongJ, Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)
   01/16/12 ronglli      Add codes to detect SOCKET_BROKEN

****************************************************************************/

#ifndef _MESSAGE_HPP
#define _MESSAGE_HPP

#include "sci.h"

#include "stream.hpp"

#include "general.hpp"

const int DEFAULT_MSG_ID = (-1024 * 1024);

class MessageQueue;
class Stream;

class Message 
{
    public:
        enum Type {
            UNKNOWN = -1,
            // used for downstream messages
            CONFIG = -1001,
            COMMAND = -1002,
            FILTER_LOAD = -1003,
            FILTER_UNLOAD = -1004,
            GROUP_CREATE = -1005,
            GROUP_FREE = -1006,
            GROUP_OPERATE = -1007,
            GROUP_OPERATE_EXT = -1008,
            QUIT = -1009,
            // used for upstream messages
            DATA = -1010,
            // used for dynamic +/- backend messages
            BE_REMOVE = -1011,
            BE_ADD = -1012,
            FILTER_LIST = -1013,
            RELEASE = -1014,
            // used for error handling
            UNCLE = -2001,
            UNCLE_LIST = -2002,
            PARENT = -2003,
            ERROR_EVENT = -2004, // failure/recovery events
            GROUP_MERGE = -2005,
            GROUP_DROP = -2006,
            // used for error injection 
            SHUTDOWN = -3001,            
            KILLNODE = -3002,
            // used for polling mode
            INVALID_POLL = -4001,
            SOCKET_BROKEN = -4002,
            ERROR_DATA = -4003,
            ERROR_THREAD = -4004,
            // used for message segmentation
            SEGMENT = -5001,
            RESCUE = -6001
        };
        
    private:
        // message header
        Type            type;
        int             msgID;
        int             filterID;
        sci_group_t     group;

        // message content
        int             len;

        int             refCount;
        char            *buf;
        
    public:
        Message(Type t = UNKNOWN);
        ~Message();

        int joinSegments(Message **segments, int segnum);
        static Message *joinSegments(Message *msg, Stream *inS, MessageQueue *inQ);
        void build(int fid, sci_group_t g, int num_bufs, char *bufs[], int sizes[], Type t, 
            int id = DEFAULT_MSG_ID);
        void setRefCount(int cnt);
        int & getRefCount();
        bool isValidType(int type);

        void setID(int id) { msgID = id; }
        void setFilterID(int id) { filterID = id; }
        void setGroup(sci_group_t g) { group = g; }
        Type getType() {  return type; }
        int getID() { return msgID; }
        int getFilterID() { return filterID; }
        sci_group_t getGroup() { return group; }
        
        char * getContentBuf() { return buf; }
        int getContentLen() { return len; }

        friend Stream & operator >> (Stream &stream, Message &msg);
        friend Stream & operator << (Stream &stream, Message &msg);
};

#endif

