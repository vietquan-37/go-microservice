package service

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/payments/processor"
)

type service struct {
	processor processor.PaymentsProcessor
}

func NewService(paymentsProcessor processor.PaymentsProcessor) *service {
	return &service{
		processor: paymentsProcessor,
	}
}
func (s *service) CreatePaymentLink(ctx context.Context, p *pb.Order) (string, error) {
	//connect payment processor
	link, err := s.processor.CreatePaymentLink(p)
	if err != nil {
		return "", err
	}
	return link, nil

}
