package service

import (
	"context"
	common "github.com/vietquan-37/go-microservice/commons"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/orders/gateway"
	"github.com/vietquan-37/go-microservice/orders/interfaces"
	"log"
)

type service struct {
	store   interfaces.OrderStore
	gateway gateway.StockGateway
}

func NewService(store interfaces.OrderStore, stockGateway gateway.StockGateway) *service {
	return &service{store, stockGateway}
}
func (s *service) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Items) (*pb.Order, error) {
	p.Items = items
	o, err := s.store.Create(ctx, p, items)

	if err != nil {
		return nil, err
	}
	return o, nil

}
func (s *service) GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error) {
	o, err := s.store.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (s *service) UpdateOrder(ctx context.Context, p *pb.Order) (*pb.Order, error) {
	err := s.store.Update(ctx, p.ID, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (s *service) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) ([]*pb.Items, error) {
	if len(p.Items) == 0 {
		return nil, common.ErrNoItems
	}
	mergedItems := mergeItems(p.Items)

	inStock, items, err := s.gateway.CheckIfItemIsInStock(ctx, mergedItems)
	if err != nil {
		return nil, err
	}
	if !inStock {
		return items, common.ErrNoItems
	}
	log.Print(items)
	return items, nil
}
func mergeItems(items []*pb.Items) []*pb.Items {
	merged := make([]*pb.Items, 0)
	for _, item := range items {
		found := false
		for _, finalItem := range merged {
			if finalItem.ID == item.ID {
				finalItem.Quantity += item.Quantity
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, item)
		}
	}
	return merged
}
