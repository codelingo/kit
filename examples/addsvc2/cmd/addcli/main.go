package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/examples/addsvc2"
	grpcclient "github.com/go-kit/kit/examples/addsvc2/client/grpc"
	httpclient "github.com/go-kit/kit/examples/addsvc2/client/http"
	thriftclient "github.com/go-kit/kit/examples/addsvc2/client/thrift"
	thriftadd "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	grpclog.SetLogger(log.New(os.Stdout, "", 0)) // Ugh! Bad gRPC!

	var (
		httpAddr         = flag.String("http.addr", "", "HTTP address of addsvc")
		grpcAddr         = flag.String("grpc.addr", "", "gRPC (HTTP) address of addsvc")
		thriftAddr       = flag.String("thrift.addr", ":8082", "Thrift listen address")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")
		method           = flag.String("method", "sum", "sum, concat")
	)
	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "error: all methods require 2 arguments\n")
		os.Exit(1)
	}

	// This is a demonstration client, which supports multiple transports.
	// Your clients will probably just define and stick with 1 transport.

	var (
		service addsvc.Service
		err     error
	)
	if *httpAddr != "" {
		service, err = httpclient.New(*httpAddr)
	} else if *grpcAddr != "" {
		conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
			os.Exit(1)
		}
		defer conn.Close()
		service = grpcclient.New(conn)
	} else if *thriftAddr != "" {
		// It's necessary to do all of this construction in the func main,
		// because (among other reasons) we need to control the lifecycle of the
		// Thrift transport, i.e. close it eventually.
		var protocolFactory thrift.TProtocolFactory
		switch *thriftProtocol {
		case "compact":
			protocolFactory = thrift.NewTCompactProtocolFactory()
		case "simplejson":
			protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
		case "json":
			protocolFactory = thrift.NewTJSONProtocolFactory()
		case "binary", "":
			protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
		default:
			fmt.Fprintf(os.Stderr, "error: invalid protocol %q\n", *thriftProtocol)
			os.Exit(1)
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
		transportSocket, err := thrift.NewTSocket(*thriftAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		transport := transportFactory.GetTransport(transportSocket)
		if err := transport.Open(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer transport.Close()
		client := thriftadd.NewAddServiceClientFactory(transport, protocolFactory)
		service = thriftclient.New(client)
	} else {
		fmt.Fprintf(os.Stderr, "error: no remote address specified\n")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch *method {
	case "sum":
		a, _ := strconv.ParseInt(flag.Args()[0], 10, 64)
		b, _ := strconv.ParseInt(flag.Args()[1], 10, 64)
		v, err := service.Sum(context.Background(), int(a), int(b))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%d + %d = %d\n", a, b, v)

	case "concat":
		a := flag.Args()[0]
		b := flag.Args()[1]
		v, err := service.Concat(context.Background(), a, b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%q + %q = %q\n", a, b, v)

	default:
		fmt.Fprintf(os.Stderr, "error: invalid method %q\n", method)
		os.Exit(1)
	}
}
