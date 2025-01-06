package gateway

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/discovery"
	"log"
)

type gateway struct {
	registry discovery.Registry
}

func NewGrpcGateway(registry discovery.Registry) *gateway {
	return &gateway{
		registry: registry,
	}
}

func (g *gateway) UpdateOrderAfterPaymentLink(ctx context.Context, orderID, paymentLink string) error {
	conn, err := discovery.ServiceConnection(ctx, "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()
	orderClient := pb.NewOrderServiceClient(conn)
	_, err = orderClient.UpdateOrder(ctx, &pb.Order{
		ID:          orderID,
		PaymentLink: paymentLink,
		Status:      "waiting_payment",
	})
	return err
}
