package brokers

import "github.com/streadway/amqp"

// Consumer is an interface to consume messages from queue.
type Consumer interface {
	GetMessages() <-chan amqp.Delivery
	Close()
}
