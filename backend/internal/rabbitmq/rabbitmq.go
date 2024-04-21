package rabbitmq

import (
	"log/slog"

	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/sl"
	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	log  *slog.Logger
	conn *amqp.Connection
}

// NewAMQPConfig creates new AMQP connection.
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

// Close closes AMQP connection.
func (ac *AMQPConfig) Close() {
	ac.conn.Close()
}
