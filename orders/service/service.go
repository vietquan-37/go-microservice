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
func (s *service) CreateOrder(ctx context.Context) error {
	return nil
}
func (s *service) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) error {
	if len(p.Items) == 0 {
		return common.ErrNoItems
	}
	mergedItems := mergeItems(p.Items)
	log.Print(mergedItems)
	//validate with stock service
	return nil
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
