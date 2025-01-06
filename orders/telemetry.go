package main

import (
	"context"
	"fmt"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/orders/interfaces"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryMiddleware struct {
	next interfaces.OrderService
}

func NewTelemetryMiddleware(next interfaces.OrderService) interfaces.OrderService {
	return &TelemetryMiddleware{
		next: next,
	}
}

func (s *TelemetryMiddleware) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Items) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("CreateOrder: %v", p))

	return s.next.CreateOrder(ctx, p, items)
}
func (s *TelemetryMiddleware) GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("GetOrder: %v", p))
	return s.next.GetOrder(ctx, p)
}
func (s *TelemetryMiddleware) UpdateOrder(ctx context.Context, p *pb.Order) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("UpdateOrder: %v", p))
	return s.next.UpdateOrder(ctx, p)
}
func (s *TelemetryMiddleware) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) ([]*pb.Items, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("Validate: %v", p))
	return s.next.ValidateOrder(ctx, p)
}
