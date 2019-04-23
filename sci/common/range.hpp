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

#ifndef _RANGE_HPP
#define _RANGE_HPP

#include <string>

using namespace std;

#define RANGE_SEPARATOR ":"

class Range
{
    private:
        int   first;
        int   last;
    public:
        Range();
        Range(const Range &r);
        Range(int _first, int _last);

        typedef int iterator;
        iterator begin() { return first; }
        iterator end()   { return last; }

        bool  operator == (Range &r);
        bool  operator != (Range &r);
        bool  Intersects(Range r);
        bool  Touches(Range r);
        bool  Splits(Range r);
        bool  Contains(Range r);
        bool  IsBefore(Range r);
        bool  IsAfter(Range r);
        Range Union(Range r);
        Range Difference(Range r);
};

#endif

