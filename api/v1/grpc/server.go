package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/config"
	"github.com/ARGOeu/ams-push-server/consumers"
	"github.com/ARGOeu/ams-push-server/push"
	"github.com/ARGOeu/ams-push-server/senders"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	gRPCHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"log/syslog"
	"net/http"
	"net/url"
	"time"
)

// PushService holds all the the information and functionality regarding the push implementation
type PushService struct {
	Cfg            *config.Config
	Client         *http.Client
	PushWorkers    map[string]push.Worker
	deactivateChan chan consumers.CancelableError
}

// NewPushService returns a pointer to a PushService and initialises its fields
func NewPushService(cfg *config.Config) *PushService {

	ps := new(PushService)

	ps.Cfg = cfg
	ps.PushWorkers = make(map[string]push.Worker)

	// build the client
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !ps.Cfg.VerifySSL},
	}

	client := &http.Client{
		Transport: transCfg,
		Timeout:   time.Duration(30 * time.Second),
	}

	ps.Client = client

	ps.deactivateChan = make(chan consumers.CancelableError)
	go ps.handleDeactivateChannel()

	return ps
}

// handleDeactivateChannel listens on the deactivate channel in order to stop any subscription that caused a cancelable error
func (ps *PushService) handleDeactivateChannel() {

	for {
		cancelErr, ok := <-ps.deactivateChan
		if ok {
			err := ps.deactivateSubscription(cancelErr.Resource)
			if err != nil {
				logrus.WithFields(
					log.Fields{
						"type":         "system_log",
						"subscription": cancelErr.Resource,
					},
				).Warning("Tried to deactivate malfunctioning subscription but was not active")
			}
			logrus.WithFields(
				logrus.Fields{
					"type":         "system_log",
					"subscription": cancelErr.Resource,
					"error":        cancelErr.ErrMsg,
				},
			).Info("Deactivated malfunctioning subscription")

		}
	}
}

// ActivateSubscription activates a subscription so the service can start handling the push functionality
func (ps *PushService) ActivateSubscription(ctx context.Context, r *amsPb.ActivateSubscriptionRequest) (*amsPb.ActivateSubscriptionResponse, error) {

	if r.Subscription == nil || r.Subscription.PushConfig == nil || r.Subscription.PushConfig.RetryPolicy == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Empty subscription")
	}

	if ps.IsSubActive(r.Subscription.FullName) {
		return nil, status.Errorf(codes.AlreadyExists, "Subscription %v is already activated", r.Subscription.FullName)
	}

	_, err := url.ParseRequestURI(r.Subscription.PushConfig.PushEndpoint)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid push endpoint, %v", err.Error())
	}

	// choose a consumer
	c, _ := consumers.New(consumers.AmsHttpConsumerType, r.Subscription.FullName, ps.Cfg, ps.Client)

	// choose a sender
	s, _ := senders.New(senders.HttpSenderType, r.Subscription.PushConfig.PushEndpoint, ps.Client)

	worker, err := push.New(r.Subscription, c, s, ps.deactivateChan)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid argument, %v", err.Error())
	}

	ps.PushWorkers[r.Subscription.FullName] = worker
	go worker.Start()

	return &amsPb.ActivateSubscriptionResponse{
		Message: fmt.Sprintf("Subscription %v activated", r.Subscription.FullName),
	}, nil
}

// DeactivateSubscription deactivates a subscription so the service can stop handling the push functionality for it
func (ps *PushService) DeactivateSubscription(ctx context.Context, r *amsPb.DeactivateSubscriptionRequest) (*amsPb.DeactivateSubscriptionResponse, error) {

	err := ps.deactivateSubscription(r.FullName)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &amsPb.DeactivateSubscriptionResponse{
		Message: fmt.Sprintf("Subscription %v deactivated", r.FullName),
	}, nil
}

// deactivateSubscription checks if the sub is active, then stops the respective worker and removes the sub from the map
func (ps *PushService) deactivateSubscription(sub string) error {

	if !ps.IsSubActive(sub) {
		return errors.Errorf("Subscription %v is not active", sub)
	}

	w := ps.PushWorkers[sub]
	w.Stop()

	delete(ps.PushWorkers, sub)

	return nil
}

// IsSubActive checks by subscription name, whether or not a subscription is already active
func (ps *PushService) IsSubActive(name string) bool {

	_, found := ps.PushWorkers[name]

	return found
}

// NewGRPCServer configures and returns a new *grpc.Server
func NewGRPCServer(cfg *config.Config) *grpc.Server {

	grpcLogger := log.New()
	fmter := &log.TextFormatter{FullTimestamp: true, DisableColors: true}
	grpcLogger.SetFormatter(fmter)
	hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		grpcLogger.AddHook(hook)
	}
	grpcLogger.SetLevel(cfg.GetLogLevel())

	var srvOptions []grpc.ServerOption

	logOpts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}

	logOptions := grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(grpcLogger), logOpts...),
	)

	srvOptions = append(srvOptions, logOptions)

	if cfg.TLSEnabled {
		srvOptions = append(srvOptions, grpc.Creds(credentials.NewTLS(cfg.GetTLSConfig())))
	}

	srv := grpc.NewServer(srvOptions...)

	healthService := health.NewServer()
	healthService.SetServingStatus("api.v1.grpc.PushService", gRPCHealth.HealthCheckResponse_SERVING)
	gRPCHealth.RegisterHealthServer(srv, healthService)

	s := NewPushService(cfg)
	amsPb.RegisterPushServiceServer(srv, s)

	return srv
}
