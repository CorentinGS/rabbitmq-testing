package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Queue struct {
	keys       []string
	Name       string
	Collection string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type RabbitMQConnection struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queues   map[string]Queue
	exchange string
}

func (connection *RabbitMQConnection) Set(conn *amqp.Connection, channel *amqp.Channel, queues map[string]Queue, exchange string) {
	connection.conn = conn
	connection.channel = channel
	connection.queues = queues
	connection.exchange = exchange
}

func (connection *RabbitMQConnection) SetQueues(queues []Queue) error {
	for _, q := range queues {
		connection.queues[q.Name] = q
		if _, err := connection.channel.QueueDeclare(q.Name, true, false, false, false, nil); err != nil {
			return err
		}
		for _, k := range q.keys {
			if err := connection.channel.QueueBind(q.Name, k, connection.exchange, false, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (connection *RabbitMQConnection) Consume() (map[string]<-chan amqp.Delivery, error) {
	m := make(map[string]<-chan amqp.Delivery)
	for _, q := range connection.queues {
		deliveries, err := connection.channel.Consume(q.Name, "", true, false, false, false, nil)
		if err != nil {
			return nil, err
		}
		m[q.Name] = deliveries
	}
	return m, nil
}

func InitConnection(exchange string) *RabbitMQConnection {
	rabbitMQConn := new(RabbitMQConnection)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	queues := make(map[string]Queue)

	rabbitMQConn.Set(conn, ch, queues, exchange)

	return rabbitMQConn
}
