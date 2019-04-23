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

 Classes: Levenshtein

 Description: Levenshtein algorithm.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#ifndef _LEVENSHTEIN_HPP
#define _LEVENSHTEIN_HPP

#include <string>

using namespace std;

class Levenshtein
{
    public:
        static int Distance(const string &s1, const string &s2);
        
    private:
        static inline int Minimum(int a, int b, int c)
        {
            return min(min(a, b), c);
        };
};

#endif

