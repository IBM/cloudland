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

#ifndef _GROUP_HPP
#define _GROUP_HPP

#include <map>
#include <string>
#include <vector>

#include "range.hpp"

using namespace std;

typedef int  ClientId;
typedef int  GroupId;
typedef vector<Range> RangeList;
typedef RangeList::iterator range_iterator;

class Group
{
    private:
        RangeList  rangeList;

    public:
        Group();
        Group(ClientId);
        Group(Range r);
        Group(Group &group);

        bool operator == (Group &);
        bool HasMember(ClientId);
        bool HasRange(Range &);
        int  Index(ClientId);

        range_iterator Add(ClientId);
        range_iterator Add(Range);
        range_iterator Add(Range, range_iterator range);
        void           Add(Group &group);
        void           Delete(ClientId);
        void           Delete(Range);
        range_iterator Delete(Range, range_iterator range);
        void           Delete(Group &group);
        void           Clear();

        class iterator { 
            private:
                range_iterator   firstRange;
                range_iterator   lastRange;
                ClientId         clientId;
            public:
                iterator(range_iterator first, range_iterator last);
                iterator   operator ++ (int);
                ClientId  &operator *  ();
                bool       operator == (iterator it);
                bool       operator != (iterator it);
        };

        iterator begin();
        iterator end();
        bool     empty() { return rangeList.empty(); }
        size_t   size();
};

#endif

