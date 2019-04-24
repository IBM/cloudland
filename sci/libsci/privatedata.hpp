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

#ifndef _PRIVATEDATA_HPP
#define _PRIVATEDATA_HPP

#include <stdlib.h>

class FilterProcessor;
class RouterProcessor;
class RoutingList;
class FilterList;
class Topology;

class PrivateData {
    private:
        RoutingList *routingList;
        FilterList *filterList;
        FilterProcessor * filterProc;
        RouterProcessor * routerProc;

    public:
        PrivateData(RoutingList *rt = NULL, FilterList *fl = NULL, FilterProcessor *fp = NULL, RouterProcessor *rp = NULL);
        ~PrivateData();
        FilterProcessor * getFilterProcessor();
        RouterProcessor * getRouterProcessor();
        RoutingList * getRoutingList();
        FilterList * getFilterList();
        Topology * getTopology();
};

#endif
