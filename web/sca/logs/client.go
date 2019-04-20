package logs

import (
	"google.golang.org/grpc"
)

func NewClientConn(endpoint string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	if len(opts) == 0 {
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.Dial(endpoint, opts...)
}
