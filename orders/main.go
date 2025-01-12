package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	common "github.com/vietquan-37/go-microservice/commons"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/commons/discovery"
	"github.com/vietquan-37/go-microservice/commons/discovery/consul"
	"github.com/vietquan-37/go-microservice/orders/consumer"
	"github.com/vietquan-37/go-microservice/orders/gateway"
	"github.com/vietquan-37/go-microservice/orders/handler"
	"github.com/vietquan-37/go-microservice/orders/service"
	"github.com/vietquan-37/go-microservice/orders/storage"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	amqpUser    = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass    = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost    = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort    = common.EnvString("RABBITMQ_PORT", "5672")
	serviceName = "orders"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2000")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	JaegerAddr  = common.EnvString("JAEGER_ADDR", "localhost:4318")
)

func main() {
	err := common.SetGlobalTracer(context.TODO(), serviceName, JaegerAddr)
	if err != nil {
		log.Fatalf("fail to init tracer: %v", err)
	}
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.HealthCheck(ctx, instanceID, serviceName); err != nil {
				log.Fatal("Failed to health check")
			}
			time.Sleep(time.Second * 1)
		}
	}()
	defer registry.DeRegister(ctx, instanceID, serviceName)
	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	grpcServer := grpc.NewServer()
	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("fail to listen: %v", err)
	}
	defer l.Close()
	store := storage.NewStore()
	gateways := gateway.NewGateway(registry)
	svc := service.NewService(store, gateways)
	telemetryService := NewTelemetryMiddleware(svc)
	amqpConsumer := consumer.NewConsumer(telemetryService)
	go amqpConsumer.Listen(ch)
	handler.NewGrpcHandler(grpcServer, telemetryService, ch)
	log.Println("Grpc server started at", grpcAddr)
	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err)
	}
}
