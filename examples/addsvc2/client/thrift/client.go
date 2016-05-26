// Package thrift provides a Thrift client for the add service.
package thrift

import (
	"github.com/go-kit/kit/examples/addsvc2"
	thriftadd "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
)

// New returns an AddService backed by a Thrift server described by the provided
// client. The caller is responsible for constructing the client, and eventually
// closing the underlying transport.
func New(client *thriftadd.AddServiceClient) addsvc.Service {
	return addsvc.Endpoints{
		SumEndpoint:    addsvc.MakeThriftSumEndpoint(client),
		ConcatEndpoint: addsvc.MakeThriftConcatEndpoint(client),
	}
}
