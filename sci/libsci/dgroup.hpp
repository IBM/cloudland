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

 Classes: Distributed Group

 Description: Distributed group manipulation (Note: STL does not guarantee 
              the safety of several readers & one writer cowork together, 
              and user threads can query group information at runtime, 
              so it's necessary to add a lock to protect these read & write 
              operations).
   
 Author: Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   05/08/09 nieyy        Initial code (F156654)
   01/16/12 ronglli      Add codes to retrieve BE list

****************************************************************************/

#ifndef _DGROUP_HPP
#define _DGROUP_HPP

#include <pthread.h>

#include <map>
#include <vector>

#include "sci.h"
#include "general.hpp"

using namespace std;

class Group;
class Message;

class DistributedGroup
{
    public:
        typedef map<int, Group*> GRP_MAP;
        typedef map<sci_group_t, GRP_MAP> GRP_MAP_MAP;
        typedef map<sci_group_t, int> INT_MAP;
        typedef map<sci_group_t, vector<int> > INTLIST_MAP;
        
    private:
        int                   parentId;
        GRP_MAP_MAP           generalInfo;
        
        INT_MAP               beInfo;
        INT_MAP               successorInfo;

        INTLIST_MAP           beListInfo;
        INTLIST_MAP           successorListInfo;

        pthread_mutex_t       mtx;

    public:
        DistributedGroup(int pid);
        ~DistributedGroup();
        
        void setPID(int id);
        int getPID();

        Message * packMsg();
        void unpackMsg(Message &msg);

        // write operations
        void create(int num_bes, int *be_list, sci_group_t group);
        void remove(sci_group_t group);
        int operate(sci_group_t group1, sci_group_t group2, sci_op_t op, 
            sci_group_t newgroup);
        int operateExt(sci_group_t group, int num_bes, int *be_list, 
            sci_op_t op, sci_group_t newgroup);

        void initSubGroup(int successor_id, int start_be_id, int end_be_id);
        void addBE(sci_group_t group, int successor_id, int be_id);
        void removeBE(int be_id);
        void dropSuccessor(int successor_id);

        void merge(int successor_id, DistributedGroup &dgroup, bool overwrite);

        // read operations
        bool isGroupExist(sci_group_t group);
        bool isSuccessorExist(int successor_id);
        
        int numOfBE(sci_group_t group);
        int numOfSuccessor(sci_group_t group);
        int numOfBEOfSuccessor(int successor_id);
        
        void retrieveBEList(sci_group_t group, int *ret_val);
        void retrieveSuccessorList(sci_group_t, int *ret_val);
        void retrieveBEListOfSuccessor(int successor_id, int * ret_val);

        int querySuccessorId(int be_id);

        // for debugging purpose
        void dump();

    private:
        void reset(sci_group_t group);
        void resetAll();

        void lock();
        void unlock();
};

#endif

