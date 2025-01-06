package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/orders/interfaces"
	"go.opentelemetry.io/otel"
	"log"
)

type consumer struct {
	service interfaces.OrderService
}

func NewConsumer(service interfaces.OrderService) *consumer {
	return &consumer{service}
}
func (c *consumer) Listen(ch *amqp.Channel) {
	//the queue name auto create by rabbitmq
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	//this mean the connection between the queue above with the order.paid exchange
	//so message come from that exchange will belong to queue above
	err = ch.QueueBind(q.Name, "", broker.OrderPaidEvent, false, nil)
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
			log.Printf("Received message: %s", d.Body)
			o := &pb.Order{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				// if requeue true infinite loop
				d.Nack(false, false)
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}

			//Extract header
			ctx := broker.ExtractAmqpHeader(context.Background(), d.Headers)
			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP -Consumer - %s", q.Name))
			//note here
			_, err := c.service.UpdateOrder(ctx, o)
			if err != nil {

				log.Printf("failed to update order: %v", err)
				if err := broker.HandleRetry(ch, &d); err != nil {

					log.Printf("failed to handle retry: %v", err)
				}
				log.Printf("exchange and routing key: %v %v", d.Exchange, d.RoutingKey)
				d.Nack(false, false)
				continue
			}
			messageSpan.AddEvent("order updated")
			messageSpan.End()

			log.Print("order has been updated from AMQP")
			d.Ack(false)

		}
	}()

	<-forever
}
