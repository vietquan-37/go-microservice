package gateway

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type StockGateway interface {
	CheckIfItemIsInStock(ctx context.Context, items []*pb.Items) ([]*pb.Items, error)
}
