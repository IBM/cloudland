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

#include "stream.hpp"
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <arpa/inet.h>

#include "socket.hpp"

const int  BUFFER_SIZE = 16384;
const int MAX_NETWORK_SIZE = 1024 * 1024 * 8; // Max length = 8M
const char END_OF_LINE = '\n';

void endl() {}

Stream::Stream()
    : socket(NULL)
{
    buffer = new char[BUFFER_SIZE];
    cursor = buffer;
    *cursor = '\0';

    readActive = false;
    writeActive = false;
}

Stream::~Stream()
{
    if (socket != NULL)
        delete socket;
    delete [] buffer;
}

int Stream::init(const char *nodeAddr, in_port_t port)
{
    peerHost = nodeAddr;
    peerPort = port;
    if (socket != NULL) {
        socket->close(Socket::BOTH);
    } else {
        socket = new Socket();
    }
   
    if ((nodeAddr == NULL) || (port <= 0)) 
        return -1;

    socket->connect(nodeAddr, port);

    readActive = true;
    writeActive = true;
    
    return 0;
}

string & Stream::getPeerHost()
{
    return peerHost;
}

int Stream::init(int sockfd)
{
    int rc;
    struct sockaddr_storage sockaddr;
    socklen_t len = sizeof(sockaddr);
    char hbuf[NI_MAXHOST], sbuf[NI_MAXSERV];

    if (socket != NULL) {
        socket->close(Socket::BOTH);
    } else {
        socket = new Socket();
    }
    socket->setFd(sockfd);
    ::getpeername(sockfd, (struct sockaddr *)&sockaddr, &len);
    rc = ::getnameinfo((struct sockaddr *)&sockaddr, len, hbuf, sizeof(hbuf), sbuf, sizeof(sbuf), NI_NAMEREQD | NI_NUMERICSERV);
    if (rc != 0) {
        peerHost = hbuf;
        peerPort = atoi(sbuf);
    }

    readActive = true;
    writeActive = true;
    
    return 0;
}

void Stream::read(char *buf, int size)
{
    int n = 0;
    int count = size;
    char *p = buf;
    
    while (n < size) {
        count = size - n;
        n += socket->recv(p, count);
        p = buf + n;
    }
}

void Stream::write(const char *buf, int size)
{
    int len = size; // including '\0' at the end
    int count = len;
    char *p = (char *) buf;

    while (len > 0) {
        checkBuffer(len);
        count = (len - BUFFER_SIZE) > 0 ? BUFFER_SIZE : len;
        memcpy(cursor, p, count);
        cursor += count;
        p += count;
        len -= count;
    }
}

void Stream::stop()
{
    stopRead();
    stopWrite();
}

void Stream::stopRead()
{
    if (readActive) {
        readActive = false;
        socket->close(Socket::READ);
    }
}

void Stream::stopWrite()
{
    if (writeActive) {
        writeActive = false;
        socket->close(Socket::WRITE);
    }
}

bool Stream::isReadActive()
{
    return readActive;
}

bool Stream::isWriteActive()
{
    return writeActive;
}

Stream & Stream::flush()
{
    try {
        socket->send(buffer, cursor - buffer);
    } catch (SocketException &e) {
        cursor = buffer;;
        throw;
    } catch (...) {
        cursor = buffer;
        throw;
    }
    cursor = buffer;

    return *this;
}

Stream & Stream::operator >> (char &value) 
{
    read(&value, sizeof(value));

    return *this;
}

Stream & Stream::operator >> (bool &value) 
{
    read((char *)&value, sizeof(value));

    return *this;
}

Stream & Stream::operator >> (int &value)
{
    read((char *)&value, sizeof(value));
    value = ntohl(value);
    
    return *this;
}

Stream & Stream::operator >> (long &value)
{
    int low = 0;
    int high = 0;
    int nbyte;

    *this >> nbyte;
    *this >> low;
    nbyte -= sizeof(int);
    if (nbyte != 0) {
        *this >> high;
    }
    value = high << sizeof(int) | low;
    
    return *this;
}

Stream & Stream::operator >> (char *value)
{
    int len;
    *this >> len;
    read(value, len);
    
    return *this;
}

Stream & Stream::operator >> (string &value)
{
    int len;
    char *buf = NULL;
    *this >> len;
    
    if ((len < 0) || (len > MAX_NETWORK_SIZE))
        throw SocketException(SocketException::NET_ERR_DATA);

    buf = new char[len];
    read(buf, len);
    value = buf;
    delete [] buf;
   
    return *this;
}

Stream & Stream::operator >> (struct iovec &value)
{
    *this >> (long &)value.iov_len;
    if (value.iov_len > (int)MAX_NETWORK_SIZE)
        throw SocketException(SocketException::NET_ERR_DATA);

    if (value.iov_len > 0) {
        value.iov_base = new char[value.iov_len]; // must free it outside
        read((char *)value.iov_base, value.iov_len);
    }

    return *this;
}

Stream & Stream::operator >> (EndOfLine)
{
    char value;
    *this >> value;
    if (value != END_OF_LINE)
        throw SocketException(SocketException::NET_ERR_DATA);
   
    return *this;
}

Stream & Stream::operator << (char value)
{
    checkBuffer(sizeof(value));
    *cursor = value;
    cursor += sizeof(value);
   
    return *this;
}

Stream & Stream::operator << (bool value)
{
    checkBuffer(sizeof(value));
    *(bool *)cursor = value;
    cursor += sizeof(value);
   
    return *this;
}

Stream & Stream::operator << (int value)
{
    int tmp = htonl(value);
    checkBuffer(sizeof(value));
    memcpy(cursor, &tmp, sizeof(tmp));
    cursor += sizeof(value);
   
    return *this;
}

Stream & Stream::operator << (long value)
{
    *this << (int)sizeof(value);
    if (sizeof(long) > sizeof(int)) {
        int low = value << sizeof(int) >> sizeof(int);
        int high = value >> sizeof(int);
        *this << low << high;
    } else {
        *this << (int)value;
    }
   
    return *this;
}

Stream & Stream::operator << (const char *value)
{
    int len = ::strlen(value) + 1; // including '\0' at the end
    *this << len;

    int count = len;
    char *p = (char *)value;
    while (len > 0) {
        checkBuffer(len);
        count = (len - BUFFER_SIZE) > 0 ? BUFFER_SIZE : len;
        ::memcpy(cursor, p, count);
        cursor += count;
        p += count;
        len -= count;
    }
    
    return *this;
}

Stream & Stream::operator << (const string &value)
{
    *this << value.c_str();

    return *this;
}

Stream & Stream::operator << (struct iovec &value)
{
    long len = (long)value.iov_len;
    int count = len;
    char *p = (char *)value.iov_base;

    *this << len;
    while (len > 0) {
        checkBuffer(len);
        count = (len - BUFFER_SIZE) > 0 ? BUFFER_SIZE : len;
        ::memcpy(cursor, p, count);
        cursor += count;
        p += count;
        len -= count;
    }

    return *this;
}

Stream & Stream::operator << (EndOfLine)
{
    *this << END_OF_LINE;

    return flush();
}

void Stream::checkBuffer(int size)
{
    if ((cursor - buffer + size) >= BUFFER_SIZE)
        flush();
}

int Stream::getSocket() 
{
   if (socket != NULL) {
       return socket->getFd(); 
   } else {
       return -1;
   }
}
