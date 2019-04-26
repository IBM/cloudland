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

Description: Runtime BEMap. 

Author: ronglli 

History:
Date     Who ID    Description
-------- --- ---   -----------
08/24/12 ronglli        Initial code

 ****************************************************************************/

#include <stdlib.h>
#include <regex.h>
#include <fstream>

#include "sci.h"
#include "bemap.hpp"
#include "log.hpp"

#define BEMAP_SUCCESS 0
#define BEMAP_ERROR -1
#define BEMAP_FULL   (-999999)

#define VALID_REGEX "^[^][%*]*(\\[[0-9]+-[0-9]+(:[0-9]+)?\\]){0,}[^][%*]*([*][0-9]+)?(%\\[(([0-9]+,){0,}[0-9]+|[0-9]+-[0-9]+(:[0-9]+)?)\\])?$"

int BEMap::input(const char *hostlist[], int num)
{
    int i = 0;

    while (hostlist[i] != NULL) {
        if ((num > 0) && (i >= num)) {
            break;
        }
        (*this)[i] = hostlist[i];
        i++;
    }

    return SCI_SUCCESS;
}

int BEMap::input(const char * filename, int num)
{
    int rc = -1;
    ifstream fs;
    string line;
    size_t pos = -1;
    int max_id, size;

    fs.open(filename);
    if (!fs) {
        return SCI_ERR_INVALID_HOSTFILE;;
    }

    clear_lineinfo();
    (*this).first_empty = 0;

    log_debug("Hostlist is: ");
    while(fs) {
        if ((num > 0) && ((*this).size() >= num)) {
            break;
        }
        getline(fs, line);
        trim_whitespace(line);
        if (line.length() == 0) {
            continue;
        }
        if ((line.find_first_of('!') == 0) || (line.find_first_of('*') == 0)) {
            continue;
        }

        rc = isValidForm(line, VALID_REGEX);
        if (rc) {
            return SCI_ERR_INVALID_HOSTFILE;
        }

        rc = expand_line(line, num);
        if (rc == BEMAP_FULL) { 
            break;
        } else if (rc != BEMAP_SUCCESS){
            return SCI_ERR_INVALID_HOSTFILE;
        }

        clear_lineinfo();
    }
    fs.close();
  
    size = (*this).size(); 
    if (size == 0) {
        log_error("BEMap error: empty host file.");
        return SCI_ERR_INVALID_HOSTFILE;
    } 

    max_id = ((*this).rbegin())->first;
    if (max_id >= size) {
        log_error("BEMap error: max_id(%d) needs to be smaller than the totalsize(%d)", 
                max_id, size);
        return SCI_ERR_INVALID_HOSTFILE;
    }

    return SCI_SUCCESS;
}

int BEMap::trim_whitespace(string & line)
{
    string::iterator it;
    for (it = line.begin(); it != line.end();) {
        if (isspace(*it))
            it = line.erase(it);
        else
            it++;
    }
    return BEMAP_SUCCESS;
}

int BEMap::isValidForm(string line, const char * regex)
{
    int rc = BEMAP_SUCCESS;
    regex_t preg;

    if((line.find_first_of('[') == string::npos) && (line.find_first_of('*') == string::npos)) {
        if((line.find_first_of('%') == string::npos) && (line.find_first_of(':') == string::npos))
            return rc;
    }

    rc = regcomp(&preg, regex, REG_EXTENDED|REG_NOSUB|REG_NEWLINE);
    if(rc != 0) {
        log_error("regcomp error, line:%s, rc = %d", line.c_str(), rc);
        return BEMAP_ERROR;
    }

    rc = regexec(&preg, line.c_str(), 0, 0, 0);
    if(rc != 0) {
        log_error("regex NOT match, line:%s, rc = %d", line.c_str(), rc);
        return BEMAP_ERROR;
    }
    return rc;
}

int BEMap::generate_host_range()
{
    if (hostinfo.range_begin == -1) {
        hostinfo.host_cnt = hostinfo.repetition;
    } else {
        int begin;
        begin = hostinfo.range_begin;
        while (begin <= hostinfo.range_end) {
            hostinfo.expanded_range.push_back(begin);
            hostinfo.host_cnt++;
            begin += hostinfo.stride;
        }
        hostinfo.host_cnt *= hostinfo.repetition;
    }
    return BEMAP_SUCCESS;
}

int BEMap::generate_host_entries()
{
    int i;
    if (hostinfo.expanded_range.empty()) {
        for (i = 0; i < hostinfo.repetition; i++) {
            hostinfo.host_entries.push_back(hostinfo.front + hostinfo.end);
        }
    } else {
        INT_VEC::iterator iter;
        char buf[128], format[32];
        sprintf(format, "%%0%dd%", hostinfo.range_digits);
        for (iter = hostinfo.expanded_range.begin(); iter != hostinfo.expanded_range.end(); iter++) {
            for (i = 0; i < hostinfo.repetition; i++) {
                sprintf(buf, format, *iter);
                hostinfo.host_entries.push_back(hostinfo.front + buf + hostinfo.end);
            }
        }
    }
    return BEMAP_SUCCESS;
}

int BEMap::expand_host_range(string range)
{
    int pos = -1;
    int pos1 = -1;
    string tmps;

    pos = range.find_first_of('-');
    tmps = range.substr(0, pos);
    hostinfo.range_begin = atoi(tmps.c_str());
    hostinfo.range_digits = tmps.size();

    pos1 = range.find_first_of(':', pos+1);
    if (pos1 == string::npos) {
        tmps = range.substr(pos+1);
        hostinfo.range_end = atoi(tmps.c_str());
        hostinfo.stride = 1;
    } else {
        tmps = range.substr(pos+1, pos1-pos-1);
        hostinfo.range_end = atoi(tmps.c_str());
        tmps = range.substr(pos1+1);
        hostinfo.stride = atoi(tmps.c_str());
        if (hostinfo.stride <= 0) {
            log_error("stride is %d, it must >= 1", hostinfo.stride);
            return BEMAP_ERROR;
        }
    }

    if ((hostinfo.range_begin < 0) || hostinfo.range_end < 0) {
        log_error("host part: both left side(%d) & right side(%d) of a range must >= 0", 
                hostinfo.range_begin, hostinfo.range_end);
        return BEMAP_ERROR;
    }
    if (hostinfo.range_end < hostinfo.range_begin) {
        log_error("host part: right side(%d) of a range must >= left side (%d) of a range",
                hostinfo.range_end, hostinfo.range_begin);
        return BEMAP_ERROR;
    }

    return BEMAP_SUCCESS;
}

int BEMap::expand_host_region(string hregion)
{
    int rc = BEMAP_SUCCESS;
    int pos = -1;
    int left = -1;
    int right = -1;

    pos = hregion.find_first_of('*');
    if (pos == string::npos) {
        hostinfo.repetition = 1;
    } else {
        string tmps = hregion.substr(pos+1);
        hostinfo.repetition = atoi(tmps.c_str());
    }

    if (hostinfo.repetition <= 0) {
        log_error("repetition(%d) of hosts must >= 1", hostinfo.repetition);
        return BEMAP_ERROR;
    }

    left = hregion.find_first_of('[');
    right = hregion.find_first_of(']');

    if ((left == string::npos) && (right == string::npos)) {
        hostinfo.front = hregion.substr(0, pos);
        hostinfo.end = "";
        hostinfo.stride = 1;
        hostinfo.range_begin = -1;
        hostinfo.range_end = -1;
    } else {
        if ((left == string::npos) || (right == string::npos)) {
            return BEMAP_ERROR;
        }
        hostinfo.front = hregion.substr(0, left);
        if (pos == string::npos) {
            hostinfo.end = hregion.substr(right+1);
        } else {
            hostinfo.end = hregion.substr(right+1, pos-right-1);
        }
        string range = hregion.substr(left+1, right-left-1);
        rc = expand_host_range(range);
        if (rc != 0) {
            return BEMAP_ERROR;
        }
    }

    generate_host_range();
    generate_host_entries();

    return BEMAP_SUCCESS;
}

int BEMap::generate_task_range()
{
    if (taskinfo.range_begin == -1) {
        taskinfo.task_cnt = taskinfo.free_form.size();
    } else {
        int begin;
        begin = taskinfo.range_begin;
        while(begin <= taskinfo.range_end) {
            taskinfo.free_form.push_back(begin);
            taskinfo.task_cnt++;
            begin += taskinfo.stride;
        }
    }
    return BEMAP_SUCCESS;
}

int BEMap::expand_task_region(string tregion)
{
    int left = -1;
    int right = -1;
    int pos = -1;
    int pos1 = -1;
    string tmps;

    left = tregion.find_first_of('[');
    right = tregion.find_first_of(']');
    if ((left == string::npos) || (right == string::npos)) {
        log_error("task region: it must start with '[' and end with ']'");
        return BEMAP_ERROR;
    }

    pos = tregion.find_first_of('-');
    if (pos != string::npos) {
        if ((pos < left) || (pos > right)) {
            log_error("task region: the '-' must be between the '[' and ']'");
            return BEMAP_ERROR;
        }
        tmps = tregion.substr(left+1, pos-left-1);
        taskinfo.range_begin = atoi(tmps.c_str());
        pos1 = tregion.find_first_of(':', pos+1);
        if (pos1 == string::npos) {
            tmps = tregion.substr(pos+1, right-pos-1);
            taskinfo.range_end = atoi(tmps.c_str());
            taskinfo.stride = 1;
        } else {
            tmps = tregion.substr(pos+1, pos1-pos-1);
            taskinfo.range_end = atoi(tmps.c_str());
            tmps = tregion.substr(pos1+1, right-pos1-1);
            taskinfo.stride = atoi(tmps.c_str());

            if (taskinfo.stride <= 0) {
                log_error("task part: stride = %d, it must >= 1", taskinfo.stride);
                return BEMAP_ERROR;
            }

            if ((taskinfo.range_begin < 0) || (taskinfo.range_end < 0)) {
                log_error("task part: both sides of the range(%d, %d) must >=0", 
                        taskinfo.range_begin, taskinfo.range_end);
                return BEMAP_ERROR;
            }

            if (taskinfo.range_end < taskinfo.range_begin) {
                log_error("task part: right side(%d) of a range must >= left side(%d) of a range",
                        taskinfo.range_end, taskinfo.range_begin);
                return BEMAP_ERROR;
            }
        }
    } else {
        int tid;
        int start = left;

        pos1 = tregion.find_first_of(',');
        if (pos1 != string::npos) {
            do {
                tmps = tregion.substr(start+1, pos1-start-1);
                tid = atoi(tmps.c_str());
                taskinfo.free_form.push_back(tid);
                start = pos1;
                pos1 = tregion.find_first_of(',', pos1+1);
            } while (pos1 != string::npos);
        }
        tmps = tregion.substr(start+1, right-left-1);
        tid = atoi(tmps.c_str());
        taskinfo.free_form.push_back(tid);
    }

    if (generate_task_range() != 0) {
        return BEMAP_ERROR;
    }

    return BEMAP_SUCCESS;
}

int BEMap::update_mapping(int num)
{
    INT_VEC::iterator t_iter;
    STRING_VEC::iterator h_iter;

    if (taskinfo.task_cnt > 0) {
        t_iter = taskinfo.free_form.begin();
        h_iter = hostinfo.host_entries.begin();
        for (; t_iter != taskinfo.free_form.end(); t_iter++, h_iter++) {
            if (*t_iter < 0) {
                log_error("task id(%d) must >= 0", *t_iter);
                return BEMAP_ERROR;
            }

            if ((*this).find(*t_iter) != (*this).end()) {
                log_error("error: duplicated task id(%d) for one job", *t_iter);
                return BEMAP_ERROR;
            }
            if ((num > 0) && ((*this).size() >= num)) {
                return BEMAP_FULL;
            }
            (*this)[*t_iter] = *h_iter;
            log_debug("[%d]: %s", *t_iter, (*h_iter).c_str());
        }
    } else {
        for (h_iter = hostinfo.host_entries.begin(); h_iter != hostinfo.host_entries.end(); h_iter++) {
            int index; 
            bool found = false;
            int max_id;
            int size = (*this).size();
            
            if ((num > 0) && (size >= num)) {
                return BEMAP_FULL;
            }
            if (size > 0) {
                max_id = ((*this).rbegin())->first;
                for (index = (*this).first_empty; index < max_id; index++) {
                    if ((*this).find(index) == (*this).end()) {
                        found = true;
                        (*this).first_empty = index + 1;
                        break;
                    }
                }
                if (!found) {
                    index = max_id + 1;
                }
            } else {
                index = 0;
                (*this).first_empty = index + 1;
            }
            (*this)[index] = *h_iter;
            log_debug("[%d]: %s", index, (*h_iter).c_str());
        }
    }

    return BEMAP_SUCCESS;
}

int BEMap::expand_line(string line, int num)
{
    int rc = BEMAP_SUCCESS;
    int pos = -1;
    string host_region;
    string task_region;

    pos = line.find_first_of('%');
    if (pos != string::npos) {
        host_region = line.substr(0, pos);
        task_region = line.substr(pos+1);
    } else {
        host_region = line;
    }

    rc = expand_host_region(host_region);
    if (rc != 0) {
        return BEMAP_ERROR;
    }
    if (pos != string::npos) {
        rc = expand_task_region(task_region);
        if (rc != 0) {
            return BEMAP_ERROR;
        }
    }

    if (taskinfo.task_cnt > 0) {
        if (hostinfo.host_cnt != taskinfo.task_cnt) {
            log_error("host count(%d) and task count(%d) not match, current line:%s", 
                    hostinfo.host_cnt, taskinfo.task_cnt, line.c_str());
            return BEMAP_ERROR;
        }
    }
    rc = update_mapping(num);

    return rc;
}

int BEMap::clear_lineinfo()
{
    hostinfo.front.clear();
    hostinfo.end.clear();
    hostinfo.host_cnt = 0;
    hostinfo.repetition = 1;
    hostinfo.stride = 1;
    hostinfo.range_begin = -1;
    hostinfo.range_end = -1;
    hostinfo.range_digits = -1;
    hostinfo.expanded_range.clear();
    hostinfo.host_entries.clear();

    taskinfo.task_cnt = 0;
    taskinfo.free_form.clear();
    taskinfo.range_begin = -1;
    taskinfo.range_end = -1;
    taskinfo.stride = 1;

    return BEMAP_SUCCESS;
}

void BEMap::dump_mappings()
{
    BEMap::iterator iter = (*this).begin();

    log_debug("Hostlist is: ");
    for (; iter != (*this).end(); iter++) {
        log_debug("[%d]: %s", iter->first, iter->second.c_str());
    }
}

