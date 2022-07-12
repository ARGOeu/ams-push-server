package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/config"
	"github.com/ARGOeu/ams-push-server/consumers"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
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

const ServiceUnavailable = "The push service is currently unable to handle any requests"

// PushService holds all the the information and functionality regarding the push implementation
type PushService struct {
	Cfg            *config.Config
	Client         *http.Client
	AmsClient      *ams.Client
	PushWorkers    map[string]push.Worker
	deactivateChan chan consumers.CancelableError
	status         string
}

// NewPushService returns a pointer to a PushService and initialises its fields
func NewPushService(cfg *config.Config) *PushService {

	ps := new(PushService)

	ps.Cfg = cfg
	ps.PushWorkers = make(map[string]push.Worker)

	// build the client
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !ps.Cfg.VerifySSL,
		},
	}

	client := &http.Client{
		Transport: transCfg,
		Timeout:   time.Duration(30 * time.Second),
	}

	ps.Client = client
	ps.AmsClient = ams.NewClient("https", ps.Cfg.AmsHost, ps.Cfg.AmsToken, ps.Cfg.AmsPort, client)

	ps.deactivateChan = make(chan consumers.CancelableError)
	go ps.handleDeactivateChannel()

	if !cfg.SkipSubsLoad {
		go ps.loadSubscriptions()
	}

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

// Status returns the stat of the service, whether or not it is functioning properly
func (ps *PushService) Status(context.Context, *amsPb.StatusRequest) (*amsPb.StatusResponse, error) {

	if ps.status != "ok" {
		return &amsPb.StatusResponse{}, status.Errorf(codes.Internal, "%v.%v", ServiceUnavailable, ps.status)
	}

	return &amsPb.StatusResponse{}, nil
}

// SubscriptionStatus returns the status of the worker that handles the respective subscription
func (ps *PushService) SubscriptionStatus(ctx context.Context, r *amsPb.SubscriptionStatusRequest) (*amsPb.SubscriptionStatusResponse, error) {

	if !ps.IsSubActive(r.FullName) {
		return nil, status.Errorf(codes.NotFound, "Subscription %v is not active", r.FullName)
	}

	w := ps.PushWorkers[r.FullName]

	return &amsPb.SubscriptionStatusResponse{
		Status: w.Status(),
	}, nil

}

// ActivateSubscription activates a subscription so the service can start handling the push functionality
func (ps *PushService) ActivateSubscription(ctx context.Context, r *amsPb.ActivateSubscriptionRequest) (*amsPb.ActivateSubscriptionResponse, error) {

	if r.Subscription == nil || r.Subscription.PushConfig == nil || r.Subscription.PushConfig.RetryPolicy == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Empty subscription")
	}

	if ps.IsSubActive(r.Subscription.FullName) {
		return nil, status.Errorf(codes.AlreadyExists, "Subscription %v is already activated", r.Subscription.FullName)
	}

	if r.Subscription.PushConfig.Type == amsPb.PushType_HTTP_ENDPOINT {
		_, err := url.ParseRequestURI(r.Subscription.PushConfig.PushEndpoint)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid push endpoint, %v", err.Error())
		}
	}

	// choose a consumer
	c, _ := consumers.New(consumers.AmsHttpConsumerType, r.Subscription.FullName, ps.AmsClient)

	// choose a sender
	s, _ := senders.New(*r.Subscription.PushConfig, ps.Client)

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

	logOpts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}

	s := NewPushService(cfg)

	srvOptions := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(grpcLogger), logOpts...),
			AuthInterceptor(cfg.ACL, cfg.TLSEnabled),
			StatusInterceptor(s),
		),
	}

	if cfg.TLSEnabled {
		srvOptions = append(srvOptions, grpc.Creds(credentials.NewTLS(cfg.GetTLSConfig())))
	}

	srv := grpc.NewServer(srvOptions...)

	healthService := health.NewServer()
	healthService.SetServingStatus("", gRPCHealth.HealthCheckResponse_SERVING)
	gRPCHealth.RegisterHealthServer(srv, healthService)

	amsPb.RegisterPushServiceServer(srv, s)

	return srv
}

// loadSubscriptions activates all the ams subscriptions that are push enabled and assigned to the current push worker
func (ps *PushService) loadSubscriptions() {

	var userInfo ams.UserInfo
	var err error

	userFound := false

	// attempt to retrieve the push worker user
	for !userFound {
		t1 := time.Now()
		userInfo, err = ps.AmsClient.GetUserByToken(context.Background(), ps.Cfg.AmsToken)
		if err != nil {
			ps.status = "Could not retrieve push worker user"
			log.WithFields(
				log.Fields{
					"type":  "system_log",
					"error": err.Error(),
				},
			).Error("Could not retrieve push worker user")
			continue
		}
		userFound = true
		log.WithFields(
			log.Fields{
				"type":            "performance_log",
				"user":            userInfo.Name,
				"processing_time": time.Since(t1).String(),
			},
		).Info("Push worker user retrieved successfully")
	}

	ps.status = "ok"

	for _, project := range userInfo.Projects {

		for _, subName := range project.Subscriptions {

			fullSubName := fmt.Sprintf("/projects/%v/subscriptions/%v", project.Project, subName)

			t1 := time.Now()
			sub, err := ps.AmsClient.GetSubscription(context.Background(), fullSubName)
			if err != nil {
				log.WithFields(
					log.Fields{
						"type":         "system_log",
						"subscription": fullSubName,
						"error":        err.Error(),
					},
				).Error("Could not retrieve subscription")
				continue
			}

			if !sub.IsPushEnabled() {
				log.WithFields(
					log.Fields{
						"type":         "system_log",
						"subscription": fullSubName,
					},
				).Error("Subscription is not push enabled")
				continue
			}

			log.WithFields(
				log.Fields{
					"type":            "performance_log",
					"subscription":    sub.FullName,
					"processing_time": time.Since(t1).String(),
				},
			).Info("Subscription retrieved successfully")

			var pushType amsPb.PushType
			if sub.PushCfg.Type == ams.HttpEndpointPushConfig {
				pushType = amsPb.PushType_HTTP_ENDPOINT
			} else {
				pushType = amsPb.PushType_MATTERMOST
			}

			_, err = ps.ActivateSubscription(context.TODO(),
				&amsPb.ActivateSubscriptionRequest{
					Subscription: &amsPb.Subscription{
						FullName:  sub.FullName,
						FullTopic: sub.FullTopic,
						PushConfig: &amsPb.PushConfig{
							Type:                pushType,
							PushEndpoint:        sub.PushCfg.Pend,
							AuthorizationHeader: sub.PushCfg.AuthorizationHeader.Value,
							MaxMessages:         sub.PushCfg.MaxMessages,
							RetryPolicy: &amsPb.RetryPolicy{
								Period: sub.PushCfg.RetPol.Period,
								Type:   sub.PushCfg.RetPol.PolicyType,
							},
							MattermostUrl:      sub.PushCfg.MattermostUrl,
							MattermostUsername: sub.PushCfg.MattermostUsername,
							MattermostChannel:  sub.PushCfg.MattermostChannel,
							Base_64Decode:      sub.PushCfg.Base64Decode,
						},
					},
				},
			)

			if err != nil {
				log.WithFields(
					log.Fields{
						"type":         "system_log",
						"subscription": fullSubName,
						"error":        err.Error(),
					},
				).Error("Could not activate subscription")
				continue
			}

			log.WithFields(
				log.Fields{
					"type":         "system_log",
					"subscription": fullSubName,
				},
			).Info("Subscription activated successfully")
		}
	}
}
