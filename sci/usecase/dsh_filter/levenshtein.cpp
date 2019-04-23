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

#include "levenshtein.hpp"

#include <algorithm>

int Levenshtein::Distance(const string &s1, const string &s2)
{
    int i, j, cost;
    int n = s1.length(), m = s2.length();
    int *matrix = new int[(m + 1) * (n + 1)]; 

    if (m == 0) 
        return n;
    if (n == 0) 
        return m;

    m++;
    n++;

    for (i = 0; i < n; i ++)    
        matrix[i] = i;
    for (i = 0; i < m; i ++)    
        matrix[i * n] = i;
    
    for (i = 1; i < n; i ++) {
        for (j = 1; j < m; j ++) {
            if (s1[i - 1] == s2[j - 1])
                cost = 0;
            else
                cost = 1;
            matrix[j * n + i] = Minimum(matrix[(j-1)*n+i]+1, matrix[j*n+i-1]+1, matrix[(j-1)*n+i-1]+cost);
        }
    }

    cost = matrix[n * m - 1];
    delete [] matrix;

    return cost;
}

