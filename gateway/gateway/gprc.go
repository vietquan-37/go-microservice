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

func (g *gateway) CreateOrder(ctx context.Context, request *pb.CreateOrderRequest) (*pb.Order, error) {
	conn, err := discovery.ServiceConnection(ctx, "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	c := pb.NewOrderServiceClient(conn)
	return c.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: request.CustomerID,
		Items:      request.Items,
	})
}
func (g *gateway) GetOrder(ctx context.Context, request *pb.GetOrderRequest) (*pb.Order, error) {
	conn, err := discovery.ServiceConnection(ctx, "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	c := pb.NewOrderServiceClient(conn)
	return c.GetOrder(ctx, &pb.GetOrderRequest{
		CustomerID: request.CustomerID,
		OrderID:    request.OrderID,
	})
}
