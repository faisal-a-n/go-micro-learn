package main

import (
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	//Connect to rabbitMQ
	rabbitConn, err := connect()
	if err != nil {
		log.Panic("Cannot start rabbitmq")
		os.Exit(1)
	}
	defer rabbitConn.Close()

	//start listening for messages
	log.Println("Listening for rabbitmq messages")
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Panic("Cannot initialize consumer", err)
		os.Exit(1)
	}

	//create consumer
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println("Consumer unable to listen", err)
	}
	//watch queue and consume event
}

func connect() (*amqp.Connection, error) {
	var counts int64
	backOff := 1 * time.Second
	connection := &amqp.Connection{}

	for {
		c, err := amqp.Dial("amqp://admin:password@rabbitmq")
		if err != nil {
			log.Println("rabbitmq not ready yet")
			counts++
		} else {
			connection = c
			break
		}
		if counts > 5 {
			log.Println("Error connecting to rabbitmq", err)
			return nil, err
		}

		time.Sleep(backOff)
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
	}

	log.Println("Connected to rabbitmq")
	return connection, nil
}
