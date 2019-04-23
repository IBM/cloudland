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

 Classes: Socket, SocketException

 Description: Socket manipulation.
   
 Author: Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D153875)

****************************************************************************/

#ifndef _SOCKET_HPP
#define _SOCKET_HPP

#include <netdb.h>
#include <netinet/in.h>
#include <string>

#ifndef SOL_TCP
#define SOL_TCP IPPROTO_TCP
#endif  // SOL_TCP

#ifndef TCP_KEEPIDLE
#define TCP_KEEPIDLE TCP_KEEPALIVE
#endif  // TCP_KEEPIDLE

using namespace std;

class Socket 
{
    public:
        enum DIRECTION 
        {
            READ,
            WRITE,
            BOTH
        };
        
    private:
        int socket;
        int accSockets[32];
        int numListenfds;
        static int disableIpv6;
        static int connTimes;

    public:
        Socket(int sockfd = -1);
        ~Socket();

        int setMode(int sockfd, bool mode); // blocking or non-blocking
        int setFd(int fd);
        int setOpts(int sockfd);
        int getFd() { return socket; }
        
        int listen(int &port, char *hname = NULL);
        int iflisten(int &port, const string &ifname);
        int connect(const char *hostName, in_port_t port);
        int accept();
        int stopAccept();
        int send(const char *buf, int len);
        int recv(char *buf, int len);
        void close(DIRECTION how);

        static int getDisableIPv6();
        static void setDisableIPv6(int flag);
        static void setConnTimes(int cnt);
        int numOfListenFds();
        int getListenSockfds(int *fds);
};

class SocketException 
{
    public:
        enum CODE 
        {
            NET_ERR_SOCKET = -101,  
            NET_ERR_CONNECT = -102,    
            NET_ERR_GETADDRINFO = -103,
            NET_ERR_SEND = -104,
            NET_ERR_RECV = -105,
            NET_ERR_CLOSED = -106,
            NET_ERR_INTR = -107,
            NET_ERR_FCNTL = -108,
            NET_ERR_ACCEPT = -109,
            NET_ERR_DATA = -110,
            NET_ERR_BIND = -111
        };
        
    private:
        int       errCode;
        int       errNum;
        string    errMsg;

    public:
        SocketException(int code) throw();
        SocketException(int code, int num) throw();

        int getErrCode() const throw();
        int getErrNum() const throw();
        string & getErrMsg() throw();
};

#endif

