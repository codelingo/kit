package rabbitmq

import (
	"errors"

	"github.com/streadway/amqp"

	"github.com/go-kit/kit/pubsub"
)

type subscriber struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue

	err error
}

// Start implements github.com/go-kit/kit/pubsub.Subscriber
func (s *subscriber) Start() <-chan pubsub.Message {
	ch := make(chan pubsub.Message, 1)
	if s.ch == nil {
		s.err = errors.New("channel is nil")
		close(ch)
		return ch
	}

	msgs, err := s.ch.Consume(
		s.q.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		s.err = err
		close(ch)
		return ch
	}

	go func(msgs <-chan amqp.Delivery, ch chan pubsub.Message) {
		// should be safe, as closing the amqp channel seems like it
		// would close the channel that is feeding the delivery of
		// messages to the msgs channel created from calling Consume
		for d := range msgs {
			ch <- createMessage(d)
		}
		close(ch)
	}(msgs, ch)

	return ch
}

// Err implements github.com/go-kit/kit/pubsub.Subscriber
func (s *subscriber) Err() error {
	return s.err
}

// Stop implements github.com/go-kit/kit/pubsub.Subscriber
func (s *subscriber) Stop() error {
	var err error
	if s.ch != nil {
		err = s.ch.Close()
		s.ch = nil
	}

	// need to try and close the connection either way.
	if s.conn != nil {
		err = s.conn.Close()
		s.conn = nil
	}

	return err
}

// NewSubscriber returns a github.com/go-kit/kit/pubsub.Subscriber
// that will connect to the given rabbitmq, and will subscribe to
// the given exchange.
//
// TODO: Exchange, Queue, QueueBind, Consume Options
func NewSubscriber(url, exchange, key string) (s pubsub.Subscriber, err error) {
	var (
		conn *amqp.Connection
		ch   *amqp.Channel
	)

	defer func() {
		// generic cleanup
		if err != nil {
			if ch != nil {
				ch.Close()
			}
			if conn != nil {
				conn.Close()
			}
		}
	}()

	conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err = conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchange,
		"fanout",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,   // queue name
		key,      // routing key
		exchange, // exchange name
		false,    // no-wait
		nil,      // arguments
	)

	if err != nil {
		return nil, err
	}

	return &subscriber{
		conn: conn,
		ch:   ch,
		q:    q,
	}, nil
}
