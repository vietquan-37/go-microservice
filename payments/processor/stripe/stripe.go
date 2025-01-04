package stripe

import (
	"fmt"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"

	common "github.com/vietquan-37/go-microservice/commons"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"log"
)

var (
	gatewayHttpAddr = common.EnvString("GATEWAY_HTTP_ADDR", "http://localhost:8080")
)

type Stripe struct {
}

func NewProcessor() *Stripe {
	return &Stripe{}
}
func (s *Stripe) CreatePaymentLink(p *pb.Order) (string, error) {
	log.Printf("Creating payment link for order: %v", p)
	// from order payload
	gatewaySuccessUrl := fmt.Sprintf("%s/success.html?customerID=%s&orderID=%s", gatewayHttpAddr, p.CustomerID, p.ID)
	gatewayCancelUrl := fmt.Sprintf("%s/cancel.html", gatewayHttpAddr)
	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range p.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}
	params := &stripe.CheckoutSessionParams{
		LineItems:  items,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(gatewaySuccessUrl),
		CancelURL:  stripe.String(gatewayCancelUrl),
	}
	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	return result.URL, nil

}
