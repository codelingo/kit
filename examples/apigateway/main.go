package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	stdopentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc"
	addsvcgrpcclient "github.com/go-kit/kit/examples/addsvc/client/grpc"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr   = flag.String("consul.addr", "", "Consul agent address")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	// Context domain.
	ctx := context.Background()

	// Set up our routes.
	//
	// Each Consul service name resolves to multiple instances of that service.
	// We connect to each instance according to its pre-determined transport: in
	// this case, we choose to access addsvc via its gRPC client, and stringsvc
	// over plain transport/http (as it has no client package).
	//
	// Each service instance implements multiple methods, and we want to map
	// each method to a unique path on the API gateway. So, we define that path
	// and its corresponding factory function, which takes an instance string
	// and returns an endpoint.Endpoint for the specific method.
	//
	// Finally, we mount that path + endpoint handler into the router.
	r := mux.NewRouter()
	for consulServiceName, methods := range map[string][]struct {
		tags        []string
		passingOnly bool
		path        string
		factory     sd.Factory
	}{
		"addsvc": {
			{
				tags:        []string{}, // you might have e.g. "prod"
				passingOnly: true,
				path:        "/api/addsvc/sum",
				factory:     grpcFactory(addsvc.MakeSumEndpoint, logger),
			},
			{
				tags:        []string{},
				passingOnly: true,
				path:        "/api/addsvc/concat",
				factory:     grpcFactory(addsvc.MakeConcatEndpoint, logger),
			},
		},
		"stringsvc": {
			{
				tags:        []string{},
				passingOnly: true,
				path:        "/api/stringsvc/uppercase",
				factory:     httpFactory(ctx, "GET", "uppercase/"),
			},
			{
				tags:        []string{},
				passingOnly: true,
				path:        "/api/stringsvc/concat",
				factory:     httpFactory(ctx, "GET", "concat/"),
			},
		},
	} {
		for _, method := range methods {
			subscriber, err := consulsd.NewSubscriber(
				client,
				method.factory,
				logger,
				consulServiceName,
				method.tags,
				method.passingOnly,
			)
			if err != nil {
				logger.Log("service", consulServiceName, "path", method.path, "err", err)
				continue
			}

			balancer := lb.NewRoundRobin(subscriber)
			endpoint := lb.Retry(*retryMax, *retryTimeout, balancer)
			handler := makeHandler(ctx, endpoint, logger)
			r.HandleFunc(method.path, handler)
		}
	}

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func grpcFactory(makeEndpoint func(addsvc.Service) endpoint.Endpoint, logger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}

		tracer := stdopentracing.GlobalTracer() // no-op tracer
		service := addsvcgrpcclient.New(conn, tracer, logger)
		endpoint := makeEndpoint(service)

		// Notice that the addsvc gRPC client converts the connection to a
		// complete addsvc, and we just throw away everything except the method
		// we're interested in. A smarter factory would mux multiple methods
		// over the same connection. But that would require more work to manage
		// the returned io.Closer, e.g. reference counting. Since this is for
		// the purposes of demonstration, we'll just keep it simple.

		return endpoint, conn, nil
	}
}

// TODO(pb): -- refactoring from this line ---

func httpFactory(ctx context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}

		u, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		u.Path = path

		return httptransport.NewClient(
			method,
			u,
			passEncode,
			passDecode,
		).Endpoint(), nil, nil
	}
}

func makeHandler(ctx context.Context, e endpoint.Endpoint, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := e(ctx, r.Body)
		if err != nil {
			logger.Log("err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b, ok := resp.([]byte)
		if !ok {
			logger.Log("err", "endpoint response is not of type []byte")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			logger.Log("err", err)
			return
		}
	}
}

func passEncode(_ context.Context, r *http.Request, request interface{}) error {
	r.Body = request.(io.ReadCloser)
	return nil
}

func passDecode(_ context.Context, r *http.Response) (interface{}, error) {
	return ioutil.ReadAll(r.Body)
}
