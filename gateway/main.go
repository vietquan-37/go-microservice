package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	common "github.com/vietquan-37/go-microservice/commons"
	"github.com/vietquan-37/go-microservice/commons/discovery"
	"github.com/vietquan-37/go-microservice/commons/discovery/consul"
	"github.com/vietquan-37/go-microservice/gateway/gateway"
	"time"

	"log"
	"net/http"
)

var (
	httpAddr    = common.EnvString("HTTP_ADDR", ":8080")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	serviceName = "gateway"
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
	if err := registry.Register(ctx, instanceID, serviceName, httpAddr); err != nil {
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
	mux := http.NewServeMux()
	orderGateway := gateway.NewGrpcGateway(registry)
	handler := NewHandler(orderGateway)
	handler.registerRoutes(mux)
	log.Printf("Starting server at %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("Failed to start http server")
	}
}
