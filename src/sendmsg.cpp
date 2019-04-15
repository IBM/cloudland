/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <iostream>
#include <string>

#include "unixsock.hpp"
#include "exception.hpp"

using namespace std;

void usage()
{
    printf("sendmsg <control> <command>\n");
    exit(-1);
}

int main(int argc, char *argv[])
{
    UnixSocket unixSock;
    char *p = argv[1];
    int len = 0;
    struct Command cmd = {{"raw"}, {0}, 0, NULL}; 

    if (argc < 2) {
        usage();
    }

    while (*p == ' ') {
        p++;
    }
    strncpy(cmd.control, p, sizeof(cmd.control) - 1);

    p = argv[2];
    while (*p == ' ') {
        p++;
    }
    cmd.size = strlen(p) + 1;
    cmd.content = new char[cmd.size];
    memset(cmd.content, '\0', cmd.size);
    strncpy(cmd.content, p, cmd.size);

    try {
        unixSock.send(cmd);
    } catch (CommonException &e) {
        cout << e.getErrMsg() << endl;
    }
    delete [] cmd.content;

    return 0;
}
