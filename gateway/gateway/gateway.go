package gateway

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type OrdersGateway interface {
	CreateOrder(ctx context.Context, request *pb.CreateOrderRequest) (*pb.Order, error)
	GetOrder(ctx context.Context, request *pb.GetOrderRequest) (*pb.Order, error)
}
