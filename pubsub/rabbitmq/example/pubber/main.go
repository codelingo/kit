package main

import (
	"bytes"
	"flag"
	"fmt"
	"time"

	"github.com/go-kit/kit/pubsub/rabbitmq"
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

	pub, err := rabbitmq.NewPublisher(args.url, args.exchange)
	if err != nil {
		panic(err)
	}

	fmt.Println(">>= press ctrl+c to kill this proces")

	var i = 1
	for {
		buf := new(bytes.Buffer)
		_, err := buf.WriteString(fmt.Sprintf("%s %d", "testing", i))
		i++
		if err != nil {
			panic(err)
		}

		err = pub.Publish("", buf)
		if err != nil {
			panic(err)
		}

		// put a delay in the publish
		time.Sleep(time.Second)
	}
}
