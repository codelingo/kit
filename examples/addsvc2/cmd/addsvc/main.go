package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc2"
	"github.com/go-kit/kit/examples/addsvc2/pb"
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
		grpcAddr = flag.String("grpc.addr", ":8081", "gRPC (HTTP) listen address")
	)
	flag.Parse()

	// Logging domain
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}
	logger.Log("msg", "hello")
	defer logger.Log("msg", "goodbye")

	// Business domain
	var service addsvc.Service
	{
		service = addsvc.NewBasicService()
		// TODO(pb): service middlewares
	}

	// Endpoint domain
	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = addsvc.MakeSumEndpoint(service)
		// TODO(pb): endpoint middlewares
	}
	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = addsvc.MakeConcatEndpoint(service)
		// TODO(pb): endpoint middlewares
	}
	endpoints := addsvc.Endpoints{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}

	// Mechanical domain
	errc := make(chan error)
	ctx := context.Background()

	// Interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Transport domain
	go func() {
		logger = log.NewContext(logger).With("transport", "HTTP")
		h := addsvc.MakeHTTPHandler(ctx, endpoints, logger)
		logger.Log("transport", "HTTP", "addr", *httpAddr, "msg", "listening")
		errc <- http.ListenAndServe(*httpAddr, h)
	}()
	go func() {
		ln, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			errc <- err
			return
		}
		logger = log.NewContext(logger).With("transport", "gRPC")
		srv := addsvc.MakeGRPCServer(ctx, endpoints, logger)
		s := grpc.NewServer()
		pb.RegisterAddServer(s, srv)
		logger.Log("transport", "gRPC", "addr", *grpcAddr, "msg", "listening")
		errc <- s.Serve(ln)
	}()
	// Run
	logger.Log("terminating", <-errc)
}
