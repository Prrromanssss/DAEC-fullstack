package brokers

import "github.com/streadway/amqp"

type Consumer interface {
	GetMessages() <-chan amqp.Delivery
	Close()
}
