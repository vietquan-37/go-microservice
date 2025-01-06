package broker

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"log"
	"time"
)

const (
	maxRetryCount = 3
	DLQ           = "dlq_main"
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
	if err := createDLQAndDLX(ch); err != nil {
		log.Fatalf("Failed to create DLQ: %s", err)
	}
	return ch, conn.Close
}
func HandleRetry(ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers == nil {
		d.Headers = amqp.Table{}
	}
	retryCount, ok := d.Headers["x-retry-count"].(int64)
	if !ok {
		retryCount = 0
	}
	retryCount++
	d.Headers["x-retry-count"] = retryCount
	log.Printf("Retry message %s ,retry count: %d", d.Body, retryCount)
	if retryCount >= maxRetryCount {
		log.Printf("Moving to DLQ %s", DLQ)
		return ch.PublishWithContext(context.Background(), "", DLQ, false, false, amqp.Publishing{
			ContentType:  "application/json",
			Headers:      d.Headers,
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})

	}
	time.Sleep(time.Duration(retryCount) * time.Second)
	return ch.PublishWithContext(context.Background(),
		d.Exchange,
		d.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Headers:      d.Headers,
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		},
	)

}
func createDLQAndDLX(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare(
		"main_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	// Declare DLX
	dlx := "dlx_main"
	err = ch.ExchangeDeclare(
		dlx,      // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	// Bind main queue to DLX
	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		dlx,    // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Declare DLQ
	_, err = ch.QueueDeclare(
		DLQ,   // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return err
}

type HeaderCarrier map[string]interface{}

func (h HeaderCarrier) Get(key string) string {
	value, ok := h[key]
	if !ok {
		return ""
	}
	return value.(string)
}
func (h HeaderCarrier) Set(key string, value string) {
	h[key] = value

}
func (h HeaderCarrier) Keys() []string {
	keys := make([]string, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}
func InjectAmqpHeader(ctx context.Context) map[string]interface{} {
	carrier := make(HeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}
func ExtractAmqpHeader(ctx context.Context, header map[string]interface{}) context.Context {

	return otel.GetTextMapPropagator().Extract(ctx, HeaderCarrier(header))
}
