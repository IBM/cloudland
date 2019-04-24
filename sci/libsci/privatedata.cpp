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

 Classes: PrivateData

 Description: thread specific data
   
 Author: Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code ()

****************************************************************************/

#include "privatedata.hpp"
#include "filterproc.hpp"
#include "routerproc.hpp"
#include "routinglist.hpp"
#include "filterlist.hpp"

PrivateData::PrivateData(RoutingList *rt, FilterList *fl, FilterProcessor *fp, RouterProcessor *rp)
    : routingList(rt), filterList(fl), routerProc(rp), filterProc(fp)
{
}

PrivateData::~PrivateData()
{
}

FilterProcessor * PrivateData::getFilterProcessor()
{
    return filterProc;
}

RouterProcessor * PrivateData::getRouterProcessor()
{
    return routerProc;
}

RoutingList * PrivateData::getRoutingList()
{
    return routingList;
}

Topology * PrivateData::getTopology()
{
    return routingList->getTopology();
}

FilterList * PrivateData::getFilterList()
{
    return filterList;
}

