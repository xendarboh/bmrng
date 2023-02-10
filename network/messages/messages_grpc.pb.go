// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package messages

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

// MessageHandlersClient is the client API for MessageHandlers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessageHandlersClient interface {
	HandleSignedMessage(ctx context.Context, in *NetworkMessage, opts ...grpc.CallOption) (*NetworkMessage, error)
	HandleSignedMessageStream(ctx context.Context, opts ...grpc.CallOption) (MessageHandlers_HandleSignedMessageStreamClient, error)
	HealthCheck(ctx context.Context, in *NetworkMessage, opts ...grpc.CallOption) (*NetworkMessage, error)
	SkipPathGen(ctx context.Context, in *SkipPathGenMessage, opts ...grpc.CallOption) (*NetworkMessage, error)
}

type messageHandlersClient struct {
	cc grpc.ClientConnInterface
}

func NewMessageHandlersClient(cc grpc.ClientConnInterface) MessageHandlersClient {
	return &messageHandlersClient{cc}
}

func (c *messageHandlersClient) HandleSignedMessage(ctx context.Context, in *NetworkMessage, opts ...grpc.CallOption) (*NetworkMessage, error) {
	out := new(NetworkMessage)
	err := c.cc.Invoke(ctx, "/messages.MessageHandlers/HandleSignedMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageHandlersClient) HandleSignedMessageStream(ctx context.Context, opts ...grpc.CallOption) (MessageHandlers_HandleSignedMessageStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &MessageHandlers_ServiceDesc.Streams[0], "/messages.MessageHandlers/HandleSignedMessageStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &messageHandlersHandleSignedMessageStreamClient{stream}
	return x, nil
}

type MessageHandlers_HandleSignedMessageStreamClient interface {
	Send(*NetworkMessage) error
	Recv() (*NetworkMessage, error)
	grpc.ClientStream
}

type messageHandlersHandleSignedMessageStreamClient struct {
	grpc.ClientStream
}

func (x *messageHandlersHandleSignedMessageStreamClient) Send(m *NetworkMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *messageHandlersHandleSignedMessageStreamClient) Recv() (*NetworkMessage, error) {
	m := new(NetworkMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *messageHandlersClient) HealthCheck(ctx context.Context, in *NetworkMessage, opts ...grpc.CallOption) (*NetworkMessage, error) {
	out := new(NetworkMessage)
	err := c.cc.Invoke(ctx, "/messages.MessageHandlers/HealthCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageHandlersClient) SkipPathGen(ctx context.Context, in *SkipPathGenMessage, opts ...grpc.CallOption) (*NetworkMessage, error) {
	out := new(NetworkMessage)
	err := c.cc.Invoke(ctx, "/messages.MessageHandlers/SkipPathGen", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessageHandlersServer is the server API for MessageHandlers service.
// All implementations must embed UnimplementedMessageHandlersServer
// for forward compatibility
type MessageHandlersServer interface {
	HandleSignedMessage(context.Context, *NetworkMessage) (*NetworkMessage, error)
	HandleSignedMessageStream(MessageHandlers_HandleSignedMessageStreamServer) error
	HealthCheck(context.Context, *NetworkMessage) (*NetworkMessage, error)
	SkipPathGen(context.Context, *SkipPathGenMessage) (*NetworkMessage, error)
	mustEmbedUnimplementedMessageHandlersServer()
}

// UnimplementedMessageHandlersServer must be embedded to have forward compatible implementations.
type UnimplementedMessageHandlersServer struct {
}

func (UnimplementedMessageHandlersServer) HandleSignedMessage(context.Context, *NetworkMessage) (*NetworkMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleSignedMessage not implemented")
}
func (UnimplementedMessageHandlersServer) HandleSignedMessageStream(MessageHandlers_HandleSignedMessageStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method HandleSignedMessageStream not implemented")
}
func (UnimplementedMessageHandlersServer) HealthCheck(context.Context, *NetworkMessage) (*NetworkMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HealthCheck not implemented")
}
func (UnimplementedMessageHandlersServer) SkipPathGen(context.Context, *SkipPathGenMessage) (*NetworkMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SkipPathGen not implemented")
}
func (UnimplementedMessageHandlersServer) mustEmbedUnimplementedMessageHandlersServer() {}

// UnsafeMessageHandlersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessageHandlersServer will
// result in compilation errors.
type UnsafeMessageHandlersServer interface {
	mustEmbedUnimplementedMessageHandlersServer()
}

func RegisterMessageHandlersServer(s grpc.ServiceRegistrar, srv MessageHandlersServer) {
	s.RegisterService(&MessageHandlers_ServiceDesc, srv)
}

func _MessageHandlers_HandleSignedMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageHandlersServer).HandleSignedMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messages.MessageHandlers/HandleSignedMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageHandlersServer).HandleSignedMessage(ctx, req.(*NetworkMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageHandlers_HandleSignedMessageStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MessageHandlersServer).HandleSignedMessageStream(&messageHandlersHandleSignedMessageStreamServer{stream})
}

type MessageHandlers_HandleSignedMessageStreamServer interface {
	Send(*NetworkMessage) error
	Recv() (*NetworkMessage, error)
	grpc.ServerStream
}

type messageHandlersHandleSignedMessageStreamServer struct {
	grpc.ServerStream
}

func (x *messageHandlersHandleSignedMessageStreamServer) Send(m *NetworkMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *messageHandlersHandleSignedMessageStreamServer) Recv() (*NetworkMessage, error) {
	m := new(NetworkMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _MessageHandlers_HealthCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageHandlersServer).HealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messages.MessageHandlers/HealthCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageHandlersServer).HealthCheck(ctx, req.(*NetworkMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageHandlers_SkipPathGen_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SkipPathGenMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageHandlersServer).SkipPathGen(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messages.MessageHandlers/SkipPathGen",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageHandlersServer).SkipPathGen(ctx, req.(*SkipPathGenMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// MessageHandlers_ServiceDesc is the grpc.ServiceDesc for MessageHandlers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MessageHandlers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "messages.MessageHandlers",
	HandlerType: (*MessageHandlersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleSignedMessage",
			Handler:    _MessageHandlers_HandleSignedMessage_Handler,
		},
		{
			MethodName: "HealthCheck",
			Handler:    _MessageHandlers_HealthCheck_Handler,
		},
		{
			MethodName: "SkipPathGen",
			Handler:    _MessageHandlers_SkipPathGen_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "HandleSignedMessageStream",
			Handler:       _MessageHandlers_HandleSignedMessageStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "messages.proto",
}
