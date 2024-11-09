/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package logs

import (
	"net"

	"github.com/uber/jaeger-client-go/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type traceAgent struct {
	queue chan []byte
}

func RegisterAgent(server *grpc.Server, listenAddr string) {
	ta := &traceAgent{
		queue: make(chan []byte, 1024),
	}
	go ta.ServeUDP(listenAddr)
	RegisterTraceAgentAdminServer(server, ta)
}

func (ta *traceAgent) List(req *ListRequest, rep TraceAgentAdmin_ListServer) (err error) {
	for {
		b := <-ta.queue
		if len(b) == 0 {
			continue
		}
		err = rep.Send(&ListReply{Data: b})
		if err != nil {
			err = grpc.Errorf(codes.Aborted, "%v", err)
			return
		}
	}
	return
}

func (ta *traceAgent) ServeUDP(listenAddr string) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return
	}
	defer pc.Close()
	for {
		b := make([]byte, utils.UDPPacketMaxLength)
		n, _, err := pc.ReadFrom(b)
		if err != nil {
			continue
		}
		if n == 0 {
			continue
		}
		b = b[0:n]
		ta.queue <- b
	}
}

func RunAgent(listenAddr string) (err error) {
	server := grpc.NewServer()
	RegisterAgent(server, listenAddr)
	var listen net.Listener
	listen, err = net.Listen("tcp", listenAddr)
	if err != nil {
		return
	}
	err = server.Serve(listen)
	return
}
