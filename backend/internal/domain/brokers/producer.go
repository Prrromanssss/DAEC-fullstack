package brokers

import (
	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEC-fullstack/internal/rabbitmq"
)

// Producer is an interface to produce messages to queue.
type Producer interface {
	PublishExpressionMessage(msg *messages.ExpressionMessage) error
	Reconnect() (*rabbitmq.AMQPProducer, error)
	Close()
}
