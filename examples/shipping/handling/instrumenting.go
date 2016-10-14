package handling

import (
	"time"

	"github.com/codelingo/kit/metrics"

	"github.com/codelingo/kit/examples/shipping/cargo"
	"github.com/codelingo/kit/examples/shipping/location"
	"github.com/codelingo/kit/examples/shipping/voyage"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(requestCount metrics.Counter, requestLatency metrics.Histogram, s Service) Service {
	return &instrumentingService{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		Service:        s,
	}
}

func (s *instrumentingService) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyage voyage.Number,
	loc location.UNLocode, eventType cargo.HandlingEventType) error {

	defer func(begin time.Time) {
		s.requestCount.With("method", "register_incident").Add(1)
		s.requestLatency.With("method", "register_incident").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.RegisterHandlingEvent(completionTime, trackingID, voyage, loc, eventType)
}
