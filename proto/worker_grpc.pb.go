// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WorkerClient is the client API for Worker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkerClient interface {
	GetInfo(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Info, error)
	RunContainer(ctx context.Context, in *Container, opts ...grpc.CallOption) (*empty.Empty, error)
	CheckContainer(ctx context.Context, in *Container, opts ...grpc.CallOption) (*State, error)
	StopContainer(ctx context.Context, in *Container, opts ...grpc.CallOption) (*empty.Empty, error)
}

type workerClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerClient(cc grpc.ClientConnInterface) WorkerClient {
	return &workerClient{cc}
}

func (c *workerClient) GetInfo(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Info, error) {
	out := new(Info)
	err := c.cc.Invoke(ctx, "/proto.Worker/GetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerClient) RunContainer(ctx context.Context, in *Container, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/proto.Worker/RunContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerClient) CheckContainer(ctx context.Context, in *Container, opts ...grpc.CallOption) (*State, error) {
	out := new(State)
	err := c.cc.Invoke(ctx, "/proto.Worker/CheckContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerClient) StopContainer(ctx context.Context, in *Container, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/proto.Worker/StopContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerServer is the server API for Worker service.
// All implementations must embed UnimplementedWorkerServer
// for forward compatibility
type WorkerServer interface {
	GetInfo(context.Context, *empty.Empty) (*Info, error)
	RunContainer(context.Context, *Container) (*empty.Empty, error)
	CheckContainer(context.Context, *Container) (*State, error)
	StopContainer(context.Context, *Container) (*empty.Empty, error)
	mustEmbedUnimplementedWorkerServer()
}

// UnimplementedWorkerServer must be embedded to have forward compatible implementations.
type UnimplementedWorkerServer struct {
}

func (UnimplementedWorkerServer) GetInfo(context.Context, *empty.Empty) (*Info, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedWorkerServer) RunContainer(context.Context, *Container) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunContainer not implemented")
}
func (UnimplementedWorkerServer) CheckContainer(context.Context, *Container) (*State, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckContainer not implemented")
}
func (UnimplementedWorkerServer) StopContainer(context.Context, *Container) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopContainer not implemented")
}
func (UnimplementedWorkerServer) mustEmbedUnimplementedWorkerServer() {}

// UnsafeWorkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkerServer will
// result in compilation errors.
type UnsafeWorkerServer interface {
	mustEmbedUnimplementedWorkerServer()
}

func RegisterWorkerServer(s grpc.ServiceRegistrar, srv WorkerServer) {
	s.RegisterService(&Worker_ServiceDesc, srv)
}

func _Worker_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Worker/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).GetInfo(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Worker_RunContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Container)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).RunContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Worker/RunContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).RunContainer(ctx, req.(*Container))
	}
	return interceptor(ctx, in, info, handler)
}

func _Worker_CheckContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Container)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).CheckContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Worker/CheckContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).CheckContainer(ctx, req.(*Container))
	}
	return interceptor(ctx, in, info, handler)
}

func _Worker_StopContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Container)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).StopContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Worker/StopContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).StopContainer(ctx, req.(*Container))
	}
	return interceptor(ctx, in, info, handler)
}

// Worker_ServiceDesc is the grpc.ServiceDesc for Worker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Worker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Worker",
	HandlerType: (*WorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetInfo",
			Handler:    _Worker_GetInfo_Handler,
		},
		{
			MethodName: "RunContainer",
			Handler:    _Worker_RunContainer_Handler,
		},
		{
			MethodName: "CheckContainer",
			Handler:    _Worker_CheckContainer_Handler,
		},
		{
			MethodName: "StopContainer",
			Handler:    _Worker_StopContainer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/worker.proto",
}
