package inmemory

import pb "github.com/vietquan-37/go-microservice/commons/api"

type inMemoryPayment struct{}

func NewInMemoryPayment() *inMemoryPayment {
	return &inMemoryPayment{}
}
func (i *inMemoryPayment) CreatePaymentLink(p *pb.Order) (string, error) {
	return "dummy-link", nil
}
