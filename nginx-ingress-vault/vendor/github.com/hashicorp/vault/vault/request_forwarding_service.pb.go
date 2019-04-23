// Code generated by protoc-gen-go.
// source: request_forwarding_service.proto
// DO NOT EDIT!

/*
Package vault is a generated protocol buffer package.

It is generated from these files:
	request_forwarding_service.proto

It has these top-level messages:
*/
package vault

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import forwarding "github.com/hashicorp/vault/helper/forwarding"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for RequestForwarding service

type RequestForwardingClient interface {
	ForwardRequest(ctx context.Context, in *forwarding.Request, opts ...grpc.CallOption) (*forwarding.Response, error)
}

type requestForwardingClient struct {
	cc *grpc.ClientConn
}

func NewRequestForwardingClient(cc *grpc.ClientConn) RequestForwardingClient {
	return &requestForwardingClient{cc}
}

func (c *requestForwardingClient) ForwardRequest(ctx context.Context, in *forwarding.Request, opts ...grpc.CallOption) (*forwarding.Response, error) {
	out := new(forwarding.Response)
	err := grpc.Invoke(ctx, "/vault.RequestForwarding/ForwardRequest", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RequestForwarding service

type RequestForwardingServer interface {
	ForwardRequest(context.Context, *forwarding.Request) (*forwarding.Response, error)
}

func RegisterRequestForwardingServer(s *grpc.Server, srv RequestForwardingServer) {
	s.RegisterService(&_RequestForwarding_serviceDesc, srv)
}

func _RequestForwarding_ForwardRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(forwarding.Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RequestForwardingServer).ForwardRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vault.RequestForwarding/ForwardRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RequestForwardingServer).ForwardRequest(ctx, req.(*forwarding.Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _RequestForwarding_serviceDesc = grpc.ServiceDesc{
	ServiceName: "vault.RequestForwarding",
	HandlerType: (*RequestForwardingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ForwardRequest",
			Handler:    _RequestForwarding_ForwardRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "request_forwarding_service.proto",
}

func init() { proto.RegisterFile("request_forwarding_service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 151 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x28, 0x4a, 0x2d, 0x2c,
	0x4d, 0x2d, 0x2e, 0x89, 0x4f, 0xcb, 0x2f, 0x2a, 0x4f, 0x2c, 0x4a, 0xc9, 0xcc, 0x4b, 0x8f, 0x2f,
	0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2d, 0x4b,
	0x2c, 0xcd, 0x29, 0x91, 0xb2, 0x48, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5,
	0xcf, 0x48, 0x2c, 0xce, 0xc8, 0x4c, 0xce, 0x2f, 0x2a, 0xd0, 0x07, 0xcb, 0xe9, 0x67, 0xa4, 0xe6,
	0x14, 0xa4, 0x16, 0xe9, 0x23, 0x8c, 0xd0, 0x2f, 0xa9, 0x2c, 0x48, 0x2d, 0x86, 0x18, 0x60, 0x14,
	0xc4, 0x25, 0x18, 0x04, 0xb1, 0xc4, 0x0d, 0xae, 0x40, 0xc8, 0x96, 0x8b, 0x0f, 0xca, 0x83, 0xca,
	0x09, 0x09, 0xeb, 0x21, 0xf4, 0xeb, 0x41, 0x05, 0xa5, 0x44, 0x50, 0x05, 0x8b, 0x0b, 0xf2, 0xf3,
	0x8a, 0x53, 0x95, 0x18, 0x92, 0xd8, 0xc0, 0x46, 0x1b, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x81,
	0xce, 0x3f, 0x7f, 0xbf, 0x00, 0x00, 0x00,
}
