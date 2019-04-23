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

 Classes: None

 Description: Unit tests.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/29/08 nieyy        Initial code (D154050)

****************************************************************************/

#include <string.h>

#include "levenshtein.hpp"
#include "dsh_header.hpp"
#include "dsh_aggregator.hpp"

#include <string>

using namespace std;

void Levenshtein_Test()
{
    string str1 = "This is a test";
    string str2 = "This is a test";

    int cost = Levenshtein::Distance(str1, str2);
    printf("Distance of [%s] and [%s] is %d\n", str1.c_str(), str2.c_str(), cost);
    if (cost == 0)
        printf("Passed\n");
    else
        printf("Failed\n");

    str1 = "This is a test";
    str2 = "This is a test!";
    cost = Levenshtein::Distance(str1, str2);
    printf("Distance of [%s] and [%s] is %d\n", str1.c_str(), str2.c_str(), cost);
    if (cost == 1)
        printf("Passed\n");
    else
        printf("Failed\n");

    str1 = "-rwxr-xr-x 1 nicole nicole 1060380 2008-10-29 13:21 dsh_be";
    str2 = "-rwxr-xr-x 1 nicole nicole 1234567 2008-10-29 13:21 dsh_be";
    cost = Levenshtein::Distance(str1, str2);
    printf("Distance of [%s] and [%s] is %d\n", str1.c_str(), str2.c_str(), cost);
    if (cost == 6)
        printf("Passed\n");
    else
        printf("Failed\n");
}

void DshLine_Test()
{
    DshLine line;
    char *str = new char[256];
    strcpy(str, "This is a test");
    line.setLine(str);

    line.addBE(0);
    line.print();

    line.addBE(1);
    line.print();

    for (int i=3; i<10; i++)
        line.addBE(i);
    line.print();

    for (int i=5; i<20; i++)
        line.addBE(i);
    line.print();

    line.addBE(30);
    line.addBE(50);
    line.addBE(70);
    line.addBE(90);
    line.print();

    delete [] str;
}

void DshLine_Test_Ext()
{
    DshLine line1, line2;

    line1.addBE(0);
    line2.addBE(0);
    if (line1 == line2)
        printf("[%d] 0 = [%d] 0\n", line1.getLineNo(), line2.getLineNo());
    else
        printf("Failed\n");

    line1.addBE(1);
    if (line1 > line2)
        printf("[%d] 0:1 > [%d] 0\n", line1.getLineNo(), line2.getLineNo());
    else
        printf("Failed\n");

    line2.setLineNo(line1.getLineNo() + 1);
    if (line1 < line2)
        printf("[%d] 0:1 < [%d] 0\n", line1.getLineNo(), line2.getLineNo());
    else
        printf("Failed\n");
}

void DshMessage_Test()
{
    DshMessage msg;

    for (int i=0; i<50; i++) {
        DshLine *line = new DshLine();
        if (i%2 == 0)
            line->setLine("Mary is a women");
        else
            line->setLine("Robert is a man");

        for (int j=i; j<i+5; j++)
            line->addBE(j);

        if (i%3 == 0)
            line->setState(EQUAL_STATE);
        else if (i%3 == 1)
            line->setState(SIMILAR_STATE);
        else
            line->setState(DIFFERENT_STATE);

        msg.addLine(line);
    }

    msg.print();
    msg.freeMemory();
}

void DshMessage_Test_Ext()
{
    DshMessage msg;

    for (int i=0; i<10; i++) {
        DshLine *line = new DshLine();
        line->addBE(i);
        if (i%2 == 0)
            line->setLine("I am Nicole Nie");
        else
            line->setLine("My wife is Awa Wang");
        msg.addLine(line);
    }
    msg.print();
    msg.freeMemory();
    
    for (int i=10; i>0; i--) {
        DshLine *line = new DshLine();
        line->addBE(0);
        line->setLineNo(i);
        if (i%2 == 0)
            line->setLine("Nicole loves Awa");
        else
            line->setLine("Awa loves Nicole");
        msg.addLine(line);
    }
    msg.print();
    msg.freeMemory();
}

void DshMessage_Test_Pack()
{
    DshMessage msg1, msg2;

    for (int i=0; i<10; i++) {
        DshLine *line = new DshLine();
        line->addBE(i);
        if (i%2 == 0)
            line->setLine("I am Nicole Nie");
        else
            line->setLine("My wife is Awa Wang");
        msg1.addLine(line);
    }

    printf("\n");
    msg1.print();

    msg2.unpack(msg1.pack());
    printf("\n");
    msg2.print();

    msg1.freeMemory();
}

void DshMessage_Test_ReadStr()
{
    DshMessage msg;

    msg.readFromString("This is a test 1\nThis is a test 2\nThis is a test 3\nThis is a test 4\nThis is a test 5\n");

    msg.print();
    for (int i=0; i<msg.getSize(); i++)
        delete msg.getLine(i)->getLine();
}

void DshAggregator_Test()
{
    DshAggregator aggregator;
    DshMessage *aggregatedMsg = NULL;

    const int NUM_MSG = 10;
    const int NUM_LINE = 10;

    // test 1
    printf("\n");
    for (int i=0; i<NUM_MSG; i++) {
        DshMessage *msg = new DshMessage();
        for (int j=0; j<NUM_LINE; j++) {
            DshLine *line = new DshLine();
            line->setLineNo(j);
            if (j%2 == 0)
                line->setLine("I am a man");
            else
                line->setLine("I am a women");
            line->addBE(i);
            msg->addLine(line);
        }
        aggregator.addMsg(msg);
    }
    
    aggregatedMsg = aggregator.getAggregatedMsg();
    aggregatedMsg->print();
    delete aggregatedMsg;

    aggregator.freeMemory();

    // test 2
    printf("\n");
    for (int i=0; i<NUM_MSG; i++) {
        DshMessage *msg = new DshMessage();
        for (int j=0; j<NUM_LINE; j++) {
            DshLine *line = new DshLine();
            line->setLineNo(j);
            if (i == (NUM_MSG - 1))
                line->setLine("I am a man");
            else
                line->setLine("Who are you");
            line->addBE(i);
            msg->addLine(line);
        }
        aggregator.addMsg(msg);
    }
    
    aggregatedMsg = aggregator.getAggregatedMsg();
    aggregatedMsg->print();
    delete aggregatedMsg;

    aggregator.freeMemory();

    // test 3
    printf("\n");
    for (int i=0; i<NUM_MSG; i++) {
        DshMessage *msg = new DshMessage();
        for (int j=0; j<NUM_LINE; j++) {
            DshLine *line = new DshLine();
            line->setLineNo(j);
            if (i == (NUM_MSG - 1))
                line->setLine("Who are you");
            else if (i == (NUM_MSG - 2))
                line->setLine("I am a women");
            else
                line->setLine("I am a woman");
            line->addBE(i);
            msg->addLine(line);
        }
        aggregator.addMsg(msg);
    }
    
    aggregatedMsg = aggregator.getAggregatedMsg();
    aggregatedMsg->print();
    delete aggregatedMsg;

    aggregator.freeMemory();
}

void DshAggregator_Test_Ext()
{
    DshAggregator aggregator;
    DshMessage *aggregatedMsg = NULL;

    const int NUM_MSG = 10;
    const int NUM_LINE = 10;

    // test 1
    printf("\n");
    for (int i=NUM_MSG-1; i>=0; i--) {
        DshMessage *msg = new DshMessage();
        for (int j=0; j<NUM_LINE; j++) {
            DshLine *line = new DshLine();
            line->setLineNo(j);
            if (j%2 == 0)
                line->setLine("I am a man");
            else
                line->setLine("I am a women");
            line->addBE(i);
            msg->addLine(line);
        }
        aggregator.addMsg(msg);
    }
    
    aggregatedMsg = aggregator.getAggregatedMsg();
    aggregatedMsg->print();
    delete aggregatedMsg;

    aggregator.freeMemory();
}

int main(int argc, char **argv)
{
    // Test Levenshtein class
    Levenshtein_Test();

    // Test DshLine class
    DshLine_Test();

    // Test DshLine class
    DshLine_Test_Ext();

    // Test DshMessage class
    DshMessage_Test();

    // Test DshMessage class
    DshMessage_Test_Ext();

    // Test DshMessage class
    DshMessage_Test_Pack();

    // Test DshMessage class
    DshMessage_Test_ReadStr();

    // Test DshAggregator class
    DshAggregator_Test();

    // Test DshAggregator class
    DshAggregator_Test_Ext();
    
    return 0;
}

