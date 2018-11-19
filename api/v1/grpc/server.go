package grpc

import (
	"context"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	gRPCHealth "google.golang.org/grpc/health/grpc_health_v1"
)

// pushService holds all the the information and functionality regarding the push implementation
type pushService struct{}

// ActivateSubscription activates a subscription so the service can start handling the push functionality
func (ps *pushService) ActivateSubscription(ctx context.Context, sub *amsPb.ActivateSubscriptionRequest) (*amsPb.Status, error) {
	return &amsPb.Status{}, nil
}

// DeactivateSubscription deactivates a subscription so the service can stop handling the push functionality for it
func (ps *pushService) DeactivateSubscription(ctx context.Context, r *amsPb.DeactivateSubscriptionRequest) (*amsPb.Status, error) {
	return &amsPb.Status{}, nil
}

// NewGRPCServer configures and returns a new *grpc.Server
func NewGRPCServer() *grpc.Server {

	srv := grpc.NewServer()

	healthService := health.NewServer()
	healthService.SetServingStatus("api.v1.grpc.PushService", gRPCHealth.HealthCheckResponse_SERVING)

	amsPb.RegisterPushServiceServer(srv, new(pushService))

	gRPCHealth.RegisterHealthServer(srv, healthService)

	return srv
}
