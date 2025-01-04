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
	items, err := h.service.ValidateOrder(ctx, p)
	if err != nil {
		return nil, err
	}

	o, err := h.service.CreateOrder(ctx, p, items)
	if err != nil {
		return nil, err
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
func (h *grpcHandler) GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error) {
	o, err := h.service.GetOrder(ctx, p)
	if err != nil {
		return nil, err
	}

	return o, nil
}
