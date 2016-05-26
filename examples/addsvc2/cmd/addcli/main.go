package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/go-kit/kit/examples/addsvc2"
	grpcclient "github.com/go-kit/kit/examples/addsvc2/client/grpc"
	httpclient "github.com/go-kit/kit/examples/addsvc2/client/http"
)

func main() {
	grpclog.SetLogger(log.New(os.Stdout, "", 0)) // Ugh! Bad gRPC!

	var (
		httpAddr = flag.String("http.addr", "", "HTTP address of addsvc")
		grpcAddr = flag.String("grpc.addr", "", "gRPC (HTTP) address of addsvc")
		method   = flag.String("method", "sum", "sum, concat")
	)
	flag.Parse()

	// This is a demonstration client, which supports multiple transports.
	// Your clients will probably just define and stick with 1 transport.

	if len(flag.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "error: all methods require 2 arguments\n")
		os.Exit(1)
	}

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
