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

 Classes: DshAggregator

 Description: Aggregation functions.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#ifndef _DSHAGGREGATOR_HPP
#define _DSHAGGREGATOR_HPP

#include <stdio.h>
#include <assert.h>

#include "dsh_header.hpp"
#include "levenshtein.hpp"

#include <string>
#include <vector>
#include <algorithm>

using namespace std;

class DshAggregator
{
    private:
        vector<DshMessage*> messages;
        
    public:        
        void addMsg(DshMessage *msg) {
            assert(msg);

            messages.push_back(msg);
        }

        void deleteAll() {
            messages.clear();
        }

        void freeMemory(bool incStr = false) {
            for (int i=0; i<messages.size(); i++) {
                if (messages[i]) {
                    messages[i]->freeMemory(incStr);
                }
            }
            deleteAll();
        }

        DshMessage * getAggregatedMsg() {
            if (messages.size() == 0)
                return NULL;

            DshMessage *msg = new DshMessage(messages[0]->getSeqNo());
            for (int i=0; i<messages.size(); i++) {
                aggregateMsg(msg, messages[i]);
            }

            calcMsgState(msg);

            return msg;
        }

        int numOfBEs() {
            int num = 0;
            for (int i=0; i<messages.size(); i++) {
                num += messages[i]->getMaxLineSize();
            }
            return num;
        }

    private:
        void aggregateMsg(DshMessage *base_msg, DshMessage *new_msg) {
            assert((base_msg!=NULL) && (new_msg!=NULL));
            assert(base_msg->getSeqNo() == new_msg->getSeqNo());

            int newly = 0, base = 0;
            while ((base < base_msg->getSize()) && (newly < new_msg->getSize())) {
                DshLine *base_line = base_msg->getLine(base);
                DshLine *new_line = new_msg->getLine(newly);

                if (base_line->getLineNo() < new_line->getLineNo()) {
                    base++;
                    continue;
                }

                if (base_line->getLineNo() > new_line->getLineNo()) {
                    base_msg->addLine(new_line);
                    base++;
                } else {
                    int pos = base;
                    bool exists = false;
                    while (base_msg->getLine(pos)->getLineNo() == base_line->getLineNo()) {
                        if (isEqual(base_msg->getLine(pos)->getLine(), new_line->getLine())) {
                            for (int i=0; i<new_line->getSize(); i++) {
                                base_msg->getLine(pos)->addBE(new_line->getBE(i));
                            }
                            exists = true;
                            break;
                        }
                        pos++;
                        if (pos >= base_msg->getSize())
                            break;
                    }
                    if (!exists) {
                        base_msg->addLine(new_line);
                        if ((*new_line) < (*base_line)) {
                            base++;
                        }
                    }
                }
                newly++;
            }

            for (int i=newly; i<new_msg->getSize(); i++) {
                base_msg->addLine(new_msg->getLine(i));
            }
        }

        void calcMsgState(DshMessage *msg) {
            int start = 0, base_pos, num_equal;

            while (start < msg->getSize()) {
                base_pos = findBasePos(msg, start, &num_equal);

                bool hasSimilar = false;
                bool hasDiff = false;
                for (int i=start; i<start+num_equal; i++) {
                    if (i == base_pos)
                        continue;
                    
                    int state = compare(msg->getLine(base_pos)->getLine(), msg->getLine(i)->getLine());
                    if (state == EQUAL_STATE) {
                        printf("\n");
                        msg->print();
                        assert(!"Should not be equal state");
                    } else if (state == SIMILAR_STATE) {
                        hasSimilar = true;
                        msg->getLine(i)->setState(SIMILAR_STATE);
                    } else {
                        hasDiff = true;
                        msg->getLine(i)->setState(DIFFERENT_STATE);
                    }
                }
                if (hasSimilar && (base_pos == start)) {
                    if (msg->getLine(base_pos)->getSize() == 1)
                        msg->getLine(base_pos)->setState(SIMILAR_STATE);
                } else if (hasDiff && (base_pos == start)) {
                    if (msg->getLine(base_pos)->getSize() == 1)
                        msg->getLine(base_pos)->setState(DIFFERENT_STATE);
                }
                
                start += num_equal;
            }
        }

        int findBasePos(DshMessage *msg, int start, int *num_equal) {
            assert(msg);
            assert(start < msg->getSize());

            int pos = start, maxPos = start;
            int maxBEs = 0;
            *num_equal = 0;
            while (msg->getLine(pos)->getLineNo() == msg->getLine(start)->getLineNo()) {
                if (maxBEs < msg->getLine(pos)->getSize()) {
                    maxBEs = msg->getLine(pos)->getSize();
                    maxPos = pos;
                }
                *num_equal = *num_equal + 1;
                pos++;
                if (pos >= msg->getSize())
                    break;
            }

            return maxPos;
        }
        
        int compare(char *str1, char *str2) {
            assert(str1 && str2);

            string string1(str1);
            string string2(str2);

            int cost = Levenshtein::Distance(string1, string2);
            if (cost == 0)
                return EQUAL_STATE;
            else if (cost * 5 < max(string1.length(), string2.length()))
                return SIMILAR_STATE;
            else
                return DIFFERENT_STATE;
        }

        int isEqual(char *str1, char *str2) {
            assert(str1 && str2);

            string string1(str1);
            string string2(str2);

            if (string1 == string2)
                return 1;
            else {
                if (compare(str1, str2) == EQUAL_STATE)
                    return 1;
                else
                    return 0;
            }
        }
};

#endif

