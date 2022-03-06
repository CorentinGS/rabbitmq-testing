package main

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Log struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Event     string    `json:"event"`
	CreatedAt time.Time `json:"createdat"`
	UserID    uint      `json:"userid"`
	DeckID    uint      `json:"deckid"`
	CardID    uint      `json:"cardid"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
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

	body, _ := json.Marshal(Log{
		Type:      "Logging",
		Message:   "Toto logged in",
		Level:     "error",
		Event:     "login",
		CreatedAt: time.Now(),
		UserID:    132,
		DeckID:    0,
	})
	err = ch.Publish(
		"logs",             // exchange
		"error.auth.login", // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}
