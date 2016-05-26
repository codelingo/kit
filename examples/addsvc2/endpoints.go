package addsvc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc2/pb"
)

// This file contains methods to make individual endpoints from services,
// request and response types to serve those endpoints, as well as encoders and
// decoders for those types, for all of our supported transport serialization
// formats.

// Endpoints collects all of the endpoints that compose an AddService. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them
// into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	SumEndpoint    endpoint.Endpoint
	ConcatEndpoint endpoint.Endpoint
}

// Sum implements Service. Primarily useful in a client.
func (e Endpoints) Sum(ctx context.Context, a, b int) (int, error) {
	request := sumRequest{A: a, B: b}
	response, err := e.SumEndpoint(ctx, request)
	if err != nil {
		return 0, err
	}
	return response.(sumResponse).V, nil
}

// Concat implements Service. Primarily useful in a client.
func (e Endpoints) Concat(ctx context.Context, a, b string) (string, error) {
	request := concatRequest{A: a, B: b}
	response, err := e.ConcatEndpoint(ctx, request)
	if err != nil {
		return "", err
	}
	return response.(concatResponse).V, err
}

// MakeSumEndpoint returns an endpoint that invokes Sum on the service.
// Primarily useful in a server.
func MakeSumEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		sumReq := request.(sumRequest)
		v, err := s.Sum(ctx, sumReq.A, sumReq.B)
		if err != nil {
			return nil, err
		}
		return sumResponse{
			V: v,
		}, nil
	}
}

// MakeConcatEndpoint returns an endpoint that invokes Concat on the service.
// Primarily useful in a server.
func MakeConcatEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		concatReq := request.(concatRequest)
		v, err := s.Concat(ctx, concatReq.A, concatReq.B)
		if err != nil {
			return nil, err
		}
		return concatResponse{
			V: v,
		}, nil
	}
}

// These types are unexported because they only exist to serve the endpoint
// domain, and are otherwise opaque to all callers.

type sumRequest struct{ A, B int }

type sumResponse struct{ V int }

type concatRequest struct{ A, B string }

type concatResponse struct{ V string }

// DecodeHTTPSumRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded sum request from the request body. Primarily useful in a server.
func DecodeHTTPSumRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req sumRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// DecodeHTTPConcatRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded concat request from the request body. Primarily useful in a
// server.
func DecodeHTTPConcatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req concatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// DecodeHTTPSumResponse is a transport/http.DecodeResponseFunc that decodes a
// JSON-encoded sum response from the response body. If the response has a
// non-200 status code, we will interpret that as an error and attempt to decode
// the specific error message from the response body. Primarily useful in a
// client.
func DecodeHTTPSumResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errorDecoder(r)
	}
	var resp sumResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// DecodeHTTPConcatResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the response body. If the response has a
// non-200 status code, we will interpret that as an error and attempt to decode
// the specific error message from the response body. Primarily useful in a
// client.
func DecodeHTTPConcatResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errorDecoder(r)
	}
	var resp concatResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// DecodeGRPCSumRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func DecodeGRPCSumRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SumRequest)
	return sumRequest{A: int(req.A), B: int(req.B)}, nil
}

// DecodeGRPCConcatRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC concat request to a user-domain concat request. Primarily useful in a
// server.
func DecodeGRPCConcatRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.ConcatRequest)
	return concatRequest{A: req.A, B: req.B}, nil
}

// DecodeGRPCSumResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC sum reply to a user-domain sum response. Primarily useful in a client.
func DecodeGRPCSumResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.SumReply)
	return sumResponse{V: int(reply.V)}, nil
}

// DecodeGRPCConcatResponse is a transport/grpc.DecodeResponseFunc that converts
// a gRPC concat reply to a user-domain concat response. Primarily useful in a
// client.
func DecodeGRPCConcatResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.ConcatReply)
	return concatResponse{V: reply.V}, nil
}

// EncodeGRPCSumResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain sum response to a gRPC sum reply. Primarily useful in a server.
func EncodeGRPCSumResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(sumResponse)
	return &pb.SumReply{V: int64(resp.V)}, nil
}

// EncodeGRPCConcatResponse is a transport/grpc.EncodeResponseFunc that converts
// a user-domain concat response to a gRPC concat reply. Primarily useful in a
// server.
func EncodeGRPCConcatResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(concatResponse)
	return &pb.ConcatReply{V: resp.V}, nil
}

// EncodeGRPCSumRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain sum request to a gRPC sum request. Primarily useful in a client.
func EncodeGRPCSumRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(sumRequest)
	return &pb.SumRequest{A: int64(req.A), B: int64(req.B)}, nil
}

// EncodeGRPCConcatRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain concat request to a gRPC concat request. Primarily useful in a
// client.
func EncodeGRPCConcatRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(concatRequest)
	return &pb.ConcatRequest{A: req.A, B: req.B}, nil
}

// EncodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func EncodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGenericResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
