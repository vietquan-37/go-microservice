package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"go.opentelemetry.io/otel"
	"time"

	"io"
	"log"
	"net/http"
	"os"
)

type PaymentHttpHandler struct {
	channel *amqp.Channel
}

func NewPaymentHttpHandler(channel *amqp.Channel) *PaymentHttpHandler {
	return &PaymentHttpHandler{channel}
}
func (h *PaymentHttpHandler) registerRoutes(router *http.ServeMux) {
	router.HandleFunc("/webhook", h.handleCheckoutWebhook)
}
func (h *PaymentHttpHandler) handleCheckoutWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintf(os.Stdout, "Checkout Webhook: %s\n", string(payload))

	signatureHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointStripeSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling checkout session data: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		if session.PaymentStatus == "paid" {
			log.Printf("Payment successfull for %s", session.ID)
			orderID := session.Metadata["orderID"]
			customerID := session.Metadata["customerID"]
			//publish message
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			o := &pb.Order{
				ID:          orderID,
				CustomerID:  customerID,
				Status:      "paid",
				PaymentLink: "",
			}
			marshalledOrder, err := json.Marshal(o)
			if err != nil {
				log.Fatal(err.Error())
			}
			tr := otel.Tracer("amqp")
			amqpContext, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - Publishing %s", broker.OrderPaidEvent))
			defer messageSpan.End()
			header := broker.InjectAmqpHeader(amqpContext)

			//publish to message to exchange fan-out, because i want broadcast this
			h.channel.PublishWithContext(amqpContext, broker.OrderPaidEvent, "", false, false, amqp.Publishing{
				ContentType:  "application/json",
				Body:         marshalledOrder,
				DeliveryMode: amqp.Persistent,
				Headers:      header,
			})
			log.Print("message publish order.paid")
		}
	}

	w.WriteHeader(http.StatusOK)
}
