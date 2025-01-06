package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/payments/interfaces"
	"go.opentelemetry.io/otel"
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
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	var forever chan struct{}
	go func() {
		for d := range msgs {

			log.Printf("Received a message: %s", d.Body)
			ctx := broker.ExtractAmqpHeader(context.Background(), d.Headers)
			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - Consumer - %s", q.Name))

			o := &pb.Order{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				d.Nack(false, false)
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}
			paymentLink, err := c.service.CreatePaymentLink(context.Background(), o)
			if err == nil {
				log.Printf("failed to create payment link: %v", err)
				if err := broker.HandleRetry(ch, &d); err != nil {
					log.Printf("failed to handle retry: %v", err)
				}
				d.Nack(false, false)
				continue

			}
			messageSpan.AddEvent("payment link created")
			messageSpan.End()
			log.Printf("Payment link created: %v", paymentLink)
			d.Ack(false)

		}

	}()
	<-forever
}
