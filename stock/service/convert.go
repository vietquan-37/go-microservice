package service

import (
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/stock/storage"
)

func covertItems(stocks []*storage.Stock) (items []*pb.Items) {
	for _, stock := range stocks {
		item := convert(stock)
		items = append(items, item)
	}
	return items
}
func convert(stock *storage.Stock) *pb.Items {
	return &pb.Items{

		ID:       int32(stock.ID),
		Quantity: stock.Quantity,
		PriceID:  stock.PriceId,
	}
}
