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
	context "context"
	"io"
	"net"

	"github.com/uber/jaeger-client-go/utils"
	grpc "google.golang.org/grpc"
)

func Transfer(conn *grpc.ClientConn, targets ...io.Writer) (err error) {
	client := NewTraceAgentAdminClient(conn)
	rep, err := client.List(context.Background(), &ListRequest{})
	if err != nil {
		return
	}
	for {
		var out *ListReply
		out, err = rep.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		data := out.GetData()
		for _, target := range targets {
			_, err = target.Write(data)
			if err != nil {
				return
			}
		}
	}
}

func NewUDPConn(endpoint string) (connUDP *net.UDPConn, err error) {
	destAddr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return
	}

	if connUDP, err = net.DialUDP(destAddr.Network(), nil, destAddr); err != nil {
		return
	}

	if err = connUDP.SetWriteBuffer(utils.UDPPacketMaxLength); err != nil {
		return
	}
	return
}
