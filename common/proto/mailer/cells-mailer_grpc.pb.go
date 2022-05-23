// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package mailer

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MailerServiceClient is the client API for MailerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MailerServiceClient interface {
	SendMail(ctx context.Context, in *SendMailRequest, opts ...grpc.CallOption) (*SendMailResponse, error)
	ConsumeQueue(ctx context.Context, in *ConsumeQueueRequest, opts ...grpc.CallOption) (*ConsumeQueueResponse, error)
}

type mailerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMailerServiceClient(cc grpc.ClientConnInterface) MailerServiceClient {
	return &mailerServiceClient{cc}
}

func (c *mailerServiceClient) SendMail(ctx context.Context, in *SendMailRequest, opts ...grpc.CallOption) (*SendMailResponse, error) {
	out := new(SendMailResponse)
	err := c.cc.Invoke(ctx, "/mailer.MailerService/SendMail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mailerServiceClient) ConsumeQueue(ctx context.Context, in *ConsumeQueueRequest, opts ...grpc.CallOption) (*ConsumeQueueResponse, error) {
	out := new(ConsumeQueueResponse)
	err := c.cc.Invoke(ctx, "/mailer.MailerService/ConsumeQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MailerServiceServer is the server API for MailerService service.
// All implementations must embed UnimplementedMailerServiceServer
// for forward compatibility
type MailerServiceServer interface {
	SendMail(context.Context, *SendMailRequest) (*SendMailResponse, error)
	ConsumeQueue(context.Context, *ConsumeQueueRequest) (*ConsumeQueueResponse, error)
	mustEmbedUnimplementedMailerServiceServer()
}

// UnimplementedMailerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMailerServiceServer struct {
}

func (UnimplementedMailerServiceServer) SendMail(context.Context, *SendMailRequest) (*SendMailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMail not implemented")
}
func (UnimplementedMailerServiceServer) ConsumeQueue(context.Context, *ConsumeQueueRequest) (*ConsumeQueueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConsumeQueue not implemented")
}
func (UnimplementedMailerServiceServer) mustEmbedUnimplementedMailerServiceServer() {}

// UnsafeMailerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MailerServiceServer will
// result in compilation errors.
type UnsafeMailerServiceServer interface {
	mustEmbedUnimplementedMailerServiceServer()
}

func RegisterMailerServiceServer(s grpc.ServiceRegistrar, srv MailerServiceServer) {
	s.RegisterService(&MailerService_ServiceDesc, srv)
}

func _MailerService_SendMail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailerServiceServer).SendMail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mailer.MailerService/SendMail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailerServiceServer).SendMail(ctx, req.(*SendMailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MailerService_ConsumeQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConsumeQueueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailerServiceServer).ConsumeQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mailer.MailerService/ConsumeQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailerServiceServer).ConsumeQueue(ctx, req.(*ConsumeQueueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MailerService_ServiceDesc is the grpc.ServiceDesc for MailerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MailerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mailer.MailerService",
	HandlerType: (*MailerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMail",
			Handler:    _MailerService_SendMail_Handler,
		},
		{
			MethodName: "ConsumeQueue",
			Handler:    _MailerService_ConsumeQueue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cells-mailer.proto",
}