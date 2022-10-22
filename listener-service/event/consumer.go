package event

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}
	return consumer, nil
}

func (this *Consumer) setup() error {
	channel, err := this.conn.Channel()
	if err != nil {
		return nil
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (this *Consumer) Listen(topics []string) error {
	ch, err := this.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, v := range topics {
		err := ch.QueueBind(
			q.Name,
			v,
			"logs_topic",
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}
	messages, err := ch.Consume(q.Name,
		"",
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			log.Println("Recieved log", string(d.Body))
			go handlePayload(payload)
		}
	}()

	log.Println("Started listening for events on logs_topic::", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println("Cannot log payload", err)
		}
	case "auth":
	case "mail":
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println("Cannot log payload", err)
		}
	}
}
