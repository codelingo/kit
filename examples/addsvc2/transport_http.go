package addsvc

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

// This file provides server-side bindings for the HTTP transport.

// MakeHTTPHandler returns a handler that makes a set of endpoints available
// on predefined paths.
func MakeHTTPHandler(ctx context.Context, endpoints Endpoints, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}
	m := http.NewServeMux()
	m.Handle("/sum", httptransport.NewServer(
		ctx,
		endpoints.SumEndpoint,
		DecodeHTTPSumRequest,
		EncodeHTTPGenericResponse,
		options...,
	))
	m.Handle("/concat", httptransport.NewServer(
		ctx,
		endpoints.ConcatEndpoint,
		DecodeHTTPConcatRequest,
		EncodeHTTPGenericResponse,
		options...,
	))
	return m
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	msg := err.Error()

	if e, ok := err.(httptransport.Error); ok {
		msg = e.Err.Error()
		switch e.Domain {
		case httptransport.DomainDecode:
			code = http.StatusBadRequest

		case httptransport.DomainDo:
			if e.Err == ErrTwoZeroes {
				code = http.StatusBadRequest
			}
		}
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorWrapper{Error: msg})
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}
