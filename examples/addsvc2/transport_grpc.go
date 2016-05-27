package addsvc

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/examples/addsvc2/pb"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// This file provides server-side bindings for the gRPC transport.

// MakeGRPCServer makes a set of endpoints available as a gRPC AddServer.
func MakeGRPCServer(ctx context.Context, endpoints Endpoints, logger log.Logger) pb.AddServer {
	// TODO(pb): add tracer
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}
	return &grpcServer{
		sum: grpctransport.NewServer(
			ctx,
			endpoints.SumEndpoint,
			DecodeGRPCSumRequest,
			EncodeGRPCSumResponse,
			options...,
		),
		concat: grpctransport.NewServer(
			ctx,
			endpoints.ConcatEndpoint,
			DecodeGRPCConcatRequest,
			EncodeGRPCConcatResponse,
			options...,
		),
	}
}

type grpcServer struct {
	sum    grpctransport.Handler
	concat grpctransport.Handler
}

func (s *grpcServer) Sum(ctx context.Context, req *pb.SumRequest) (*pb.SumReply, error) {
	_, rep, err := s.sum.ServeGRPC(ctx, req)
	return rep.(*pb.SumReply), err
}

func (s *grpcServer) Concat(ctx context.Context, req *pb.ConcatRequest) (*pb.ConcatReply, error) {
	_, rep, err := s.concat.ServeGRPC(ctx, req)
	return rep.(*pb.ConcatReply), err
}
