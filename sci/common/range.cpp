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

 Classes: Range

 Description: Range manipulation.
   
 Author: Hanhong Xue

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 hxue         Initial code (D153875)

****************************************************************************/

#include "range.hpp"
#include <assert.h>
#include <stdio.h>

Range::Range()
    : first(0), last(0)
{
}

Range::Range(const Range &r)
    : first(r.first), last(r.last)
{
}

Range::Range(int _first, int _last)
    : first(_first), last(_last)
{
}

bool Range::operator == (Range &r)
{
    return first == r.first && last == r.last;
}

bool Range::operator != (Range &r)
{
    return first != r.first || last != r.last;
}

bool Range::Intersects(Range r)
{
    return (((first <= r.first) && (r.first < last))
             || ((first < r.last) && (r.last <= last))
             || this->Contains(r) || r.Contains(*this));
}

bool Range::Touches(Range r)
{
    return (((first <= r.first) && (r.first <= last))
             || ((first <= r.last) && (r.last <= last))
             || this->Contains(r) || r.Contains(*this));
}

bool Range::IsBefore(Range r)
{
    return last <= r.first;
}

bool Range::IsAfter(Range r)
{
    return r.last <= first;
}

bool Range::Contains(Range r)
{
    return (first <= r.first && r.last <= last);
}

bool Range::Splits(Range r)
{
    return (r.first < first && last < r.last);
}

Range Range::Union(Range r)
{
    assert(this->Touches(r));
    return Range(min(first, r.first), max(last, r.last));
}

Range Range::Difference(Range r)
{
    assert(!r.Splits(*this));
    if (this->Intersects(r)) {
        if (first < r.first) {
            return Range(first, r.first);
        } else {
            assert(r.last < last);
            return Range(r.last, last);
        }
    } else
        return *this;  // unchanged
}

