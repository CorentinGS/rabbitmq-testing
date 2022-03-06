package main

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

var mg MongoInstance

type Topic struct {
	Key        string
	Collection string
}

func main() {
	err := Connect()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		fmt.Println("Disconnect")
		err := mg.Client.Disconnect(context.TODO())
		if err != nil {
			return
		}
	}()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			failOnError(err, "Failed to close connection")
		}
	}(conn)

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			failOnError(err, "Failed to close channel")
		}
	}(ch)

	err = ch.ExchangeDeclare(
		"logs",  // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	forever := make(chan bool)

	topics := []Topic{{Key: "error.#", Collection: "error"}, {Key: "info.#", Collection: "info"}}
	for _, topic := range topics {
		go func(topic Topic) {
			ErrorChan(ch, topic)
		}(topic)
	}

	<-forever

	fmt.Println("Test")
}
