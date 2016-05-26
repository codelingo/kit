package grpc

import (
	"github.com/go-kit/kit/examples/addsvc2"
	"github.com/go-kit/kit/examples/addsvc2/pb"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

// New returns an AddService backed by a gRPC client connection. It is the
// responsibility of the caller to dial, and later close, the connection.
func New(conn *grpc.ClientConn) addsvc.Service {
	// TODO(pb): wire in some client-side middlewares

	sumEndpoint := grpctransport.NewClient(
		conn,
		"Add",
		"Sum",
		addsvc.EncodeGRPCSumRequest,
		addsvc.DecodeGRPCSumResponse,
		pb.SumReply{},
	).Endpoint()

	concatEndpoint := grpctransport.NewClient(
		conn,
		"Add",
		"Concat",
		addsvc.EncodeGRPCConcatRequest,
		addsvc.DecodeGRPCConcatResponse,
		pb.ConcatReply{},
	).Endpoint()

	return addsvc.Endpoints{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}
