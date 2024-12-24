package service

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type service struct {
}

func NewService() *service {
	return &service{}
}
func (s *service) CreatePaymentLink(ctx context.Context, p *pb.Order) (string, error) {
	//connect payment processor
	return "", nil
}
