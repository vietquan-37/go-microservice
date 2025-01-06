package service

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/payments/gateway"
	"github.com/vietquan-37/go-microservice/payments/processor"
)

type service struct {
	processor    processor.PaymentsProcessor
	orderGateway gateway.OrdersGateway
}

func NewService(paymentsProcessor processor.PaymentsProcessor, orderGateway gateway.OrdersGateway) *service {
	return &service{
		processor:    paymentsProcessor,
		orderGateway: orderGateway,
	}
}
func (s *service) CreatePaymentLink(ctx context.Context, p *pb.Order) (string, error) {
	//connect payment processor
	link, err := s.processor.CreatePaymentLink(p)
	if err != nil {
		return "", err
	}
	//update order with the link
	err = s.orderGateway.UpdateOrderAfterPaymentLink(ctx, p.ID, link)
	if err != nil {
		return "", err
	}
	return link, nil

}
