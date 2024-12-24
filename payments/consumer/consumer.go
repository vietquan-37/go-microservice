package consumer

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/payments/interfaces"
	"log"
)

type consumer struct {
	service interfaces.PaymentService
}

func NewConsumer(service interfaces.PaymentService) *consumer {
	return &consumer{service}
}
func (c *consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.OrderCreateEvent, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			o := &pb.Order{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}
			paymentLink, err := c.service.CreatePaymentLink(context.Background(), o)
			if err != nil {
				log.Printf("failed to create payment link: %v", err)
				continue
			}
			log.Printf("Payment link created: %v", paymentLink)

		}

	}()
	<-forever
}
