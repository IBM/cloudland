/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <grpc++/grpc++.h>
#include "remotexec.grpc.pb.h"

#include <fstream>

using grpc::Channel;
using grpc::ClientContext;
using grpc::ClientWriter;
using grpc::Status;
using com::ibm::cloudland::scripts::FileChunk;
using com::ibm::cloudland::scripts::TransmitAck;
using com::ibm::cloudland::scripts::RemoteExec;

using namespace std;

class RemoteClient {
    public:
        RemoteClient(std::shared_ptr<Channel> channel)
            : stub_(RemoteExec::NewStub(channel)) {}
        std::string Transmit(int msg_id, char *control, char *srcfile, char *destpath, int trunksize) {
            ClientContext context;
            FileChunk chunk;
            TransmitAck ack;
            std::unique_ptr<ClientWriter<FileChunk> > writer(
                    stub_->Transmit(&context, &ack));

            ifstream myfile(srcfile, ios::in | ios::binary);
            char *buffer = new char[trunksize];
            if (!myfile.is_open()) {
                return "Open error";
            }

            myfile.seekg(0, myfile.end);
            int filesize = myfile.tellg();
            myfile.seekg(0, myfile.beg);
            chunk.set_id(msg_id);
            chunk.set_control(control);
            chunk.set_filepath(destpath);
            chunk.set_filesize(filesize);
            chunk.set_checksum(0);
            chunk.set_extra(msg_id);
            while (true) {
                chunk.set_fileseek(myfile.tellg());
                myfile.read(buffer, trunksize);
                chunk.set_content(buffer, myfile.gcount());
                if (myfile.eof()) {
                    chunk.set_extra(0);
                    writer->Write(chunk);
                    break;
                }
                writer->Write(chunk);
            };
            myfile.close();
            writer->WritesDone();
            Status status = writer->Finish();

            return "RPC OK";
        }
    private:
        std::unique_ptr<RemoteExec::Stub> stub_;
};

int main(int argc, char** argv) {
    if (argc < 5) {
        cout << argv[0] << " <extra> <control> <file_src> <file_dest> [trunksize]" << endl;
        exit(0);
    }
    int extra = atoi(argv[1]);
    char *ctl = argv[2];
    char *src = argv[3];
    char *dest = argv[4];
    int csize = 1024;
    if (argc == 6) {
        csize = atoi(argv[5]);
    }
    int msg_id = ::time(NULL);
    string endpoint = "localhost:50051";
    char *envp = getenv("GRPC_CLIENT_ENDPOINT");
    if (envp != NULL) {
        endpoint = envp;
    }
    RemoteClient client(grpc::CreateChannel(
                endpoint, grpc::InsecureChannelCredentials()));
//    for (i = 0; i < 10000; i++) {
    string reply = client.Transmit(msg_id, ctl, src, dest, csize);
    cout << "Remote received: " << reply << std::endl;
//    }

    return 0;
}
