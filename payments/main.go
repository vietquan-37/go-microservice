package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stripe/stripe-go/v81"
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
	endpointStripeSecret = common.EnvString("STRIPE_SECRET", "whsec_4739a0edd9d7d631b2d6bb6f6a2bdb4e32f510da622622efc1527461a5269eac")
)

func main() {
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
	svc := service.NewService(stripeP)
	amqpConsumer := consumer.NewConsumer(svc)
	go amqpConsumer.Listen(ch)
	//http server
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
