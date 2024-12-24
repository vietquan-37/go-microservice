package handler

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/orders/interfaces"
	"google.golang.org/grpc"
	"log"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer
	service interfaces.OrderService
	channel *amqp.Channel
}

func NewGrpcHandler(grpcServer *grpc.Server, service interfaces.OrderService, channel *amqp.Channel) {
	handler := &grpcHandler{
		service: service,
		channel: channel,
	}
	pb.RegisterOrderServiceServer(grpcServer, handler)
}

func (h *grpcHandler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {
	log.Println("New order received !")
	o := &pb.Order{
		ID: "42",
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	q, err := h.channel.QueueDeclare(broker.OrderCreateEvent, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	//Producer
	h.channel.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         marshalledOrder,
		DeliveryMode: amqp.Persistent,
	})
	return o, nil
}
