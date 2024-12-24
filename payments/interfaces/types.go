package interfaces

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
)

type PaymentService interface {
	CreatePaymentLink(context.Context, *pb.Order) (string, error)
}
