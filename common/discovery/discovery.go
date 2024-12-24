package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	Register(ctx context.Context, instanceID, serviceName, hostPort string) error
	DeRegister(ctx context.Context, instanceID, serverName string) error
	Discover(ctx context.Context, serviceName string) ([]string, error)
	HealthCheck(ctx context.Context, instanceID, serviceName string) error
}

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s_%d", serviceName,
		rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
