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
   01/16/12 ronglli      Add code to retrieve BE list

****************************************************************************/

#include "dgroup.hpp"
#include <stdlib.h>
#include <stdio.h>

#include "packer.hpp"
#include "group.hpp"

#include "message.hpp"

DistributedGroup::DistributedGroup(int pid)
    : parentId(pid)
{
    generalInfo.clear();

    beInfo.clear();
    successorInfo.clear();

    beListInfo.clear();
    successorListInfo.clear();

    ::pthread_mutex_init(&mtx, NULL);
}

DistributedGroup::~DistributedGroup()
{  
    GRP_MAP_MAP::iterator it = generalInfo.begin();
    for (; it != generalInfo.end(); ++it) {
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git != (*it).second.end(); ++git) {
            delete (*git).second;
        }
    }
    generalInfo.clear();

    beInfo.clear();
    successorInfo.clear();

    beListInfo.clear();
    successorListInfo.clear();

    ::pthread_mutex_destroy(&mtx);
}

void DistributedGroup::setPID(int id)
{
    parentId = id;
}

int DistributedGroup::getPID()
{
    return parentId;
}

Message * DistributedGroup::packMsg()
{
    Packer packer;
    
    packer.packInt(parentId);
    packer.packInt(generalInfo.size());
    GRP_MAP_MAP::iterator it = generalInfo.begin();
    for (; it != generalInfo.end(); ++it) {
        packer.packInt((int) (*it).first);
        packer.packInt((*it).second.size());
        
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git != (*it).second.end(); ++git) {
            packer.packInt((*git).first);
            packer.packInt((*git).second->size());

            Group::iterator ggit = (*git).second->begin();
            for (; ggit != (*git).second->end(); ggit++) {
                packer.packInt((*ggit));
            }
        }
    }

    char *bufs[1];
    int sizes[1];
    
    bufs[0] = packer.getPackedMsg();
    sizes[0] = packer.getPackedMsgLen();

    Message *msg = new Message();
    msg->build(SCI_FILTER_NULL, SCI_GROUP_ALL, 1, bufs, sizes, Message::GROUP_MERGE);
    return msg;
}

void DistributedGroup::unpackMsg(Message & msg)
{
    Packer packer(msg.getContentBuf());
    
    parentId = packer.unpackInt();
    int size1 = packer.unpackInt();
    for (int i=0; i<size1; i++) {
        sci_group_t groupId = (sci_group_t) packer.unpackInt();

        int size2 = packer.unpackInt();
        for (int j=0; j<size2; j++) {
            int childId = packer.unpackInt();

            Group *group = new Group();
            int size3 = packer.unpackInt();
            for (int k=0; k<size3; k++) {
                group->Add(packer.unpackInt());
            }

            generalInfo[groupId][childId] = group;
        }
    }
}

void DistributedGroup::create(int num_bes, int * be_list, sci_group_t group)
{
    Group total;
    for (int i=0; i<num_bes; i++) {
        total.Add(be_list[i]);
    }

    lock();

    GRP_MAP_MAP::iterator it = generalInfo.find(SCI_GROUP_ALL);
    
    GRP_MAP::iterator git = (*it).second.begin();
    for (; git!=(*it).second.end(); ++git) {
        int childHndl = (*git).first;
        Group *grp = (*git).second;

        Group diff(*grp);
        diff.Delete(total);

        Group *intersect = new Group(*grp);
        intersect->Delete(diff);
        if (!intersect->empty()) { // if not empty
            generalInfo[group][childHndl] = intersect;
        } else {
            delete intersect;
        }
    }

    reset(group);

    unlock();
}

void DistributedGroup::remove(sci_group_t group)
{
    lock();

    GRP_MAP_MAP::iterator it = generalInfo.find(group);
    
    GRP_MAP::iterator git = (*it).second.begin();
    for (; git!=(*it).second.end(); ++git) {
        delete (*git).second;
    }
    generalInfo.erase(group);

    beInfo.erase(group);
    successorInfo.erase(group);

    beListInfo.erase(group);
    successorListInfo.erase(group);

    unlock();
}

int DistributedGroup::operate(sci_group_t group1, sci_group_t group2, sci_op_t op,
    sci_group_t newgroup)
{
    bool hasMember = false;

    lock();
    
    if (op == SCI_UNION) {     
        // Add all members in group_info[group1]
        GRP_MAP::iterator git = generalInfo[group1].begin();
        for (; git!=generalInfo[group1].end(); ++git) {            
            int childHndl = (*git).first;
            Group *grp = (*git).second;
            
            Group *uni = new Group(*grp);
            if (generalInfo[group2].find(childHndl) != generalInfo[group2].end()) {
                // if found
                uni->Add(*generalInfo[group2][childHndl]);
            }
            generalInfo[newgroup][childHndl] = uni;
        }

        // Add members of group_info[group2] missing in the previous step
        git = generalInfo[group2].begin();
        for (; git!=generalInfo[group2].end(); ++git) {            
            int childHndl = (*git).first;
            Group *grp = (*git).second;
            
            if (generalInfo[group1].find(childHndl) == generalInfo[group1].end()) {
                // if not found
                Group *uni = new Group(*grp);
                generalInfo[newgroup][childHndl] = uni;
            }
        }

        // should always be true
        hasMember = true;
    } else if (op == SCI_INTERSECTION) {
        GRP_MAP_MAP::iterator it = generalInfo.find(group1);
        
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git!=(*it).second.end(); ++git) {
            int childHndl = (*git).first;
            Group *grp = (*git).second;
            
            if (generalInfo[group2].find(childHndl) == generalInfo[group2].end()) {
                // if not found
                continue;
            }
            Group diff(*grp);
            diff.Delete(*generalInfo[group2][childHndl]);

            Group *intersect = new Group(*grp);
            intersect->Delete(diff);
            if (!intersect->empty()) { // if not empty
                generalInfo[newgroup][childHndl] = intersect;
                hasMember = true;
            } else {
                delete intersect;
            }
        }
    } else if (op == SCI_DIFFERENCE) {
        GRP_MAP_MAP::iterator it = generalInfo.find(group1);
        int childHndl;
        Group *grp, *diff;
        
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git!=(*it).second.end(); ++git) {
            childHndl = (*git).first;
            grp = (*git).second;
            diff = new Group(*grp);

            if (generalInfo[group2].find(childHndl) != generalInfo[group2].end()) {
                diff->Delete(*generalInfo[group2][childHndl]);
            }
            if (!diff->empty()) { // if not empty
                generalInfo[newgroup][childHndl] = diff;
                hasMember = true;
            } else {
                delete diff;
            }
        }
    }

    int rc = SCI_SUCCESS;
    if (hasMember) {
        reset(newgroup);
    } else {
        rc = SCI_ERR_GROUP_EMPTY;
    }

    unlock();

    return rc;
}

int DistributedGroup::operateExt(sci_group_t group, int num_bes, int * be_list, 
    sci_op_t op, sci_group_t newgroup)
{
    Group total;
    for (int i=0; i<num_bes; i++) {
        total.Add(be_list[i]);
    }

    lock();
    
    bool hasMember = false;
    if (op == SCI_UNION) {
        GRP_MAP_MAP::iterator it = generalInfo.find(SCI_GROUP_ALL);
        
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git!=(*it).second.end(); ++git) {
            int childHndl = (*git).first;
            Group *grp = (*git).second;

            Group diff(*grp);
            diff.Delete(total);

            Group intersect(*grp);
            intersect.Delete(diff);

            if (generalInfo[group].find(childHndl) == generalInfo[group].end()) {
                // if not found
                if (!intersect.empty()) { // if not empty
                    Group *uni = new Group();
                    uni->Add(intersect);
                    generalInfo[newgroup][childHndl] = uni;
                }
            } else {
                if (!intersect.empty()) { // if not empty
                    Group *uni = new Group(*grp);
                    uni->Add(intersect);
                    generalInfo[newgroup][childHndl] = uni;
                } else {
                    Group *uni = new Group(*grp);
                    generalInfo[newgroup][childHndl] = uni;
                }
            }
        }

        // should always be true
        hasMember = true;
    } else if (op == SCI_INTERSECTION) {
        GRP_MAP_MAP::iterator it = generalInfo.find(group);
        
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git!=(*it).second.end(); ++git) {
            int childHndl = (*git).first;
            Group *grp = (*git).second;

            Group diff(*grp);
            diff.Delete(total);

            Group *intersect = new Group(*grp);
            intersect->Delete(diff);
            if (!intersect->empty()) { // if not empty
                generalInfo[newgroup][childHndl] = intersect;
                hasMember = true;
            } else {
                delete intersect;
            }
        }
    } else if (op == SCI_DIFFERENCE) {
        GRP_MAP_MAP::iterator it = generalInfo.find(group);

        GRP_MAP::iterator git = (*it).second.begin();
        for (; git!=(*it).second.end(); ++git) {
            int childHndl = (*git).first;
            Group *grp = (*git).second;

            Group *diff = new Group(*grp);
            diff->Delete(total);
            if (!diff->empty()) { // if not empty
                generalInfo[newgroup][childHndl] = diff;
                hasMember = true;
            } else {
                delete diff;
            }
        }
    }

    int rc = SCI_SUCCESS;
    if (hasMember) {
        reset(newgroup);
    } else {
        rc = SCI_ERR_GROUP_EMPTY;
    }

    unlock();

    return rc;
}

void DistributedGroup::initSubGroup(int successor_id, int start_be_id, int end_be_id)
{
    lock();

    // init generalInfo
    Group *grp = new Group();
    grp->Add(Range(start_be_id, end_be_id+1));
    generalInfo[SCI_GROUP_ALL][successor_id] = grp;

    // init beListInfo & beInfo
    for (int id=start_be_id; id<=end_be_id; id++) {
        beListInfo[SCI_GROUP_ALL].push_back(id);
    }
    if (beInfo.find(SCI_GROUP_ALL) == beInfo.end()) {
        beInfo[SCI_GROUP_ALL] = (end_be_id - start_be_id + 1);
    } else {
        beInfo[SCI_GROUP_ALL] += (end_be_id - start_be_id + 1);
    }

    // init successorListInfo & successorInfo
    if (successorInfo.find(SCI_GROUP_ALL) == successorInfo.end()) {
        successorInfo[SCI_GROUP_ALL] = 0;
    }
    
    if (successor_id != VALIDBACKENDIDS) {
        successorListInfo[SCI_GROUP_ALL].push_back(successor_id);
        successorInfo[SCI_GROUP_ALL] += 1;
    } else {
        for (int id=start_be_id; id<=end_be_id; id++) {
            successorListInfo[SCI_GROUP_ALL].push_back(id);
        }
        successorInfo[SCI_GROUP_ALL] += (end_be_id - start_be_id + 1);
    }

    unlock();
}

void DistributedGroup::addBE(sci_group_t group, int successor_id, int be_id)
{
    lock();

    GRP_MAP_MAP::iterator it = generalInfo.find(group);
    if (generalInfo.find(group) == generalInfo.end()) {
        Group *grp = new Group();
        grp->Add(be_id);
        generalInfo[group][successor_id] = grp;
    } else {
        GRP_MAP::iterator git = (*it).second.find(successor_id);
        if (git == (*it).second.end()) {
            Group *grp = new Group();
            grp->Add(be_id);
            ((*it).second)[successor_id] = grp;
        } else {
            (*git).second->Add(be_id);
        }
    }
    reset(group);

    unlock();
}

void DistributedGroup::removeBE(int be_id)
{
    lock();
    
    // this function will remove all empty groups after 'be_id' is removed
    GRP_MAP_MAP::iterator it = generalInfo.begin();
    vector<sci_group_t> junkGrps;
    for(; it != generalInfo.end(); ++it) {
        GRP_MAP::iterator git = (*it).second.begin();
        vector<int> junkSubGrps;
        for (; git != (*it).second.end(); ++git) {
            Group *grp = (*git).second;
            grp->Delete(be_id);
            if (grp->size() == 0) {
                junkSubGrps.push_back((*git).first);
            }
        }
        
        for (int j=0; j<(int) junkSubGrps.size(); j++) {
            if ((*it).first == SCI_GROUP_ALL) {
                if (junkSubGrps[j] >= 0) { // back end id
                    delete ((*it).second)[junkSubGrps[j]];
                    (*it).second.erase(junkSubGrps[j]);
                }
            } else {
                delete ((*it).second)[junkSubGrps[j]];
                (*it).second.erase(junkSubGrps[j]);
            }
        }

        if ((*it).second.size() == 0) {
            junkGrps.push_back((*it).first);
        }
    }
    for (int i=0; i<(int) junkGrps.size(); i++) {
        generalInfo.erase(junkGrps[i]);
        beInfo.erase(junkGrps[i]);
        successorInfo.erase(junkGrps[i]);
        beListInfo.erase(junkGrps[i]);
        successorListInfo.erase(junkGrps[i]);
    }

    resetAll();

    unlock();
}

void DistributedGroup::dropSuccessor(int successor_id)
{
    lock();
    
    // delete all group inforamtion related to 'successor_id'
    if (successor_id >= 0) {
        GRP_MAP_MAP::iterator it = generalInfo.begin();
        for (; it!=generalInfo.end(); ++it) {
            GRP_MAP::iterator git = (*it).second.find(VALIDBACKENDIDS);
            if (git != (*it).second.end()) {
                (*git).second->Delete(successor_id);
                reset((*it).first);
            }
        }
    } else {
        GRP_MAP_MAP::iterator it = generalInfo.begin();
        for (; it!=generalInfo.end(); ++it) {
            GRP_MAP::iterator git = (*it).second.find(successor_id);
            if (git != (*it).second.end()) {
                delete (*git).second;
                (*it).second.erase(successor_id);
                reset((*it).first);
            }
        }
    }

    unlock();
}

void DistributedGroup::merge(int successor_id, DistributedGroup & dgroup, bool overwrite)
{
    lock();
    
    // overwrite - whether or not use the information from 'group' if group id does not exist
    GRP_MAP_MAP::iterator dit = dgroup.generalInfo.begin();
    for(; dit != dgroup.generalInfo.end(); ++dit) {
        GRP_MAP_MAP::iterator it = generalInfo.find((*dit).first);
        if (it == generalInfo.end()) {
            // if this group id does not exist here
            if (overwrite) {
                if (successor_id >= 0) {
                    Group *group = new Group();
                    group->Add(successor_id);
                    
                    generalInfo[(*dit).first][VALIDBACKENDIDS] = group;
                } else {
                    Group *group = new Group();
                    GRP_MAP::iterator git = (*dit).second.begin();
                    for (; git != (*dit).second.end(); ++git) {
                        group->Add(*((*git).second));
                    }
                
                    generalInfo[(*dit).first][successor_id] = group;
                }
                reset((*dit).first);
            } else {
                continue;
            }
        } else {
            // or this group does exist here
            if (successor_id >= 0) {
                if ((*it).second.find(VALIDBACKENDIDS) == (*it).second.end()) {
                    ((*it).second)[VALIDBACKENDIDS] = new Group();
                }
                ((*it).second)[VALIDBACKENDIDS]->Add(successor_id);

                GRP_MAP::iterator git = (*it).second.begin();
                for (; git != (*it).second.end(); ++git) {
                    if ((*git).first != VALIDBACKENDIDS) {
                        (*git).second->Delete(successor_id);
                    }
                }
            } else {
                if ((*it).second.find(successor_id) == (*it).second.end()) {
                    ((*it).second)[successor_id] = new Group();
                }
                GRP_MAP::iterator git = (*dit).second.begin();
                for (; git != (*dit).second.end(); ++git) {
                    ((*it).second)[successor_id]->Add(*((*git).second));
                }
                
                git = (*it).second.begin();
                for (; git != (*it).second.end(); ++git) {
                    if ((*git).first != successor_id) {
                        Group *grp = ((*it).second)[successor_id];
                        (*git).second->Delete(*grp);
                    }
                }
            }
            reset((*dit).first);
        }
    }

    unlock();
}

bool DistributedGroup::isGroupExist(sci_group_t group)
{
    bool rc = false;

    lock();
    if (generalInfo.find(group) != generalInfo.end()) {
        rc = true;
    }
    unlock();

    return rc;
}

bool DistributedGroup::isSuccessorExist(int successor_id)
{
    bool rc = false;

    lock();
    GRP_MAP_MAP::iterator it = generalInfo.find(SCI_GROUP_ALL);
    if (it != generalInfo.end()) {
        GRP_MAP::iterator git = (*it).second.find(successor_id);
        if (git != (*it).second.end()) {
            rc = true;
        } else {
            git = (*it).second.find(VALIDBACKENDIDS);
            if (git != (*it).second.end()) {
                Group *grp = (*git).second;
                if (grp->HasMember(successor_id)) {
                    rc = true;
                }
            }
        }
    }
    unlock();

    return rc;
}

int DistributedGroup::numOfBE(sci_group_t group)
{
    int num = 0;

    lock();
    INT_MAP::iterator it = beInfo.find(group);
    if (it != beInfo.end()) {
        num = (*it).second;
    }
    unlock();

    return num;
}

int DistributedGroup::numOfSuccessor(sci_group_t group)
{
    int num = 0;

    lock();
    INT_MAP::iterator it = successorInfo.find(group);
    if (it != successorInfo.end()) {
        num = (*it).second;
    }
    unlock();

    return num;
}

int DistributedGroup::numOfBEOfSuccessor(int successor_id)
{
    int num = 0;
    
    lock();
    GRP_MAP_MAP::iterator it = generalInfo.find(SCI_GROUP_ALL);
    if (it != generalInfo.end()) {
        GRP_MAP::iterator git = (*it).second.find(successor_id);
        if (git != (*it).second.end()) {
            num = (*git).second->size();
        }
    }
    unlock();

    return num;
}

void DistributedGroup::retrieveBEList(sci_group_t group, int * ret_val)
{
    lock();
    INTLIST_MAP::iterator it = beListInfo.find(group);
    if (it != beListInfo.end()) {
        for (int i=0; i<(int) (*it).second.size(); i++) {
            ret_val[i] = ((*it).second)[i];
        }
    }
    unlock();
}

void DistributedGroup::retrieveSuccessorList(sci_group_t group, int * ret_val)
{
    lock();
    INTLIST_MAP::iterator it = successorListInfo.find(group);
    if (it != successorListInfo.end()) {
        for (int i=0; i<(int) (*it).second.size(); i++) {
            ret_val[i] = ((*it).second)[i];
        }
    }
    unlock();
}

void DistributedGroup::retrieveBEListOfSuccessor(int successor_id, int * ret_val)
{
    lock();
    GRP_MAP_MAP::iterator it = generalInfo.find(SCI_GROUP_ALL);
    if (it != generalInfo.end()) {
        GRP_MAP::iterator git = (*it).second.find(successor_id);
        if (git != (*it).second.end()) {
            int i = 0;
            Group::iterator ggit = (*git).second->begin();
            for (; ggit != (*git).second->end(); ggit++) {
                ret_val[i++] = (*ggit);
            }
        }
    }
    unlock();
}

int DistributedGroup::querySuccessorId(int be_id)
{
    int id = INVLIDSUCCESSORID;

    lock();
    GRP_MAP_MAP::iterator it = generalInfo.find(SCI_GROUP_ALL);
    if (it != generalInfo.end()) {
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git!=(*it).second.end(); ++git) {
            Group *grp = (*git).second;
            if (grp->HasMember(be_id)) {
                id = (*git).first;
                break;
            }
        }
    }
    unlock();

    return id;
}

void DistributedGroup::dump()
{
    printf("Here below is the distributed group information (pid %d):\n\n", parentId);

    GRP_MAP_MAP::iterator it = generalInfo.begin();
    for (; it != generalInfo.end(); ++it) {
        printf("[group id %d]:\n", (*it).first);
        printf("\tnum of back ends: %d\n", beInfo[(*it).first]);
        printf("\tnum of successors: %d\n", successorInfo[(*it).first]);
        
        GRP_MAP::iterator git = (*it).second.begin();
        for (; git != (*it).second.end(); ++git) {
            printf("\tchild id %d: ", (*git).first);

            Group::iterator ggit = (*git).second->begin();
            for (; ggit != (*git).second->end(); ggit++) {
                printf("%d ", (*ggit));
            }

            printf("\n");
        }
    }

    printf("\nEnd\n\n");
}

void DistributedGroup::reset(sci_group_t group)
{  
    int num;
    
    // reset beListInfo & beInfo
    beListInfo.erase(group);
    num = 0;
    
    GRP_MAP::iterator it = generalInfo[group].begin();
    for (; it!=generalInfo[group].end(); ++it) {
        Group *grp = (*it).second;
        Group::iterator git = grp->begin();
        for (; git!=grp->end(); git++) {
            beListInfo[group].push_back((*git));
            num++;
        }
    }

    beInfo[group] = num;

    // reset successorListInfo & successorInfo
    successorListInfo.erase(group);
    num = 0;
    
    it = generalInfo[group].begin();
    for (; it!=generalInfo[group].end(); ++it) {
        int hndl = (*it).first;
        if (hndl != VALIDBACKENDIDS) {
            successorListInfo[group].push_back(hndl);
            num++;
        } else {
            Group *grp = (*it).second;
            Group::iterator git = grp->begin();
            for (; git!=grp->end(); git++) {
                successorListInfo[group].push_back((*git));
                num++;
            }
        }
    }

    successorInfo[group] = num;
}

void DistributedGroup::resetAll()
{
    GRP_MAP_MAP::iterator it = generalInfo.begin();
    for(; it != generalInfo.end(); ++it) {
        reset((*it).first);
    }
}

void DistributedGroup::lock()
{
    ::pthread_mutex_lock(&mtx);
}

void DistributedGroup::unlock()
{
    ::pthread_mutex_unlock(&mtx);
}

