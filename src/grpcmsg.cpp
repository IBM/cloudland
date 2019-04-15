/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include "rpcworker.hpp"
#include <sys/time.h>
// #include <unistd.h>

using namespace std;

int main(int argc, char** argv) {
    int i = 0;

    if (argc < 3) {
        cout << argv[0] << " <extra> <control> [command]" << endl;
        exit(0);
    }
    int extra = atoi(argv[1]);
    char *ctl = argv[2];
    char *cmd = "";
    if (argc >= 4) {
        cmd = argv[3];
    }
    int msg_id = ::time(NULL);
    string endpoint = "localhost:50051";
    char *envp = getenv("GRPC_CLIENT_ENDPOINT");
    if (envp != NULL) {
        endpoint = envp;
    }
    FrontBack client(grpc::CreateChannel(
                endpoint, grpc::InsecureChannelCredentials()));
//    for (i = 0; i < 10000; i++) {
    string reply = client.Execute(msg_id, extra, ctl, cmd, "");
    cout << "Remote received: " << reply << endl;
//    }

    return 0;
}
