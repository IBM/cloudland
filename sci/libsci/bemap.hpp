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

 Classes: BEMap

 Description: Runtime BEMap 
   
 Author: ronglli 

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   08/24/12 ronglli   Initial code 

****************************************************************************/

#ifndef _BEMAP_HPP
#define _BEMAP_HPP

#include <map>
#include <string>
#include <vector>

using namespace std;

typedef vector<int> INT_VEC;
typedef vector<string> STRING_VEC;

typedef struct _sci_host_info
{
    int host_cnt;           /* # of host entries in current line */
    string front;      /* string before [..] */
    string end;        // string after [..]
    INT_VEC expanded_range;  //expanded range of inside [..]
    int repetition;         // repetition count after '*'
    int stride;             //stride after a range
    int range_begin, range_end, range_digits;
    STRING_VEC host_entries;
} sci_host_info;

typedef struct _sci_task_info
{
    int task_cnt;
    INT_VEC free_form;
    int range_begin, range_end;
    int stride;
} sci_task_info;

class BEMap : public map<int, string>
{
    private:
        sci_host_info hostinfo; // hostinfo for the current line
        sci_task_info taskinfo; // taskinfo for the current line
        int first_empty;

    public:
        int input(const char * filename, int num);
        int input(const char *hostlist[], int num);

        int trim_whitespace(string & line);
        int isValidForm(string line, const char * regex);
   
        int generate_host_range();
        int generate_host_entries();
        int expand_host_range(string range);
        int expand_host_region(string hregion);
    
        int generate_task_range();
        int expand_task_region(string tregion);
        int update_mapping(int num);
        void dump_mappings();
        int clear_lineinfo();
        int expand_line(string line, int num);
};

#endif

