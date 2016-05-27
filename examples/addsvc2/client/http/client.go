// Package http provides an HTTP client for the add service.
package http

import (
	"net/url"
	"strings"

	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
)

// New returns an AddService backed by an HTTP server living at the remote
// instance. We expect instance to come from a service discovery system, so
// likely of the form "host:port".
func New(instance string, tracer stdopentracing.Tracer, logger log.Logger) (addsvc.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// TODO(pb): wire in some client-side middlewares

	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/sum"),
			addsvc.EncodeHTTPGenericRequest,
			addsvc.DecodeHTTPSumResponse,
			httptransport.SetClientBefore(opentracing.FromHTTPRequest(tracer, "Sum", logger)),
		).Endpoint()
		sumEndpoint = opentracing.TraceClient(tracer, "Sum")(sumEndpoint)
	}

	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/concat"),
			addsvc.EncodeHTTPGenericRequest,
			addsvc.DecodeHTTPConcatResponse,
			httptransport.SetClientBefore(opentracing.FromHTTPRequest(tracer, "Concat", logger)),
		).Endpoint()
		concatEndpoint = opentracing.TraceClient(tracer, "Concat")(concatEndpoint)
	}

	return addsvc.Endpoints{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}
