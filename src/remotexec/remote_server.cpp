#include "remote_server.hpp"

Status RemoteExecServiceImpl::Execute(ServerContext* context, const ExecuteRequest* request, ExecuteReply* reply) 
{
    reply->set_status("OK");
    return Status::OK;
}

void RunServer() {
  std::string server_address("0.0.0.0:50051");
  RemoteExecServiceImpl service;

  ServerBuilder builder;
  builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
  builder.RegisterService(&service);
  std::unique_ptr<Server> server(builder.BuildAndStart());
  std::cout << "Server listening on " << server_address << std::endl;
  server->Wait();
}

/*
int main(int argc, char** argv) {
  RunServer();

  return 0;
}
*/
