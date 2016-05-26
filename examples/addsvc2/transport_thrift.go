package addsvc

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	thriftadd "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
)

// This file provides server-side bindings for the Thrift transport.

// MakeThriftHandler makes a set of endpoints available as a Thrift service.
func MakeThriftHandler(ctx context.Context, e Endpoints) thriftadd.AddService {
	return &thriftServer{
		ctx:    ctx,
		sum:    e.SumEndpoint,
		concat: e.ConcatEndpoint,
	}
}

type thriftServer struct {
	ctx    context.Context
	sum    endpoint.Endpoint
	concat endpoint.Endpoint
}

func (s *thriftServer) Sum(a int64, b int64) (*thriftadd.SumReply, error) {
	request := sumRequest{A: int(a), B: int(b)}
	response, err := s.sum(s.ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(sumResponse)
	return &thriftadd.SumReply{Value: int64(resp.V)}, nil
}

func (s *thriftServer) Concat(a string, b string) (*thriftadd.ConcatReply, error) {
	request := concatRequest{A: a, B: b}
	response, err := s.concat(s.ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(concatResponse)
	return &thriftadd.ConcatReply{Value: resp.V}, nil
}
