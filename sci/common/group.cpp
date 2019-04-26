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

 Classes: Group

 Description: Group manipulation.
   
 Author: Hanhong Xue

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 hxue         Initial code (D153875)

****************************************************************************/

#include "group.hpp"
#include <assert.h>
#include <ctype.h>
#include <stdio.h>

Group::Group()
{
}

Group::Group(ClientId clientId)
{
    Add(Range(clientId, clientId+1));
}

Group::Group(Range r)
{
    Add(r);
}

Group::Group(Group &group)
{
    Add(group);
}

bool Group::operator == (Group &g)
{
    if (rangeList.size() != g.rangeList.size())
        return false;
    for (int i = 0; i < (int)rangeList.size(); i++)
        if (rangeList[i] != g.rangeList[i])
            return false;
    return true;
}

bool Group::HasMember(ClientId clientId)
{
    range_iterator it;
    for (it=rangeList.begin(); it!=rangeList.end(); ++it) {
        if (clientId >= (*it).end()) {
            continue;
        } else if (clientId >= (*it).begin()) {
            return true;
        }
    }

    return false;
}

bool Group::HasRange(Range & r)
{    
    for(Range::iterator i=r.begin(); i< r.end(); i++ )
        if (this->HasMember(i))
            continue;
        else 
            return false;    
    return true;
}

int Group::Index(ClientId clientId)
{
    int index = 0;
    range_iterator it = rangeList.begin();
    for ( ; it!=rangeList.end(); ++it) {
        if (clientId < (*it).begin()) {
            break;
        } else if (clientId < (*it).end()) {
            return index + clientId - (*it).begin();
        } else {
            index += (*it).end() - (*it).begin();
        }
    }
       
    assert(!"Should have found index");
    return index;
}

range_iterator Group::Add(ClientId clientId)
{
    return Add(Range(clientId, clientId+1));
}

range_iterator Group::Add(Range r)
{
    return Add(r, rangeList.begin());
}

range_iterator Group::Add(Range r, range_iterator range)
{
    // shortcut for a higher range
    if (!rangeList.empty() && !r.Touches(*rangeList.rbegin())
            && !r.IsBefore(*rangeList.rbegin()))
        return rangeList.insert(rangeList.end(), r);

    for (; range != rangeList.end(); range++)
        if (r.Touches(*range) || r.IsBefore(*range))
            break;
    if (range == rangeList.end()) {
        return rangeList.insert(range, r);
    } else if (r.Touches(*range)) {
        *range = r.Union(*range);
        while (range + 1 != rangeList.end() && range->Touches(*(range+1))) {
            *range = range->Union(*(range+1));
            rangeList.erase(range+1);
        }
        return range;
    } else {
        return rangeList.insert(range, r);
    }
}

void Group::Add(Group &group)
{
    range_iterator pos = rangeList.begin();
    range_iterator it = group.rangeList.begin();
    for ( ; it!=group.rangeList.end(); ++it) {
        pos = Add((*it), pos);
    }
}

void Group::Delete(ClientId clientId)
{
    Delete(Range(clientId, clientId+1));
}

void Group::Delete(Range r)
{
    Delete(r, rangeList.begin());
}

range_iterator Group::Delete(Range r, range_iterator range)
{
    for (; range != rangeList.end(); range++)
        if (r.Intersects(*range) || r.IsBefore(*range))
            break;
    if (range == rangeList.end())
        return range;
    else if (r.Intersects(*range)) {
        range_iterator savedRange = range;
        if (r.Splits(*range)) {
            Range newRange(r.end(), range->end());
            *range = Range(range->begin(), r.begin());
            range = rangeList.insert(range+1, newRange);
            return range;
        } else {
            while (range != rangeList.end()) {
                if (r.Contains(*range)) {
                    rangeList.erase(range);
                } else {
                    *range = range->Difference(r);
                    range++;
                }
            }
        }
        return savedRange;
    } else
        return range;
}

void Group::Delete(Group &group)
{
    range_iterator pos = rangeList.begin();
    range_iterator it = group.rangeList.begin();
    for ( ; it!=group.rangeList.end(); ++it) {
        pos = Delete((*it), pos);
    }
}

void Group::Clear()
{
    rangeList.clear();
}

Group::iterator Group::begin()
{
    return iterator(rangeList.begin(), rangeList.end());
}

Group::iterator Group::end()
{
    return iterator(rangeList.end(), rangeList.end()); 
}

size_t Group::size()
{
    int len = 0;
    range_iterator it = rangeList.begin();
    for ( ; it!=rangeList.end(); ++it) {
        len += (*it).end() - (*it).begin();
    }

    return len;
}

/////////////////////////////////////////

Group::iterator::iterator(range_iterator first, range_iterator last)
    : firstRange(first), lastRange(last)
{
    if (firstRange != lastRange)
        clientId = firstRange->begin();
    else
        clientId = -1;
}

ClientId & Group::iterator::operator *()
{
    return clientId;
}

Group::iterator Group::iterator::operator ++ (int)
{
    if (++clientId == firstRange->end()) {
        firstRange++;
        if (firstRange != lastRange)
            clientId = firstRange->begin();
        else
            clientId = -1;
    }
    return *this;
}

bool Group::iterator::operator == (Group::iterator it)
{
    return firstRange == it.firstRange && clientId == it.clientId;
}

bool Group::iterator::operator != (Group::iterator it)
{
    return !(*this == it);
}

