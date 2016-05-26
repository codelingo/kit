package addsvc

import (
	"errors"

	"golang.org/x/net/context"
)

// This file has the Service definition, and a basic service implementation.

// Service describes a service that adds things together.
type Service interface {
	Sum(ctx context.Context, a, b int) (int, error)
	Concat(ctx context.Context, a, b string) (string, error)
}

// ErrTwoZeroes is an arbitrary business rule.
var ErrTwoZeroes = errors.New("can't sum two zeroes")

// NewBasicService returns a naïve, stateless implementation of Service.
func NewBasicService() Service {
	return basicService{}
}

type basicService struct{}

// Sum implements Service.
func (s basicService) Sum(_ context.Context, a, b int) (int, error) {
	if a == 0 && b == 0 {
		return 0, ErrTwoZeroes
	}
	return a + b, nil
}

// Concat implements Service.
func (s basicService) Concat(_ context.Context, a, b string) (string, error) {
	return a + b, nil
}
