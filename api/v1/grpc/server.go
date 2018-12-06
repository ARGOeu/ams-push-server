package grpc

import (
	"context"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"

	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	gRPCHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// PushService holds all the the information and functionality regarding the push implementation
type PushService struct {
	Subscriptions map[string]*amsPb.Subscription
}

// NewPushService returns a pointer to a PushService and initialises its fields
func NewPushService() *PushService {
	ps := new(PushService)
	ps.Subscriptions = make(map[string]*amsPb.Subscription)
	return ps
}

// ActivateSubscription activates a subscription so the service can start handling the push functionality
func (ps *PushService) ActivateSubscription(ctx context.Context, r *amsPb.ActivateSubscriptionRequest) (*amsPb.ActivateSubscriptionResponse, error) {

	if r.Subscription == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Empty subscription")
	}

	if ps.IsSubActive(r.Subscription.FullName) {
		return nil, status.Errorf(codes.AlreadyExists, "Subscription %v is already activated", r.Subscription.FullName)
	}

	ps.Subscriptions[r.Subscription.FullName] = r.Subscription

	return &amsPb.ActivateSubscriptionResponse{
		Message: fmt.Sprintf("Subscription %v activated", r.Subscription.FullName),
	}, nil
}

// DeactivateSubscription deactivates a subscription so the service can stop handling the push functionality for it
func (ps *PushService) DeactivateSubscription(ctx context.Context, r *amsPb.DeactivateSubscriptionRequest) (*amsPb.DeactivateSubscriptionResponse, error) {
	return nil, nil
}

// GetSubscription finds and returns a subscription
func (ps *PushService) GetSubscription(ctx context.Context, r *amsPb.GetSubscriptionRequest) (*amsPb.GetSubscriptionResponse, error) {

	if !ps.IsSubActive(r.FullName) {
		return nil, status.Errorf(codes.NotFound, "Subscription %v is not active", r.FullName)
	}

	return &amsPb.GetSubscriptionResponse{
		Subscription: ps.Subscriptions[r.FullName],
	}, nil
}

// IsSubActive checks by subscription name, whether or not a subscription is already active
func (ps *PushService) IsSubActive(name string) bool {

	_, found := ps.Subscriptions[name]

	return found
}

// NewGRPCServer configures and returns a new *grpc.Server
func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {

	srv := grpc.NewServer(opt...)

	healthService := health.NewServer()
	healthService.SetServingStatus("api.v1.grpc.PushService", gRPCHealth.HealthCheckResponse_SERVING)

	amsPb.RegisterPushServiceServer(srv, NewPushService())

	gRPCHealth.RegisterHealthServer(srv, healthService)

	return srv
}
