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

 Classes: BEMap, Topology, Launcher

 Description: Runtime topology manipulation.
   
 Author: Nicole Nie, Liu Wei, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code (D153875)

****************************************************************************/

#ifndef _TOPOLOGY_HPP
#define _TOPOLOGY_HPP

#include <map>
#include <string>

using namespace std;

#include "sci.h"
#include "general.hpp"
#include "stream.hpp"
#include "envvar.hpp"
#include "bemap.hpp"
#include "message.hpp"

class Launcher;
class RoutingList;
class FilterList;
class Listener;

class Topology 
{        
    private:
        // primary members
        int                  initID;
        int                  agentID;
        int                  fanOut;
        int                  level;
        int                  height;
        string               bePath;
        string               agentPath;
        BEMap                beMap;

        // other members
        int                  nextAgentID;
        RoutingList         *routingList;
        FilterList          *filterList;

        // weight factors
        map<int, int> weightMap;

    public: 
        Topology(int id);
        ~Topology();

        Message * packMsg(Message::Type type = Message::CONFIG);
        Topology & unpackMsg(Message &msg);

        int init(); // only called by FE
        int deploy(bool rescue = false);
        int getInitID();
        void setInitID();

        int addBE(Message *msg);
        int removeBE(Message *msg);

        bool hasBE(int beID);
        int getBENum();
        int getLevel();
        int getFanout() { return fanOut; }

        void incWeight(int id);
        void decWeight(int id);
        RoutingList * getRoutingList();
        void setRoutingList(RoutingList *rlist);
        FilterList * getFilterList();
        void setFilterList(FilterList *flist);

        friend class Launcher;
        friend class Listener;

    private:
        bool isFullTree(int beNum);
};

const int MAX_FD = 256;

#endif

