package sd

import "github.com/codelingo/kit/endpoint"

// FixedSubscriber yields a fixed set of services.
type FixedSubscriber []endpoint.Endpoint

// Endpoints implements Subscriber.
func (s FixedSubscriber) Endpoints() ([]endpoint.Endpoint, error) { return s, nil }
