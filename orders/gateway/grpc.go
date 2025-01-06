package gateway

import (
	"context"
	pb "github.com/vietquan-37/go-microservice/commons/api"
	"github.com/vietquan-37/go-microservice/commons/discovery"
	"log"
)

type Gateway struct {
	registry discovery.Registry
}

func NewGateway(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) CheckIfItemIsInStock(ctx context.Context, items []*pb.Items) (bool, []*pb.Items, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "stock", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewStockServiceClient(conn)

	res, err := c.CheckStock(ctx, &pb.CheckStockRequest{
		Items: items,
	})

	return res.InStock, res.Items, err
}
