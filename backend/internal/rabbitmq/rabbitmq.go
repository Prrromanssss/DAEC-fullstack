package rabbitmq

import (
	"log/slog"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/sl"
	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	log  *slog.Logger
	conn *amqp.Connection
}

func NewAMQPConfig(log *slog.Logger, amqpUrl string) (*AMQPConfig, error) {
	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Error("can't connect to RabbitMQ", sl.Err(err))
		return nil, err
	}

	log.Info("successfully connected to RabbitMQ instance")

	return &AMQPConfig{
		log:  log,
		conn: conn,
	}, nil
}
