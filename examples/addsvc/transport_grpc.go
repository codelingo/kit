package addsvc

// This file provides server-side bindings for the gRPC transport.
// It utilizes the transport/grpc.Server.

import (
	stdopentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// MakeGRPCServer makes a set of endpoints available as a gRPC AddServer.
func MakeGRPCServer(ctx context.Context, endpoints Endpoints, tracer stdopentracing.Tracer, logger log.Logger) pb.AddServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}
	return &grpcServer{
		sum: grpctransport.NewServer(
			ctx,
			endpoints.SumEndpoint,
			DecodeGRPCSumRequest,
			EncodeGRPCSumResponse,
			append(options, grpctransport.ServerBefore(opentracing.FromGRPCRequest(tracer, "Sum", logger)))...,
		),
		concat: grpctransport.NewServer(
			ctx,
			endpoints.ConcatEndpoint,
			DecodeGRPCConcatRequest,
			EncodeGRPCConcatResponse,
			append(options, grpctransport.ServerBefore(opentracing.FromGRPCRequest(tracer, "Concat", logger)))...,
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
