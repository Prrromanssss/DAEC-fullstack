package rabbitmq

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/sl"
	"github.com/streadway/amqp"
)

type AMQPProducer struct {
	log     *slog.Logger
	amqpCfg *AMQPConfig
	Queue   amqp.Queue
	Channel *amqp.Channel
}

// NewAMQPProducer creates new Producer for AMQP protocol.
func NewAMQPProducer(log *slog.Logger, amqpCfg *AMQPConfig, queueName string) (*AMQPProducer, error) {
	chProd, err := amqpCfg.conn.Channel()
	if err != nil {
		log.Error("can't create a channel from RabbitMQ", sl.Err(err))
		return nil, err
	}

	queue, err := chProd.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("can't create a RabbitMQ queue", sl.Err(err))
		return nil, err
	}

	return &AMQPProducer{
		log:     log,
		amqpCfg: amqpCfg,
		Queue:   queue,
		Channel: chProd,
	}, nil
}

// PublishExpressionMessage publishes messages to queue.
func (ap *AMQPProducer) PublishExpressionMessage(msg *messages.ExpressionMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return errors.New("failed to encode message to JSON")
	}

	err = ap.Channel.Publish(
		"",
		ap.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)

	if err != nil {
		ap.log.Error("can't publish message to queue", slog.String("queue", ap.Queue.Name), sl.Err(err))
		return errors.New("can't publish message to queue")
	}
	ap.log.Info("publishing message to queue", slog.String("queue", ap.Queue.Name))

	return nil
}

// Reconnect reconnects to AMQP instance.
func (ap *AMQPProducer) Reconnect() (*AMQPProducer, error) {
	ap.Close()

	return NewAMQPProducer(ap.log, ap.amqpCfg, ap.Queue.Name)
}

// Close closes Producer channel.
func (ap *AMQPProducer) Close() {
	ap.Channel.Close()
}
