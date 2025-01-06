package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
	"github.com/vietquan-37/go-microservice/payments/gateway"
	"net/http"

	common "github.com/vietquan-37/go-microservice/commons"
	"github.com/vietquan-37/go-microservice/commons/broker"
	"github.com/vietquan-37/go-microservice/commons/discovery"
	"github.com/vietquan-37/go-microservice/commons/discovery/consul"
	"github.com/vietquan-37/go-microservice/payments/consumer"
	stripeProccessor "github.com/vietquan-37/go-microservice/payments/processor/stripe"
	"github.com/vietquan-37/go-microservice/payments/service"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	amqpUser             = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass             = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost             = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort             = common.EnvString("RABBITMQ_PORT", "5672")
	serviceName          = "payments"
	grpcAddr             = common.EnvString("GRPC_ADDR", "localhost:2001")
	consulAddr           = common.EnvString("CONSUL_ADDR", "localhost:8500")
	stripeApiKey         = common.EnvString("STRIPE_API_KEY", "")
	httpAddr             = common.EnvString("HTTP_ADDR", "localhost:8081")
	endpointStripeSecret = common.EnvString("STRIPE_SECRET", "")
	JaegerAddr           = common.EnvString("JAEGER_ADDR", "localhost:4318")
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
	//stripe set up
	stripe.Key = stripeApiKey
	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()
	stripeP := stripeProccessor.NewProcessor()
	g := gateway.NewGrpcGateway(registry)
	svc := service.NewService(stripeP, g)
	telemetrySvc := NewTelemetryMiddleware(svc)
	amqpConsumer := consumer.NewConsumer(telemetrySvc)
	go amqpConsumer.Listen(ch)
	//http server webhhok
	mux := http.NewServeMux()
	httpserver := NewPaymentHttpHandler(ch)
	httpserver.registerRoutes(mux)
	go func() {
		log.Printf("Starting http server on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, mux); err != nil {
			log.Fatal("failed to start http server ", err)
		}
	}()
	grpcServer := grpc.NewServer()
	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("fail to listen: %v", err)
	}
	defer l.Close()
	log.Println("Grpc server started at", grpcAddr)
	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err)
	}

}
