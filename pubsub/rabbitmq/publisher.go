package rabbitmq

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/streadway/amqp"

	"github.com/codelingo/kit/pubsub"
)

type publisher struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
}

// Publish implements github.com/go-kit/kit/pubsub.Publisher
func (p *publisher) Publish(key string, r io.Reader) error {
	if p.ch == nil {
		return errors.New("channel is nil")
	}

	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return p.ch.Publish(
		p.exchange, // exchange
		key,        // key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        content,
		},
	)
}

// Stop implements github.com/go-kit/kit/pubsub.Publisher
func (p *publisher) Stop() error {
	var err error
	if p.ch != nil {
		err = p.ch.Close()
		p.ch = nil
	}

	// need to try and close the connection either way.
	if p.conn != nil {
		err = p.conn.Close()
		p.conn = nil
	}

	return err
}

// NewPublisher will return a new Publisher that will connect to the given
// rabbitmq server, and identify it's pubsub with the provided exchange
// name
//
// TODO: PublisherOptions for configuring the exchange, and publish
// parameters
func NewPublisher(url, exchange string) (pubsub.Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		// close the connection
		conn.Close()
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		// close the channel and the connection
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &publisher{
		conn:     conn,
		ch:       ch,
		exchange: exchange,
	}, nil
}
