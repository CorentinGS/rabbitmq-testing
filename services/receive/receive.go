package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

var mg MongoInstance

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

	connection := InitConnection("logs")

	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			failOnError(err, "Failed to close connection")
		}
	}(connection.conn)

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			failOnError(err, "Failed to close channel")
		}
	}(connection.channel)

	queues := []Queue{{
		keys:       []string{"error.#"},
		Name:       "error",
		Collection: "error",
	}, {
		keys:       []string{"info.#"},
		Name:       "info",
		Collection: "info",
	}}

	err = connection.SetQueues(queues)
	if err != nil {
		return
	}

	forever := make(chan bool)

	deliveries, err := connection.Consume()
	if err != nil {
		panic(err)
	}

	for q, d := range deliveries {
		go func(q string, delivery <-chan amqp.Delivery) {
			for d := range delivery {
				log.Printf("Received a message: %s from %s", d.Body, q)
				logObject := new(Log)
				err := json.Unmarshal(d.Body, &logObject)
				collection := mg.Db.Collection(connection.queues[q].Collection)
				_, err = collection.InsertOne(context.TODO(), logObject)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}(q, d)
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	fmt.Println("Test")
}
