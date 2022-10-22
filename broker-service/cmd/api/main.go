package main

import (
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	amqp "github.com/rabbitmq/amqp091-go"
)

const PORT = "0.0.0.0:80"

type Config struct {
	rabbitConn *amqp.Connection
}

func main() {
	var err error
	app := Config{}

	//Connect to rabbitmq
	app.rabbitConn, err = connect()
	if err != nil {
		log.Panic("Cannot connect to rabbitmq", err)
	}
	defer app.rabbitConn.Close()

	log.Println("Starting broker service on port", PORT)

	server := &http.Server{
		Addr:    PORT,
		Handler: app.routes(),
	}

	setupValidationBindings()

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("There was an error starting the server", err)
	}
}

func setupValidationBindings() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("action", validAction)
	}
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
