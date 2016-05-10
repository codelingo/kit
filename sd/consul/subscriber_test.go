package consul

import (
	"testing"

	consul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
)

var consulState = []*consul.ServiceEntry{
	{
		Node: &consul.Node{
			Address: "10.0.0.0",
			Node:    "app00.local",
		},
		Service: &consul.AgentService{
			ID:      "search-api-0",
			Port:    8000,
			Service: "search",
			Tags: []string{
				"api",
				"v1",
			},
		},
	},
	{
		Node: &consul.Node{
			Address: "10.0.0.1",
			Node:    "app01.local",
		},
		Service: &consul.AgentService{
			ID:      "search-api-1",
			Port:    8001,
			Service: "search",
			Tags: []string{
				"api",
				"v2",
			},
		},
	},
	{
		Node: &consul.Node{
			Address: "10.0.0.1",
			Node:    "app01.local",
		},
		Service: &consul.AgentService{
			Address: "10.0.0.10",
			ID:      "search-db-0",
			Port:    9000,
			Service: "search",
			Tags: []string{
				"db",
			},
		},
	},
}

func TestSubscriber(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s, err := NewSubscriber(client, testFactory, logger, "search", []string{"api"}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	eps, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 2, len(eps); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestSubscriberNoService(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s, err := NewSubscriber(client, testFactory, logger, "feed", []string{}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	services, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(services); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestSubscriberWithTags(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s, err := NewSubscriber(client, testFactory, logger, "search", []string{"api", "v2"}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	services, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 1, len(services); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestSubscriberAddressOverride(t *testing.T) {
	s, err := NewSubscriber(newTestClient(consulState), testFactory, log.NewNopLogger(), "search", []string{"db"}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	services, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 1, len(services); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	endpoint, err := services[0].Endpoint("irrelevant")
	if err != nil {
		t.Fatal(err)
	}

	response, err := endpoint(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "10.0.0.10:9000", response.(string); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}