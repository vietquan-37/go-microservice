package handler

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/stock/interfaces"
	"google.golang.org/grpc"
)

type StockGrpcHandler struct {
	pb.UnimplementedStockServiceServer
	stockService interfaces.StockService
}

func NewStockGrpcHandler(grpcService *grpc.Server, stockService interfaces.StockService) {
	handler := &StockGrpcHandler{
		stockService: stockService,
	}
	pb.RegisterStockServiceServer(grpcService, handler)
}
func (s *StockGrpcHandler) CheckStock(ctx context.Context, p *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	items, err := s.stockService.CheckItemInStock(ctx, p)
	if err != nil {
		return nil, err
	}

	return &pb.CheckStockResponse{

		Items: items,
	}, nil
}
