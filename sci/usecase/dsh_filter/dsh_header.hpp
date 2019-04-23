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

 Classes: DshLine, DshMessage

 Description: Common definitions.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#ifndef _DSHHEADER_HPP
#define _DSHHEADER_HPP

#include <stdio.h>
#include <assert.h>
#include <string.h>

#include "dsh_packer.hpp"

#include <vector>
#include <algorithm>

using namespace std;

const int EQUAL_STATE  = 0;
const int SIMILAR_STATE = 1;
const int DIFFERENT_STATE = 2;

const char *EQUAL_COLOR = "\x1b[1;32m"; // green
const char *SIMILAR_COLOR = "\x1b[2;32m"; // dark green
const char *DIFFERENT_COLOR = "\x1b[1;31m"; // red
const char *END_COLOR = "\x1b[00m"; // end sign

class DshLine 
{
    private:
        vector<int>     be_ids;
        int             state;
        int             min_id;
        int             max_id;
        int             line_no;
        char            *line;

    public:
        DshLine(int st = EQUAL_STATE) {
            be_ids.clear();
            line_no = 0;
            line = NULL;
            state = st;
            min_id = max_id = 0;
        }

        void addBE(int be_id) {
            if (find(be_ids.begin(), be_ids.end(), be_id) == be_ids.end()) {
                be_ids.push_back(be_id);
                sort(be_ids.begin(), be_ids.end());
                if (be_id > max_id)
                    max_id = be_id;
                else if (be_id < min_id)
                    min_id = be_id;
            }
        }

        void deleteAll() {
            be_ids.clear();
        }
        
        void print() {
            int first, last;

            if (state == EQUAL_STATE)
                printf("%s", EQUAL_COLOR);
            else if (state == SIMILAR_STATE)
                printf("%s", SIMILAR_COLOR);
            else if (state == DIFFERENT_STATE)
                printf("%s", DIFFERENT_COLOR);
            else
                assert(!"Undefined message state");
            
            first = last = 0;
            for (int i=0; i<be_ids.size(); i++) {
                if (be_ids[i] == be_ids[first] + (i - first)) {
                    last = i;
                } else {
                    if (first == last)
                        printf("%d ", be_ids[first]);
                    else
                        printf("%d:%d ", be_ids[first], be_ids[last]);
                    first = last = i;
                }
            }
            if (first == last)
                printf("%d ", be_ids[first]);
            else
                printf("%d:%d ", be_ids[first], be_ids[last]);
            printf("| %s%s\n", line, END_COLOR);
        }

        void print(int be_id) {
            if (find(be_ids.begin(), be_ids.end(), be_id) == be_ids.end())
                return;

            if (state == EQUAL_STATE)
                printf("%s", EQUAL_COLOR);
            else if (state == SIMILAR_STATE)
                printf("%s", SIMILAR_COLOR);
            else if (state == DIFFERENT_STATE)
                printf("%s", DIFFERENT_COLOR);
            else
                assert(!"Undefined message state");

            printf("%d | %s%s\n", be_id, line, END_COLOR);
        }
        
        void setState(int st) {
            state = st;
        }
        
        int getState() {
            return state;
        }

        void setLineNo(int no) {
            line_no = no;
        }

        int getLineNo() {
            return line_no;
        }

        void setLine(char *str) {
            line = str;
        }

        char * getLine() {
            return line;
        }

        int getBE(int index) {
            assert(index>=0 || index<getSize());

            return be_ids[index];
        }
        
        int getSize() {
            return be_ids.size();
        }

        int getMinBEId() {
            return min_id;
        }

        int getMaxBEId() {
            return max_id;
        }

        bool  operator > (DshLine &dl) {
            assert((be_ids.size()>0) && (dl.be_ids.size()>0));
            
            if (line_no > dl.line_no)
                return true;

            if (line_no < dl.line_no)
                return false;

            if (be_ids[be_ids.size()-1] <= dl.be_ids[dl.be_ids.size()-1])
                return false;

            return true;
        }

        bool  operator < (DshLine &dl) {
            assert((be_ids.size()>0) && (dl.be_ids.size()>0));
            
            if (line_no < dl.line_no)
                return true;

            if (line_no > dl.line_no)
                return false;

            if (be_ids[0] >= dl.be_ids[0])
                return false;

            return true;
        }

        bool  operator == (DshLine &dl) {
            assert((be_ids.size()>0) && (dl.be_ids.size()>0));
            
            if (line_no != dl.line_no)
                return false;

            if (be_ids.size() != dl.be_ids.size())
                return false;

            for (int i=0; i<be_ids.size(); i++) {
                if (be_ids[i] != dl.be_ids[i])
                    return false;
            }

            return true;
        }
};

class DshMessage 
{
    private:
        int                seq_no;
        vector<DshLine*>   lines;
        int                max_id;
        int                min_id;

    public:
        DshMessage(int seq = 0) {
            seq_no = seq;
            lines.clear();
            max_id = min_id = 0;
        }

        DshMessage & readFromString(const char *str, int be_id = 0) {
            char *from = (char *) str, *to = NULL;

            lines.clear();
            int line_no = 0;
            while (1) {
                to = strstr(from, "\n");
                if (to == NULL) {
                    if (strlen(from) > 0) {
                        int len = strlen(from);
                        char *text = new char[len + 1];
                        strncpy(text, from, len);
                        text[len] = '\0';
                        chomp(text);

                        DshLine *line = new DshLine();
                        line->setLineNo(line_no++);
                        line->setLine(text);
                        line->addBE(be_id);
                        lines.push_back(line);
                    }
                    break;
                } else {
                    int len = strlen(from) - strlen(to);
                    char *text = new char[len + 1];
                    strncpy(text, from, len);
                    text[len] = '\0';
                    chomp(text);

                    DshLine *line = new DshLine();
                    line->setLineNo(line_no++);
                    line->setLine(text);
                    line->addBE(be_id);
                    lines.push_back(line);
                }
                from = to + 1;
            }

            max_id = min_id = be_id;
            
            return *this;
        }

        void addLine(DshLine *line) {
            assert(line);

            if (line->getMaxBEId() > max_id)
                max_id = line->getMaxBEId();
            else if (line->getMinBEId() < min_id)
                min_id = line->getMinBEId();

            vector<DshLine*>::iterator it = lines.begin();
            for (; it!=lines.end(); ++it) {
                DshLine *tmp = (*it);
                if ((*line) < (*tmp)) {
                    lines.insert(it, line);
                    return;
                }
            }
            
            lines.push_back(line);
        }
        
        void freeMemory(bool incStr = false) {
            for (int i=0; i<lines.size(); i++) {
                if (lines[i]) {
                    if (incStr)
                        delete [] lines[i]->getLine();
                    delete lines[i];
                }
            }
            deleteAll();
        }

        void deleteAll() {
            lines.clear();
        }

        DshLine * getLine(int index) {
            assert(index>=0 && index<lines.size());
            if (lines.size() == 0)
                return NULL;

            return lines[index];
        }

        void print(bool expand = false) {
            if (!expand) {
                for (int i=0; i<lines.size(); i++) {
                    lines[i]->print();
                }
            } else {
                for (int i=min_id; i<=max_id; i++) {
                    for (int j=0; j<lines.size(); j++) {
                        lines[j]->print(i);
                    }
                }
            }
        }

        void setSeqNo(int seq) {
            seq_no = seq;
        }

        int getSeqNo() {
            return seq_no;
        }

        void setLineState(int index, int state) {
            assert(index>=0 || index<lines.size());
            assert((state==EQUAL_STATE) || (state==SIMILAR_STATE) || (state==DIFFERENT_STATE));

            lines[index]->setState(state);
        }

        int getLineState(int index) {
            assert(index>=0 || index<lines.size());

            return lines[index]->getState();
        }

        int getLineNo(int index) {
            assert(index>=0 || index<lines.size());

            return lines[index]->getLineNo();
        }

        int getSize() {
            return lines.size();
        }

        int getMaxLineSize() {
            int size = 1;
            for (int i=0; i<lines.size(); i++) {
                if (lines[i]->getSize() > size)
                    size = lines[i]->getSize();
            }

            return size;
        }

        int getMaxBEId() {
            return max_id;
        }

        int getMinBEId() {
            return min_id;
        }

        void * pack(int *size = NULL) {
            DshPacker packer;
            int line_size;

            packer.packInt(seq_no);
            packer.packInt(lines.size());
            for (int i=0; i<lines.size(); i++) {
                packer.packInt(lines[i]->getSize());
                for (int j=0; j<lines[i]->getSize(); j++)
                    packer.packInt(lines[i]->getBE(j));
                packer.packInt(lines[i]->getState());
                packer.packInt(lines[i]->getLineNo());
                packer.packStr(lines[i]->getLine());
            }

            if (size)
                *size = packer.getPackedMsgLen();
            return packer.getPackedMsg();
        }
    
        DshMessage & unpack(void *buf) {
            DshPacker packer;
            packer.setPackedMsg((char *)buf);
        
            seq_no = packer.unpackInt();
            int num_lines = packer.unpackInt();
            for (int i=0; i<num_lines; i++) {
                DshLine *li = new DshLine();

                int num_bes = packer.unpackInt();
                for (int j=0; j<num_bes; j++)
                    li->addBE(packer.unpackInt());
                li->setState(packer.unpackInt());
                li->setLineNo(packer.unpackInt());
                li->setLine(packer.unpackStr());
        
                addLine(li);
            }

            return *this;
        }

    private:
        void chomp(char* str) {
            int end = strlen(str);
            while (end >= 0) {
                if ((str[end]==' ') || (str[end]=='\t'))
                    str[end] = '\0';
                else
                    break;
                end--;
            }
        }
};

#endif

