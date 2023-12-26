// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.1
// source: kvrpc.proto

package kv

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

// RaftKVClient is the client API for RaftKV service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftKVClient interface {
	AppendEntries(ctx context.Context, in *AppendEntriesArgs, opts ...grpc.CallOption) (*AppendEntriesReply, error)
	RequestVote(ctx context.Context, in *RequestVoteArgs, opts ...grpc.CallOption) (*RequestVoteReply, error)
	Operate(ctx context.Context, in *Operation, opts ...grpc.CallOption) (*Opreturn, error)
}

type raftKVClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftKVClient(cc grpc.ClientConnInterface) RaftKVClient {
	return &raftKVClient{cc}
}

func (c *raftKVClient) AppendEntries(ctx context.Context, in *AppendEntriesArgs, opts ...grpc.CallOption) (*AppendEntriesReply, error) {
	out := new(AppendEntriesReply)
	err := c.cc.Invoke(ctx, "/RaftKV/AppendEntries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftKVClient) RequestVote(ctx context.Context, in *RequestVoteArgs, opts ...grpc.CallOption) (*RequestVoteReply, error) {
	out := new(RequestVoteReply)
	err := c.cc.Invoke(ctx, "/RaftKV/RequestVote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftKVClient) Operate(ctx context.Context, in *Operation, opts ...grpc.CallOption) (*Opreturn, error) {
	out := new(Opreturn)
	err := c.cc.Invoke(ctx, "/RaftKV/Operate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftKVServer is the server API for RaftKV service.
// All implementations must embed UnimplementedRaftKVServer
// for forward compatibility
type RaftKVServer interface {
	AppendEntries(context.Context, *AppendEntriesArgs) (*AppendEntriesReply, error)
	RequestVote(context.Context, *RequestVoteArgs) (*RequestVoteReply, error)
	Operate(context.Context, *Operation) (*Opreturn, error)
	mustEmbedUnimplementedRaftKVServer()
}

// UnimplementedRaftKVServer must be embedded to have forward compatible implementations.
type UnimplementedRaftKVServer struct {
}

func (UnimplementedRaftKVServer) AppendEntries(context.Context, *AppendEntriesArgs) (*AppendEntriesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (UnimplementedRaftKVServer) RequestVote(context.Context, *RequestVoteArgs) (*RequestVoteReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedRaftKVServer) Operate(context.Context, *Operation) (*Opreturn, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Operate not implemented")
}
func (UnimplementedRaftKVServer) mustEmbedUnimplementedRaftKVServer() {}

// UnsafeRaftKVServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftKVServer will
// result in compilation errors.
type UnsafeRaftKVServer interface {
	mustEmbedUnimplementedRaftKVServer()
}

func RegisterRaftKVServer(s grpc.ServiceRegistrar, srv RaftKVServer) {
	s.RegisterService(&RaftKV_ServiceDesc, srv)
}

func _RaftKV_AppendEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendEntriesArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftKVServer).AppendEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RaftKV/AppendEntries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftKVServer).AppendEntries(ctx, req.(*AppendEntriesArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftKV_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestVoteArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftKVServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RaftKV/RequestVote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftKVServer).RequestVote(ctx, req.(*RequestVoteArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftKV_Operate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Operation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftKVServer).Operate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RaftKV/Operate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftKVServer).Operate(ctx, req.(*Operation))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftKV_ServiceDesc is the grpc.ServiceDesc for RaftKV service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftKV_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "RaftKV",
	HandlerType: (*RaftKVServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AppendEntries",
			Handler:    _RaftKV_AppendEntries_Handler,
		},
		{
			MethodName: "RequestVote",
			Handler:    _RaftKV_RequestVote_Handler,
		},
		{
			MethodName: "Operate",
			Handler:    _RaftKV_Operate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kvrpc.proto",
}