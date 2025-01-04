package interfaces

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type OrderService interface {
	CreateOrder(ctx context.Context, request *pb.CreateOrderRequest, items []*pb.Items) (*pb.Order, error)
	ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) ([]*pb.Items, error)
	GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error)
}
type OrderStore interface {
	Create(ctx context.Context, request *pb.CreateOrderRequest, items []*pb.Items) (*pb.Order, error)
	Get(ctx context.Context, request *pb.GetOrderRequest) (*pb.Order, error)
}
