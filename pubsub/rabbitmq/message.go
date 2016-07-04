package rabbitmq

import (
	"bytes"
	"io"

	"github.com/streadway/amqp"

	"github.com/go-kit/kit/pubsub"
)

// Message is a minimal interface to describe payloads received by subscribers.
// Clients may type-assert to more concrete types (e.g. pubsub/kafka.Message) to
// get access to more specific behaviors.
type Message interface {
	// Messages implement io.Reader to access the payload data.
	io.Reader

	// Done indicates the client is finished with the message, and the
	// underlying implementation may free its resources. Clients should ensure
	// to call Done for every received message.
	Done() error
}

type message struct {
	io.Reader

	// stuff for ack?
	// deliveryTag uint64
	// acker       amqp.Acknowledger
}

// Done implements github.com/go-kit/kit/pubsub.Message
func (m *message) Done() error {
	return nil
	// return m.acker.Ack(m.deliveryTag, false)
}

// createMessage will convert the amqp.Delivery to a pubsub.Message
func createMessage(d amqp.Delivery) pubsub.Message {
	return &message{
		Reader: bytes.NewBuffer(d.Body),
		// deliveryTag: d.DeliveryTag,
		// acker:       d.Acknowledger,
	}
}
