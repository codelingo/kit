package http

import (
	"net/url"
	"strings"

	"github.com/go-kit/kit/examples/addsvc2"
	httptransport "github.com/go-kit/kit/transport/http"
)

// New returns an AddService backed by an HTTP server living at the remote
// instance. We expect instance to come from a service discovery system, so
// likely of the form "host:port".
func New(instance string) (addsvc.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// TODO(pb): wire in some client-side middlewares

	sumEndpoint := httptransport.NewClient(
		"POST",
		copyURL(u, "/sum"),
		addsvc.EncodeHTTPGenericRequest,
		addsvc.DecodeHTTPSumResponse,
	).Endpoint()

	concatEndpoint := httptransport.NewClient(
		"POST",
		copyURL(u, "/concat"),
		addsvc.EncodeHTTPGenericRequest,
		addsvc.DecodeHTTPConcatResponse,
	).Endpoint()

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
