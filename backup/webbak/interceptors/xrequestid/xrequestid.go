/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package xrequestid

import (
	"context"

	"github.com/google/uuid"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	OutgoingKey = "X-Request-ID"
	IncomingKey = "x-request-id"
	TagKey      = "Request-ID"
)

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestID := RequestIDFromMetadata(ctx)
	// Set to context
	ctx = context.WithValue(ctx, OutgoingKey, requestID)
	// Mark span
	if sp := opentracing.SpanFromContext(ctx); sp != nil {
		sp.SetTag(TagKey, requestID)
	}
	rep, err := handler(ctx, req)
	// Send request id header back
	grpc.SendHeader(ctx, metadata.Pairs(OutgoingKey, requestID))
	return rep, err
}

func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	requestID := RequestIDFromMetadata(ctx)
	// Set to context
	ctx = context.WithValue(ctx, OutgoingKey, requestID)
	// Mark span
	if sp := opentracing.SpanFromContext(ctx); sp != nil {
		sp.SetTag(TagKey, requestID)
	}
	wrap := middleware.WrapServerStream(stream)
	wrap.WrappedContext = ctx
	err := handler(srv, wrap)
	wrap.SendHeader(metadata.Pairs(OutgoingKey, requestID))
	return err
}

func UnaryClientInterceptor(ctx context.Context, method string, req, rep interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	requestID := RequestIDFromContext(ctx)
	if requestID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, OutgoingKey, requestID)
	}
	err := invoker(ctx, method, req, rep, cc, opts...)
	return err
}

func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	requestID := RequestIDFromContext(ctx)
	if requestID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, OutgoingKey, requestID)
	}
	stream, err := streamer(ctx, desc, cc, method, opts...)
	return stream, err
}

func RequestIDFromContext(ctx context.Context) (id string) {
	if v := ctx.Value(OutgoingKey); v != nil { // Read context
		if s, ok := v.(string); ok {
			id = s
		}
	}
	return
}

func RequestIDFromMetadata(ctx context.Context) (id string) {
	if id == "" { // Read from metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if header, ok := md[IncomingKey]; ok && len(header) > 0 {
				id = header[0]
			}
		}
	}
	if id == "" { // Generate new
		id = uuid.New().String()
	}
	return
}
