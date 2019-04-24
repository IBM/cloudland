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

 Classes: IPConverter

 Description: Convert ifname to ip addresses.
   
 Author: Nicole Nie

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   12/10/09 nieyy      Initial code

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include "ipconverter.hpp"
#include <assert.h>
#include <netinet/in.h>
#include <string.h>

#include "exception.hpp"

IPConverter::IPConverter() 
    : ip_addr("")
{
    int ret = 0;
#if defined(_SCI_LINUX) || defined(__APPLE__)
    ret = ::getifaddrs(&ifa);
    if (ret != 0) {
        throw Exception(Exception::SYS_CALL);
    }
#else
    int fd, ifsize;
    if ((fd = ::socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
        ret = -1;
    }

    if (ret == 0 && ::ioctl(fd, SIOCGSIZIFCONF, (caddr_t)&ifsize) < 0) {
        ret = -1;
    }

    if (ret == 0 && (ifc.ifc_req = (struct ifreq *)::malloc(ifsize)) == NULL) {
        ret = -1;
    }

    ifc.ifc_len = ifsize;

    if (ret == 0 && ::ioctl(fd, SIOCGIFCONF, (caddr_t)&ifc) < 0) {
        ret = -1;
    }
    if (ret != 0) {
        throw Exception(Exception::SYS_CALL);
    }
#endif
}

IPConverter::~IPConverter()
{
#if defined(_SCI_LINUX) || defined(__APPLE__)
    if (ifa) {
        ::freeifaddrs(ifa);
        ifa = NULL;
    }
#else
    if (ifc.ifc_req) {
        ::free(ifc.ifc_req);
    }
#endif
}

int IPConverter::getIP(const string &ifname, bool ipv4, string &addr)
{
    if (getIP(ifname, ipv4)) {
        return -1;
    }
    addr = ip_addr;
    ip_addr = "";
    return 0;
}

int IPConverter::getIP(const string &ifname, bool ipv4, struct sockaddr_in *addr)
{
    if (!ipv4 || getIP(ifname, ipv4)) {
        return -1;
    }
    ::memcpy(addr, &sin, sizeof(struct sockaddr_in));
    return 0;
}

int IPConverter::getIP(const string &ifname, bool ipv4, struct sockaddr_in6 *addr)
{
    if (ipv4 || getIP(ifname, ipv4)) {
        return -1;
    }
    ::memcpy(addr, &sin6, sizeof(struct sockaddr_in6));
    return 0;
}

int IPConverter::getIP(const string &ifname, bool ipv4)
{
#if defined(_SCI_LINUX) || defined(__APPLE__)
    return getIPLinux(ifname, ipv4);
#else
    return getIPAIX(ifname, ipv4);
#endif
}

#if defined(_SCI_LINUX) || defined(__APPLE__)

int IPConverter::getIPLinux(const string &ifname, bool ipv4)
{
    struct ifaddrs *ifa_tmp;
    char addr[INET6_ADDRSTRLEN];
    int ret = -1;

    ifa_tmp = ifa;

    for (; ifa_tmp; ifa_tmp = ifa_tmp->ifa_next) {
        string name(ifa_tmp->ifa_name);
        if (ifname != name ||
            ifa_tmp->ifa_addr->sa_family != (ipv4 ? AF_INET : AF_INET6)) {
            continue;
        }

        if (ifa_tmp->ifa_addr) {
            if (ipv4) {
                ::memcpy(&sin, ifa_tmp->ifa_addr, sizeof(struct sockaddr_in));
                if (::inet_ntop(AF_INET, &sin.sin_addr, addr, sizeof(addr)) == NULL) {
                    continue;
                }
            } else {
                ::memcpy(&sin6, ifa_tmp->ifa_addr, sizeof(struct sockaddr_in6));
                if (::inet_ntop(AF_INET6, &sin6.sin6_addr, addr, sizeof(addr)) == NULL) {
                    continue;
                }
            }
            ip_addr = addr;
            ret = 0;
            break;
        }
    }

    return(ret);        //return 0 is okay
}

#else /* AIX */

#define REAL_SIZE(a) (((a).sa_len) > (sizeof(a)) ? ((a).sa_len) : (sizeof(a)))

int IPConverter::getIPAIX(const string &ifname, bool ipv4)
{
    char *ifr_ch, addr[INET6_ADDRSTRLEN];
    struct ifreq *ifr = ifc.ifc_req;
    struct sockaddr *sa;
    int ret = -1;
    /*
     * On AIX, actual size of ifr->ifr_addr is possibly larger than
     * size of the structure, real size is in sa_len
     */
    for (ifr_ch = (char *)ifc.ifc_req; ifr_ch < (char *)ifc.ifc_req + ifc.ifc_len;
        ifr_ch += (sizeof(ifr->ifr_name) + REAL_SIZE(ifr->ifr_addr))) {
        ifr = (struct ifreq *)ifr_ch;
        sa = (struct sockaddr *)&(ifr->ifr_addr);

        if (::strcasecmp(ifr->ifr_name, ifname.c_str()) ||
            sa->sa_family != (ipv4 ? AF_INET : AF_INET6)) {
            continue;
        }

        if (ipv4) {
            ::memcpy((void *)&sin, (void *)sa, sizeof(struct sockaddr_in));
            if (::inet_ntop(AF_INET, (void *)(&sin.sin_addr), addr, sizeof(addr)) == NULL) {
                continue;
            }
        } else {
            ::memcpy((void *)&sin6, (void *)sa, sizeof(struct sockaddr_in6));
            if (::inet_ntop(AF_INET6, (void *)(&sin6.sin6_addr), addr, sizeof(addr)) == NULL) {
                continue;
            }
        }
        ip_addr = addr;
        ret = 0;

        break;
    }

    return ret;     //return 0 is okay
}

#endif

