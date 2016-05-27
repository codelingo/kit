// Package grpc provides a gRPC client for the add service.
package grpc

import (
	stdopentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc2"
	"github.com/go-kit/kit/examples/addsvc2/pb"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// New returns an AddService backed by a gRPC client connection. It is the
// responsibility of the caller to dial, and later close, the connection.
func New(conn *grpc.ClientConn, tracer stdopentracing.Tracer, logger log.Logger) addsvc.Service {
	// TODO(pb): wire in some client-side middlewares

	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = grpctransport.NewClient(
			conn,
			"Add",
			"Sum",
			addsvc.EncodeGRPCSumRequest,
			addsvc.DecodeGRPCSumResponse,
			pb.SumReply{},
			grpctransport.SetClientBefore(opentracing.FromGRPCRequest(tracer, "Sum", logger)),
		).Endpoint()
		sumEndpoint = opentracing.TraceClient(tracer, "Sum")(sumEndpoint)
	}

	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = grpctransport.NewClient(
			conn,
			"Add",
			"Concat",
			addsvc.EncodeGRPCConcatRequest,
			addsvc.DecodeGRPCConcatResponse,
			pb.ConcatReply{},
			grpctransport.SetClientBefore(opentracing.FromGRPCRequest(tracer, "Concat", logger)),
		).Endpoint()
		concatEndpoint = opentracing.TraceClient(tracer, "Concat")(concatEndpoint)
	}

	return addsvc.Endpoints{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}
