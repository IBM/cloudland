#include <iostream>
#include <memory>
#include <string>

#include <grpc++/grpc++.h>

#ifdef BAZEL_BUILD
#include "examples/protos/remotexec.grpc.pb.h"
#else
#include "remotexec.grpc.pb.h"
#endif

using grpc::Channel;
using grpc::ClientContext;
using grpc::Status;
using remotexec::ExecuteRequest;
using remotexec::ExecuteReply;
using remotexec::RemoteExec;

class RemoteClient {
    public:
        RemoteClient(std::shared_ptr<Channel> channel)
            : stub_(RemoteExec::NewStub(channel)) {}

        std::string Execute() {
            ExecuteRequest request;

            request.set_id(1);
            request.set_control("inter");
            request.set_command("/opt/cloudland/scripts/backend/benchmark.sh 12345");
            ExecuteReply reply;
            ClientContext context;
            Status status = stub_->Execute(&context, request, &reply);

            // Act upon its status.
            if (status.ok()) {
                return reply.status();
            } else {
                std::cout << status.error_code() << ": " << status.error_message()
                    << std::endl;
                return "RPC failed";
            }
        }

    private:
        std::unique_ptr<RemoteExec::Stub> stub_;
};

/*
int main(int argc, char** argv) {
  RemoteClient client(grpc::CreateChannel(
      "localhost:50051", grpc::InsecureChannelCredentials()));
  std::string reply = client.Execute();
  std::cout << "Remote received: " << reply << std::endl;
  return 0;
}
*/
