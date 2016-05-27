package addsvc

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// This file has the Service definition, and a basic service implementation.
// It also includes service middlewares.

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

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// ServiceLoggingMiddleware returns a service middleware that logs the
// parameters and result of each method invocation.
func ServiceLoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return serviceLoggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type serviceLoggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw serviceLoggingMiddleware) Sum(ctx context.Context, a, b int) (v int, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Sum",
			"a", a, "b", b, "result", v, "error", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Sum(ctx, a, b)
}

func (mw serviceLoggingMiddleware) Concat(ctx context.Context, a, b string) (v string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Concat",
			"a", a, "b", b, "result", v, "error", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Concat(ctx, a, b)
}

// ServiceInstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and character concatenated over the lifetime of
// the service.
func ServiceInstrumentingMiddleware(integersSummed, charactersConcatenated metrics.Counter) Middleware {
	return func(next Service) Service {
		return serviceInstrumentingMiddleware{
			ints:  integersSummed,
			chars: charactersConcatenated,
			next:  next,
		}
	}
}

type serviceInstrumentingMiddleware struct {
	ints  metrics.Counter
	chars metrics.Counter
	next  Service
}

func (mw serviceInstrumentingMiddleware) Sum(ctx context.Context, a, b int) (int, error) {
	v, err := mw.next.Sum(ctx, a, b)
	mw.ints.Add(uint64(v))
	return v, err
}

func (mw serviceInstrumentingMiddleware) Concat(ctx context.Context, a, b string) (string, error) {
	v, err := mw.next.Concat(ctx, a, b)
	mw.chars.Add(uint64(len(v)))
	return v, err
}
