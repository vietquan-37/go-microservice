package main

import (
	"context"
	"fmt"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/payments/interfaces"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryMiddleware struct {
	service interfaces.PaymentService
}

func NewTelemetryMiddleware(service interfaces.PaymentService) interfaces.PaymentService {
	return &TelemetryMiddleware{
		service: service,
	}

}
func (s *TelemetryMiddleware) CreatePaymentLink(ctx context.Context, p *pb.Order) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("CreatePaymentLink: %v", p))
	return s.service.CreatePaymentLink(ctx, p)
}
