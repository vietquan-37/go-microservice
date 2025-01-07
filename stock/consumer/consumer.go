package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/stock/interfaces"
	"go.opentelemetry.io/otel"
	"log"
)

type consumer struct {
	service interfaces.StockService
}

func NewConsumer(service interfaces.StockService) *consumer {
	return &consumer{service}
}
func (c *consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = ch.QueueBind(q.Name, "", broker.OrderPaidEvent, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	msg, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	var forever chan struct{}
	go func() {
		for d := range msg {
			log.Printf("Received a message: %s", d.Body)
			ctx := broker.ExtractAmqpHeader(context.Background(), d.Headers)
			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - consume - %s", q.Name))

			o := &pb.Order{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				d.Nack(false, false)
				log.Printf("json.Unmarshal err: %v", err)
				continue
			}
			err := c.service.DecreaseStock(context.Background(), o.Items)
			if err != nil {
				d.Nack(false, false)
				log.Printf("DecreaseStock err: %v", err)
				if err := broker.HandleRetry(ch, &d); err != nil {
					log.Printf("Error handling retry: %v", err)
				}
				continue
			}
			messageSpan.AddEvent("stock.updated")
			messageSpan.End()

			log.Println("stock has been updated from AMQP")
			d.Ack(false)
		}
	}()
	<-forever
}
