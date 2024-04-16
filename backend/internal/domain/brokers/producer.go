package brokers

import (
	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/rabbitmq"
)

type Producer interface {
	PublishExpressionMessage(msg *messages.ExpressionMessage) error
	Reconnect() (*rabbitmq.AMQPProducer, error)
	Close()
}
