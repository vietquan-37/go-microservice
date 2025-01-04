package main

import (
	"context"
	"github.com/stretchr/testify/require"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/payments/processor/inmemory"
	"github.com/vietquan-37/go-microservice/payments/service"
	"testing"
)

func TestPaymentService(t *testing.T) {
	//so this is the advantages of interface in go
	processor := inmemory.NewInMemoryPayment()
	svc := service.NewService(processor)
	t.Run("test payment service", func(t *testing.T) {
		link, err := svc.CreatePaymentLink(context.Background(), &pb.Order{})

		require.NoError(t, err)
		require.NotEmpty(t, link)
	})
}
