package broker

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func Connect(user, pass, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, pass, host, port)
	conn, err := amqp.Dial(address)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	if err := ch.ExchangeDeclare(OrderCreateEvent, "direct", true, false, false, false, nil); err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
	}
	if err := ch.ExchangeDeclare(OrderPaidEvent, "fanout", true, false, false, false, nil); err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
	}
	return ch, conn.Close
}
