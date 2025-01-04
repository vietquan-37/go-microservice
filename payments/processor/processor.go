package processor

import pb "github.com/vietquan-37/go-microservice/commons/api"

type PaymentsProcessor interface {
	CreatePaymentLink(order *pb.Order) (string, error)
}
