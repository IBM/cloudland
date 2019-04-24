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

#ifndef _IPCONVERTER_HPP
#define _IPCONVERTER_HPP

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#if defined(_SCI_LINUX) || defined(__APPLE__)
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <ifaddrs.h>
#else /* AIX */
#include <sys/ioctl.h>
#include <net/if.h>
#include <net/if_arp.h>
#include <net/netopt.h>
#include <arpa/inet.h>
#endif

#include <string>
using namespace std;

class IPConverter
{
    private:
        string ip_addr;
        struct sockaddr_in sin;
        struct sockaddr_in6 sin6;

#if defined(_SCI_LINUX) || defined(__APPLE__)
        struct ifaddrs *ifa;
#else /* AIX */
        struct ifconf ifc;
#endif

    public:
        IPConverter();
        ~IPConverter();
        
        int getIP(const string &ifname, bool ipv4, string &addr);
        int getIP(const string &ifname, bool ipv4, struct sockaddr_in *addr);
        int getIP(const string &ifname, bool ipv4, struct sockaddr_in6 *addr);

    private:
        int getIP(const string &ifname, bool ipv4);

#if defined(_SCI_LINUX) || defined(__APPLE__)
        int getIPLinux(const string &ifname, bool ipv4);
#else /* AIX */
        #define REAL_SIZE(a) (((a).sa_len) > (sizeof(a)) ? ((a).sa_len) : (sizeof(a)))
        int getIPAIX(const string &ifname, bool ipv4);
#endif       
};

#endif

