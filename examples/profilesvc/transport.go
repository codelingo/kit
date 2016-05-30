package profilesvc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func MakeHTTPHandler(ctx context.Context, s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// POST    /profiles                           adds another profile
	// GET     /profiles/:id                       retrieves the given profile by id
	// PUT     /profiles/:id                       post updated profile information about the profile
	// PATCH   /profiles/:id                       partial updated profile information
	// DELETE  /profiles/:id                       remove the given profile
	// GET     /profiles/:id/addresses             retrieve addresses associated with the profile
	// GET     /profiles/:id/addresses/:addressID  retrieve a particular profile address
	// POST    /profiles/:id/addresses             add a new address
	// DELETE  /profiles/:id/addresses/:addressID  remove an address

	r.Methods("POST").Path("/profiles/").Handler(httptransport.NewServer(
		ctx,
		e.PostProfileEndpoint,
		decodePostProfileRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/profiles/{id}").Handler(httptransport.NewServer(
		ctx,
		e.GetProfileEndpoint,
		decodeGetProfileRequest,
		encodeResponse,
		options...,
	))
	r.Methods("PUT").Path("/profiles/{id}").Handler(httptransport.NewServer(
		ctx,
		e.PutProfileEndpoint,
		decodePutProfileRequest,
		encodeResponse,
		options...,
	))
	r.Methods("PATCH").Path("/profiles/{id}").Handler(httptransport.NewServer(
		ctx,
		e.PatchProfileEndpoint,
		decodePatchProfileRequest,
		encodeResponse,
		options...,
	))
	r.Methods("DELETE").Path("/profiles/{id}").Handler(httptransport.NewServer(
		ctx,
		e.DeleteProfileEndpoint,
		decodeDeleteProfileRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/profiles/{id}/addresses/").Handler(httptransport.NewServer(
		ctx,
		e.GetAddressesEndpoint,
		decodeGetAddressesRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/profiles/{id}/addresses/{addressID}").Handler(httptransport.NewServer(
		ctx,
		e.GetAddressEndpoint,
		decodeGetAddressRequest,
		encodeResponse,
		options...,
	))
	r.Methods("POST").Path("/profiles/{id}/addresses/").Handler(httptransport.NewServer(
		ctx,
		e.PostAddressEndpoint,
		decodePostAddressRequest,
		encodeResponse,
		options...,
	))
	r.Methods("DELETE").Path("/profiles/{id}/addresses/{addressID}").Handler(httptransport.NewServer(
		ctx,
		e.DeleteAddressEndpoint,
		decodeDeleteAddressRequest,
		encodeResponse,
		options...,
	))
	return r
}

func decodePostProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req postProfileRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Profile); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeGetProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return getProfileRequest{ID: id}, nil
}

func decodePutProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return putProfileRequest{
		ID:      id,
		Profile: profile,
	}, nil
}

func decodePatchProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return patchProfileRequest{
		ID:      id,
		Profile: profile,
	}, nil
}

func decodeDeleteProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return deleteProfileRequest{ID: id}, nil
}

func decodeGetAddressesRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return getAddressesRequest{ProfileID: id}, nil
}

func decodeGetAddressRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	addressID, ok := vars["addressID"]
	if !ok {
		return nil, ErrBadRouting
	}
	return getAddressRequest{
		ProfileID: id,
		AddressID: addressID,
	}, nil
}

func decodePostAddressRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	var address Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		return nil, err
	}
	return postAddressRequest{
		ProfileID: id,
		Address:   address,
	}, nil
}

func decodeDeleteAddressRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	addressID, ok := vars["addressID"]
	if !ok {
		return nil, ErrBadRouting
	}
	return deleteAddressRequest{
		ProfileID: id,
		AddressID: addressID,
	}, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		if e, ok := err.(httptransport.Error); ok {
			switch e.Domain {
			case httptransport.DomainDecode:
				return http.StatusBadRequest
			case httptransport.DomainDo:
				return http.StatusServiceUnavailable
			default:
				return http.StatusInternalServerError
			}
		}
		return http.StatusInternalServerError
	}
}
