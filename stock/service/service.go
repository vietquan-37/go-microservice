package service

import (
	"context"
	"errors"

	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/stock/interfaces"
)

type service struct {
	store interfaces.StockStore
}

func NewService(store interfaces.StockStore) *service {
	return &service{store}
}

// coi lai stock
func (s *service) CheckItemInStock(ctx context.Context, request *pb.CheckStockRequest) (bool, []*pb.Items, error) {
	i := make(map[int32]int32, len(request.Items))
	for _, item := range request.Items {
		i[item.ID] = item.Quantity
	}
	ids := make([]int32, 0, len(i))
	for id := range i {
		ids = append(ids, id)
	}
	stocks, err := s.store.GetItems(ctx, ids)
	if err != nil {
		return false, nil, err
	}
	for _, stock := range stocks {
		if ItemQuantity, ok := i[int32(stock.ID)]; ok {
			if ItemQuantity > stock.Quantity {
				return false, covertItems(stocks), nil
			}

		}
	}
	return true, covertItems(stocks), nil
}

func (s *service) DecreaseStock(ctx context.Context, items []*pb.Items) error {
	i := make(map[int32]int32, len(items))
	for _, item := range items {
		i[item.ID] = item.Quantity
	}
	ids := make([]int32, 0, len(i))
	for id := range i {
		ids = append(ids, id)
	}
	stocks, err := s.store.GetItems(ctx, ids)
	if err != nil {
		return err
	}
	for _, stock := range stocks {
		if ItemQuantity, ok := i[int32(stock.ID)]; ok {
			if ItemQuantity > stock.Quantity {
				return errors.New("Quantity is not enough ")
			}
			stock.Quantity -= ItemQuantity
		}
	}
	if err := s.store.UpdateStock(ctx, stocks); err != nil {
		return err
	}
	return nil
}
