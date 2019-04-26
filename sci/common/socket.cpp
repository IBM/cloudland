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

#include <assert.h>
#include <errno.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/tcp.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdlib.h>
#include <stdio.h>
#include <poll.h>

#include "socket.hpp"
#include "tools.hpp"
#include "ipconverter.hpp"


#define CONNECTING_TIMES 30

int Socket::disableIpv6 = 0;
int Socket::connTimes = CONNECTING_TIMES;

Socket::Socket(int sockfd)
    : socket(sockfd)
{
	int i = 0;

    numListenfds = 0;
	for (i = 0; i < NELEMS(accSockets); i++) {
        accSockets[i] = -1;
    }
}

Socket::~Socket()
{
	int i = 0;

	for (i = 0; i < NELEMS(accSockets); i++) {
        ::close(accSockets[i]);
    }
    ::close(socket);
}

int Socket::setMode(int sockfd, bool mode)
{
    int flags, newflags;

    flags = ::fcntl(sockfd, F_GETFL);
    if (flags < 0)
        throw SocketException(SocketException::NET_ERR_FCNTL, errno);

    if (mode)
        newflags = flags & ~O_NONBLOCK;
    else
        newflags = flags | O_NONBLOCK;

    if (newflags != flags) {
        if (::fcntl(sockfd, F_SETFL, newflags) < 0) {
            throw SocketException(SocketException::NET_ERR_FCNTL, errno);
        }
    }
    
    return 0;
}

int Socket::setFd(int fd)
{
    if (fd < 0) {
        throw SocketException(SocketException::NET_ERR_SOCKET, errno);
    }
    socket = fd;
    setOpts(socket);

    return 0;
}

int Socket::setOpts(int sockfd)
{
    int value = 1;

    setsockopt(sockfd, SOL_SOCKET, SO_KEEPALIVE, (void *)&value, sizeof(value));
    value = 60;
    setsockopt(sockfd, SOL_TCP, TCP_KEEPIDLE, (void*)&value, sizeof(value));
    value = 5;
    setsockopt(sockfd, SOL_TCP, TCP_KEEPINTVL, (void *)&value, sizeof(value));
    value = 3;
    setsockopt(sockfd, SOL_TCP, TCP_KEEPCNT, (void *)&value, sizeof(value));
    value = 1;
    setsockopt(sockfd, IPPROTO_TCP, TCP_NODELAY, (void *)&value, sizeof(value));

    return 0;
}

int Socket::listen(int &port, char *hname)
{
    int sockfd;
    int yes, rc;
	int accCount = 0;
    struct addrinfo hints, *host, *ressave;
    char service[NI_MAXSERV] = {0};

    ::memset(&hints, 0, sizeof(struct addrinfo));
    hints.ai_flags = (hname == NULL) ? AI_PASSIVE : 0;
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    ::sprintf(service, "%d", port);
    ::getaddrinfo(hname, service, &hints, &host);
    ressave = host;

    while (host && (accCount < NELEMS(accSockets))) {
		if ((host->ai_family != AF_INET) && (host->ai_family != AF_INET6)) {
            host = host->ai_next;
            continue;
        }

        if ((host->ai_family == AF_INET6) && (getDisableIPv6() == 1)) {
            host = host->ai_next;
            continue;
        }

        sockfd = ::socket(host->ai_family, host->ai_socktype, host->ai_protocol);
        if (sockfd >= 0) {
			yes = 1;
			rc = ::setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes));
			if (host->ai_family == AF_INET6) {
				setsockopt(sockfd, IPPROTO_IPV6, IPV6_V6ONLY, &yes, sizeof(yes));
				sockaddr_in6 *addr6 = (sockaddr_in6 *)host->ai_addr;
				if (port != 0)
					addr6->sin6_port = htons(port);
			} else {
				sockaddr_in *addr4 = (sockaddr_in *)host->ai_addr;
				if (port != 0)
					addr4->sin_port = htons(port);
			}
			setMode(sockfd, false);
            rc = ::bind(sockfd, host->ai_addr, host->ai_addrlen);
            if (rc == 0) {
                struct sockaddr_storage sockaddr;
                socklen_t len = sizeof(sockaddr);

                ::getsockname(sockfd, (struct sockaddr *)&sockaddr, &len);
                ::getnameinfo((struct sockaddr *)&sockaddr, len, NULL, 0,
                        service, sizeof(service), NI_NUMERICSERV);
                port = ::atoi(service);
				rc = ::listen(sockfd, SOMAXCONN);
                accSockets[accCount] = sockfd;
				accCount++;
            } else {
				::close(sockfd);
			}
        }
        host = host->ai_next;
    }

    if (accCount <= 0) {
        throw SocketException(SocketException::NET_ERR_BIND, errno);
    }
    ::freeaddrinfo(ressave);

    numListenfds = accCount;
    return accCount;
}

int Socket::iflisten(int & port, const string & ifname)
{
    int accCount = 0;
    char service[NI_MAXSERV] = {0};
    ::sprintf(service, "%d", port);
    
    int sockfd = ::socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) {
        throw SocketException(SocketException::NET_ERR_SOCKET, errno);
    }

    int yes = 1;
    ::setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes));

    IPConverter converter;
    struct sockaddr_in addr;
    converter.getIP(ifname, true, &addr);

    int rc = ::bind(sockfd, (struct sockaddr *)&addr, sizeof(addr));
    if (rc == 0) {
        struct sockaddr_storage sockaddr;
        socklen_t len = sizeof(sockaddr);

        ::getsockname(sockfd, (struct sockaddr *)&sockaddr, &len);
        ::getnameinfo((struct sockaddr *)&sockaddr, len, NULL, 0,
            service, sizeof(service), NI_NUMERICSERV);
        port = ::atoi(service);
    } else {
        throw SocketException(SocketException::NET_ERR_BIND, errno);
    }

    ::listen(sockfd, SOMAXCONN);
    accSockets[accCount] = sockfd;
    accCount++;
    numListenfds = accCount;

    return sockfd;
}

int Socket::connect(const char *hostName, in_port_t port)
{
    int rc = -1;
    int sockfd;
    char service[NI_MAXSERV] = {0};
    struct addrinfo *host = NULL, *ressave;
    int count = 0;
	bool connected = false;
    char *envp = NULL;

    envp = getenv("SCI_RECONN_TIME");
    if (envp != NULL) {
        int tmp = atoi(envp);
        if (tmp > 0) {
            connTimes = tmp * 10;
        }
    }
    while (count < connTimes) {
        struct addrinfo hints = {0};
        ::sprintf(service, "%d", port);
        hints.ai_family = AF_UNSPEC;
        hints.ai_socktype = SOCK_STREAM;
        hints.ai_flags = AI_NUMERICHOST | AI_NUMERICSERV;

        rc = ::getaddrinfo(hostName, service, &hints, &host);
        if (rc == EAI_NONAME) {
            hints.ai_flags = 0;
            rc = ::getaddrinfo(hostName, service, &hints, &host);
        }
        if (!host) {
            throw SocketException(SocketException::NET_ERR_GETADDRINFO, errno);
        }

		ressave = host;
        while (host) {
            if ((host->ai_family == AF_INET6) && (getDisableIPv6() == 1)) {
                host = host->ai_next;
                continue;
            }

			sockfd = ::socket(host->ai_family, host->ai_socktype, host->ai_protocol);
			if (sockfd < 0) {
				::freeaddrinfo(host);
				throw SocketException(SocketException::NET_ERR_SOCKET, errno);
			}

			rc = ::connect(sockfd, host->ai_addr, host->ai_addrlen);
			if (rc == 0) {
                rc = setOpts(sockfd);
				connected = true;
				break;
			}
			host = host->ai_next;
		}
        count++;
        ::freeaddrinfo(ressave);
		if (connected)
			break;

        ::close(sockfd);
        SysUtil::sleep(100000);
    }
    if (rc < 0) {
        throw SocketException(SocketException::NET_ERR_CONNECT, errno);
    }
    socket = sockfd;

    return sockfd;
}

int Socket::stopAccept()
{
	int i = 0;

	for (i = 0; i < NELEMS(accSockets); i++) {
        if (accSockets[i] != -1) {
            ::shutdown(accSockets[i], SHUT_RDWR);
            ::close(accSockets[i]);
            accSockets[i] = -1;
        }
	}

	return 0;
}

int Socket::accept()
{
    int client = -1;
    struct sockaddr_storage sockaddr;
    socklen_t len = sizeof(sockaddr);
	int i = 0;
	int n = 0;
    struct pollfd fds[NELEMS(accSockets)] = {0};
    int accCount = 0;

    for (i = 0; i < NELEMS(accSockets); i++){
        if (accSockets[i] == -1) {
            break;
        }
        accCount++;
        fds[i].fd = accSockets[i];
        fds[i].events = POLLIN;
    }

    n = poll(fds, accCount, 100);
    if (n > 0) {
        for (i = 0; i < accCount; i++) {
            if (fds[i].revents) {
                client = ::accept(fds[i].fd, (struct sockaddr *)&sockaddr, &len);
                if (client < 0) {
                    throw (SocketException(SocketException::NET_ERR_ACCEPT, errno));
                }
                setOpts(client);
                setMode(client, true);
                break;
            }
        }
    }

    return client;
}

int Socket::send(const char *buf, int len)
{
    int n;
    char *pos = NULL;
    int left;

    pos = (char *) buf;
    left = len;

    while (left > 0) {
        n = ::send(socket, pos, left, 0);
        if (n < 0) {
            if ((errno == EAGAIN) || (errno == EWOULDBLOCK) || (errno == EINTR)) {
                continue;
            } 
            throw (SocketException(SocketException::NET_ERR_SEND, errno));
        }
        pos += n;
        left -= n;
    }

    return 0;
}

int Socket::recv(char *buf, int len)
{
    int n;
    int left;
    char *pos;

    pos = buf;
    left    = len;

    while (left > 0) {
        n = ::recv(socket, pos, left, 0);
        if (n < 0) {
            if (errno == EINTR) {
                continue;
            }
            if ((errno == EAGAIN) || (errno == EWOULDBLOCK)) {
                break;
            }
            throw (SocketException(SocketException::NET_ERR_RECV, errno));
        } else if (n == 0) {
            throw (SocketException(SocketException::NET_ERR_CLOSED));
        }

        pos += n;
        left -= n;
    }

    return (len - left);
}

void Socket::close(Socket::DIRECTION how)
{
    if (socket < 0)
        return;

    switch (how) {
        case READ:
            ::shutdown(socket, SHUT_RD); 
            break;
        case WRITE:
            ::shutdown(socket, SHUT_WR);
            break;
        case BOTH:
            ::shutdown(socket, SHUT_RDWR);
            ::close(socket);
        default:
            break;
    }
}

int Socket::getDisableIPv6()
{
    return disableIpv6;
}

void Socket::setDisableIPv6(int flag)
{
    disableIpv6 =  flag;
}

void Socket::setConnTimes(int cnt)
{
    connTimes = cnt;
}

int Socket::numOfListenFds()
{
    return numListenfds;
}

int Socket::getListenSockfds(int *fds)
{
    int i = 0;
    for (i = 0; i < numListenfds; i++) {
        fds[i] = accSockets[i];
    }
    return i;
}

SocketException::SocketException(int code) throw()
    : errCode(code), errNum(0)
{
}

SocketException::SocketException(int code, int num) throw()
    : errCode(code), errNum(num)
{
}

int SocketException::getErrCode() const throw()
{
    return errCode;
}

int SocketException::getErrNum() const throw()
{
    return errNum;
}

string & SocketException::getErrMsg() throw()
{
    switch (errCode) {
        case NET_ERR_SOCKET:
            errMsg = "Function ::socket()";
            break;
        case NET_ERR_CONNECT:
            errMsg = "Function ::connect()";
            break;
        case NET_ERR_GETADDRINFO:
            errMsg = "Function ::getaddrinfo()";
            break;
        case NET_ERR_SEND:
            errMsg = "Function ::send()";
            break;
        case NET_ERR_RECV:
            errMsg = "Function ::recv()";
            break;
        case NET_ERR_FCNTL:
            errMsg = "Function ::fcntl()";
            break;
        case NET_ERR_CLOSED:
            errMsg = "Function ::recv() connection was closed by peer";
            break;
        case NET_ERR_DATA:
            errMsg = "Received unexpected data";
            break;
        case NET_ERR_BIND:
            errMsg = "Function ::bind()";
        default:
            errMsg = "Unknown error";
            break;
    }

    if (errNum != 0) {
        errMsg += "; system error: ";
        errMsg += ::strerror(errNum);
    }

    return errMsg;
}

