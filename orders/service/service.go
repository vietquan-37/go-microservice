package service

import (
	"context"
	common "github.com/vietquan-37/go-microservice/commons"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/orders/interfaces"
	"log"
)

type service struct {
	store interfaces.OrderStore
}

func NewService(store interfaces.OrderStore) *service {
	return &service{store}
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
func (s *service) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) ([]*pb.Items, error) {
	if len(p.Items) == 0 {
		return nil, common.ErrNoItems
	}
	mergedItems := mergeItems(p.Items)

	//validate with stock service
	// temporary to test the payment
	var items []*pb.Items
	for _, item := range mergedItems {
		items = append(items, &pb.Items{
			PriceID:  "price_1Qa6aTDWJlYhjZLPkH1W2KJy",
			ID:       item.ID,
			Quantity: item.Quantity,
		})
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
