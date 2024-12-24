package interfaces

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type OrderService interface {
	CreateOrder(ctx context.Context) error
	ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) error
}
type OrderStore interface {
	Create(ctx context.Context) error
}
