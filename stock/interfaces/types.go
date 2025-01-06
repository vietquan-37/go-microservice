package interfaces

import (
	"context"
	_ "github.com/google/uuid"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/stock/storage"
)

type StockService interface {
	CheckItemInStock(ctx context.Context, request *pb.CheckStockRequest) (bool, []*pb.Items, error)
	DecreaseStock(ctx context.Context, items []*pb.Items) error
}
type StockStore interface {
	GetItems(ctx context.Context, ids []int32) ([]*storage.Stock, error)
	UpdateStock(context.Context, []*storage.Stock) error
}
