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

 Classes: Stream

 Description: Data stream processing.
   
 Author: Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)

****************************************************************************/

#ifndef _STREAM_HPP
#define _STREAM_HPP

#include <netdb.h>
#include <netinet/in.h>
#include <sys/uio.h>
#include <string>

#include "socket.hpp"


using namespace std;

typedef void (EndOfLine)();
void endl();

class Stream 
{
    private:
        Socket       *socket;
        char         *buffer;
        char         *cursor;

        bool         readActive;
        bool         writeActive;

        string       peerHost;
        in_port_t    peerPort;

    public:
        Stream();
        ~Stream();

        int init(const char *nodeAddr, in_port_t port);
        int init(int sockfd);
        int getSocket(); 
        string & getPeerHost();

        void read(char *buf, int size);
        void write(const char *buf, int size);
        void stop();
        void stopRead();
        void stopWrite();
        bool isReadActive();
        bool isWriteActive();
        Stream & flush();

        Stream & operator >> (char &value);
        Stream & operator >> (bool &value);
        Stream & operator >> (int &value);
        Stream & operator >> (long &value);
        Stream & operator >> (char *value);
        Stream & operator >> (string &value);
        Stream & operator >> (struct iovec &value);
        Stream & operator >> (EndOfLine);
        
        Stream & operator << (char value);
        Stream & operator << (bool value);
        Stream & operator << (int value);
        Stream & operator << (long value);
        Stream & operator << (const char *value);
        Stream & operator << (const string &value);
        Stream & operator << (struct iovec &value);
        Stream & operator << (EndOfLine);

    private:
        void checkBuffer(int size);
};

#endif

