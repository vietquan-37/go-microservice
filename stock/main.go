package main

import (
	"context"
	common "github.com/vietquan-37/go-microservice/commons"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/commons/discovery"
	"github.com/vietquan-37/go-microservice/commons/discovery/consul"
	"github.com/vietquan-37/go-microservice/stock/consumer"
	"github.com/vietquan-37/go-microservice/stock/handler"
	"github.com/vietquan-37/go-microservice/stock/service"
	"github.com/vietquan-37/go-microservice/stock/storage"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"time"
)

var (
	serviceName = "stock"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2002")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	amqpUser    = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass    = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost    = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort    = common.EnvString("RABBITMQ_PORT", "5672")
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
		log.Fatal("failed to listen")
	}
	defer l.Close()
	db := DbConn("postgresql://postgres:12345@localhost:5437/stock_db?sslmode=disable")
	store := storage.NewStore(db)
	svc := service.NewService(store)
	telemetrySvc := NewTelemetryMiddleware(svc)

	handler.NewStockGrpcHandler(grpcServer, telemetrySvc)

	consumer := consumer.NewConsumer(telemetrySvc)
	go consumer.Listen(ch)

	log.Printf("Starting gRPC server: %s", grpcAddr)

	if err := grpcServer.Serve(l); err != nil {
		log.Fatal("failed to serve")
	}
}
func DbConn(DbSource string) *gorm.DB {
	db, err := gorm.Open(
		postgres.Open(DbSource), &gorm.Config{TranslateError: true},
	)
	err = db.AutoMigrate(storage.Stock{})
	if err != nil {
		log.Fatal("fail to migrate model:")
	}
	if err != nil {
		log.Fatal("fail to open database connection:")
		return nil
	}
	return db
}
