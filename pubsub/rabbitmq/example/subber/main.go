package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/codelingo/kit/pubsub/rabbitmq"
)

type arguments struct {
	url      string
	exchange string
}

var args arguments

func init() {
	flag.StringVar(&args.url, "url", "amqp://guest:guest@localhost:5672/", "specify this to overwrite credentials, host, or port")
	flag.StringVar(&args.exchange, "exchange", "test", "this will change the exchange's name (think topic)")
}

func main() {
	flag.Parse()

	sub, err := rabbitmq.NewSubscriber(args.url, args.exchange, "")
	if err != nil {
		panic(err)
	}

	fmt.Println(">>= press ctrl+c to kill this proces")

	ch := sub.Start()

	for m := range ch {
		b, err := ioutil.ReadAll(m)
		if err != nil {
			panic(err)
		}

		// print the message as a string
		fmt.Printf("%s\n", b)
	}
}
