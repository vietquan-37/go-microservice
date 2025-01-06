package main

import (
	"context"
	"fmt"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/stock/interfaces"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryMiddleware struct {
	service interfaces.StockService
}

func NewTelemetryMiddleware(service interfaces.StockService) *TelemetryMiddleware {
	return &TelemetryMiddleware{
		service: service,
	}

}
func (s *TelemetryMiddleware) CheckItemInStock(ctx context.Context, request *pb.CheckStockRequest) (bool, []*pb.Items, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("CheckItemInStock: %v", request))
	return s.service.CheckItemInStock(ctx, request)
}
func (s *TelemetryMiddleware) DecreaseStock(ctx context.Context, items []*pb.Items) error {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("DecreaseStock: %v", items))
	return s.service.DecreaseStock(ctx, items)
}
