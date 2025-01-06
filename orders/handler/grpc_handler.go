package handler

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/orders/interfaces"
	"go.opentelemetry.io/otel"
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
func (h *grpcHandler) UpdateOrder(ctx context.Context, p *pb.Order) (*pb.Order, error) {
	return h.service.UpdateOrder(ctx, p)
}
func (h *grpcHandler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {
	tr := otel.Tracer("amqp")
	q, err := h.channel.QueueDeclare(broker.OrderCreateEvent, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	amqpContext, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - Publishing %s", q.Name))
	defer messageSpan.End()
	log.Println("New order received !")
	items, err := h.service.ValidateOrder(amqpContext, p)
	if err != nil {
		return nil, err
	}

	o, err := h.service.CreateOrder(amqpContext, p, items)
	if err != nil {
		return nil, err
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	//inject
	header := broker.InjectAmqpHeader(amqpContext)
	//Producer publish message directly to queue name order.created , proper explain that wouble q.name is routing key
	h.channel.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         marshalledOrder,
		DeliveryMode: amqp.Persistent,
		Headers:      header,
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
