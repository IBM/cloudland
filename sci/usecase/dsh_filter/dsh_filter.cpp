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

 Classes: None

 Description: Filter.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#include "sci.h"
#include "dsh_header.hpp"
#include "dsh_aggregator.hpp"

#include <string.h>
#include <vector>
#include <map>

using namespace std;

// typedefs
typedef map<int, DshAggregator *> DSHAGGRAGATOR_MAP;
typedef vector<void *> MEM_VEC;
typedef map<int, MEM_VEC> MEM_VEC_MAP;

// global structures
DSHAGGRAGATOR_MAP gDshAggregatorMap;
MEM_VEC_MAP gMemoryMap;

extern "C" {

int filter_initialize(void **user_param);
int filter_terminate(void *user_param);
int filter_input(void *user_param, sci_group_t group, void *buf, int size);

} /* extern "C" */

bool isDataAvailable(int seq_no);
void removeData(int seq_no);
int outputData(sci_group_t group, DshMessage *msg);

int filter_initialize(void **user_param)
{
    gDshAggregatorMap.clear();
    gMemoryMap.clear();
    
    return SCI_SUCCESS;
}

int filter_terminate(void *user_param)
{
    for (int i=0; i<gDshAggregatorMap.size(); i++) {
        gDshAggregatorMap[i]->freeMemory();
    }
    gDshAggregatorMap.clear();
    gMemoryMap.clear();
    
    return SCI_SUCCESS;
}

int filter_input(void *user_param, sci_group_t group, void *buf, int size)
{
    int rc = SCI_SUCCESS;

    DshMessage *msg = new DshMessage();
    assert(msg);

    void *new_buf = malloc(size);
    memcpy(new_buf, buf, size);
    gMemoryMap[msg->getSeqNo()].push_back(new_buf);
    
    msg->unpack(new_buf);
    if (gDshAggregatorMap.find(msg->getSeqNo()) == gDshAggregatorMap.end()) {
        gDshAggregatorMap[msg->getSeqNo()] = new DshAggregator();
    }
    gDshAggregatorMap[msg->getSeqNo()]->addMsg(msg);

    if (isDataAvailable(msg->getSeqNo())) {
        DshMessage *new_msg = gDshAggregatorMap[msg->getSeqNo()]->getAggregatedMsg();
        rc = outputData(group, new_msg);
        removeData(msg->getSeqNo());
        delete new_msg;
    }

    return rc;
}

// functions
bool isDataAvailable(int seq_no)
{
    assert(gDshAggregatorMap.find(seq_no) != gDshAggregatorMap.end());
    
    int rc, ret_val;
    rc = SCI_Group_query(SCI_GROUP_ALL, GROUP_MEMBER_NUM, &ret_val);
    if (rc != SCI_SUCCESS) {
        assert(!"SCI_Query_info_ext did not work!");
    }

    if (ret_val <= gDshAggregatorMap[seq_no]->numOfBEs())
        return true;

    return false;
}

void removeData(int seq_no)
{
    assert(gDshAggregatorMap.find(seq_no) != gDshAggregatorMap.end());
    
    gDshAggregatorMap[seq_no]->freeMemory();
    delete gDshAggregatorMap[seq_no];
    gDshAggregatorMap.erase(seq_no);

    for (int i=0; i<gMemoryMap[seq_no].size(); i++) {
        free(gMemoryMap[seq_no][i]);
    }
    gMemoryMap.erase(seq_no);
}

int outputData(sci_group_t group, DshMessage *msg)
{
    assert(msg);

    void *bufs[1];
    int rc, sizes[1];

    bufs[0] = msg->pack(&sizes[0]);
    rc = SCI_Filter_upload(SCI_FILTER_NULL, group, 1, bufs, sizes);
    
    return rc;
}

