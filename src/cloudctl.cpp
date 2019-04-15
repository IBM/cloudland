/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <iostream>
#include <string>

#include "unixsock.hpp"
#include "exception.hpp"

using namespace std;

int main(int argc, char *argv[])
{
    UnixSocket unixSock;
    char msg[1024] = {0};
    char *p = msg;
    int len = 0;
    struct Command cmd = {{"raw"}, {"exec"}, 0, NULL}; 

    while (true) {
        cout << ">>> " << ends;

        memset(msg, '\0', sizeof(msg));
        cin.getline(msg, sizeof(msg));
        if ((strcmp(msg, "exit") == 0) || cin.eof()) {
            cout << endl;
            break;
        }

        if (msg[0] == '\0')
            continue;

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
    }

    return 0;
}
