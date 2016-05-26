package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc2"
	"github.com/go-kit/kit/examples/addsvc2/pb"
	thriftadd "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		httpAddr         = flag.String("http.addr", ":8080", "HTTP listen address")
		grpcAddr         = flag.String("grpc.addr", ":8081", "gRPC (HTTP) listen address")
		thriftAddr       = flag.String("thrift.addr", ":8082", "Thrift listen address")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")
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

	// HTTP transport
	go func() {
		logger := log.NewContext(logger).With("transport", "HTTP")
		h := addsvc.MakeHTTPHandler(ctx, endpoints, logger)
		logger.Log("addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, h)
	}()

	// gRPC transport
	go func() {
		logger := log.NewContext(logger).With("transport", "gRPC")

		ln, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			errc <- err
			return
		}

		srv := addsvc.MakeGRPCServer(ctx, endpoints, logger)
		s := grpc.NewServer()
		pb.RegisterAddServer(s, srv)

		logger.Log("addr", *grpcAddr)
		errc <- s.Serve(ln)
	}()

	// Thrift transport
	go func() {
		logger := log.NewContext(logger).With("transport", "Thrift")

		var protocolFactory thrift.TProtocolFactory
		switch *thriftProtocol {
		case "binary":
			protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
		case "compact":
			protocolFactory = thrift.NewTCompactProtocolFactory()
		case "json":
			protocolFactory = thrift.NewTJSONProtocolFactory()
		case "simplejson":
			protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
		default:
			errc <- fmt.Errorf("invalid Thrift protocol %q", *thriftProtocol)
			return
		}

		var transportFactory thrift.TTransportFactory
		if *thriftBufferSize > 0 {
			transportFactory = thrift.NewTBufferedTransportFactory(*thriftBufferSize)
		} else {
			transportFactory = thrift.NewTTransportFactory()
		}
		if *thriftFramed {
			transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
		}

		transport, err := thrift.NewTServerSocket(*thriftAddr)
		if err != nil {
			errc <- err
			return
		}

		logger.Log("addr", *thriftAddr)
		errc <- thrift.NewTSimpleServer4(
			thriftadd.NewAddServiceProcessor(addsvc.MakeThriftHandler(ctx, endpoints)),
			transport,
			transportFactory,
			protocolFactory,
		).Serve()
	}()

	// Run
	logger.Log("exit", <-errc)
}
