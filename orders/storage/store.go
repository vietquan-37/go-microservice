package storage

import (
	"context"
	"errors"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type store struct {
	//database here
}

var orders = make([]*pb.Order, 0)

func NewStore() *store {
	return &store{}
}
func (s *store) Create(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Items) (*pb.Order, error) {
	o := &pb.Order{
		ID:         "42",
		CustomerID: p.CustomerID,
		Status:     "PENDING",
		Items:      items,
	}
	orders = append(orders, o)
	return o, nil
}
func (s *store) Get(ctx context.Context, request *pb.GetOrderRequest) (*pb.Order, error) {
	for _, order := range orders {
		if order.CustomerID == request.CustomerID && order.ID == request.OrderID {
			return order, nil
		}
	}
	return nil, errors.New("order not found")
}
