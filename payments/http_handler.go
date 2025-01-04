package main

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"

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
			//publish message
		}
	}

	w.WriteHeader(http.StatusOK)
}
