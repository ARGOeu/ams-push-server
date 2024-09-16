package grpc

import (
	"context"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/config"
	"github.com/ARGOeu/ams-push-server/consumers"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/ARGOeu/ams-push-server/push"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
}

func (suite *ServerTestSuite) TestStatus() {

	ps := &PushService{}

	ps.status = "not ok"

	_, e1 := ps.Status(context.Background(), &amsPb.StatusRequest{})
	suite.Equal(status.Error(codes.Internal, "The push service is currently unable to handle any requests.not ok"), e1)

	ps.status = "ok"
	_, e2 := ps.Status(context.Background(), &amsPb.StatusRequest{})
	suite.Nil(e2)

}

// TestActivateSubscriptionOK tests the normal case where a subscription is added successfully
func (suite *ServerTestSuite) TestActivateSubscriptionOK() {

	ps := NewPushService(config.NewMockConfig())

	retry := amsPb.RetryPolicy{
		Type:   "linear",
		Period: 30,
	}

	pCfg := amsPb.PushConfig{
		Type:         amsPb.PushType_HTTP_ENDPOINT,
		PushEndpoint: "https://127.0.0.1:5000/receive_here",
		RetryPolicy:  &retry,
	}

	sub := amsPb.Subscription{
		FullName:   "/projects/p1/subscription/sub1",
		FullTopic:  "/projects/p1/topics/topic1",
		PushConfig: &pCfg,
	}

	s, e := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{Subscription: &sub})

	suite.Equal(&amsPb.ActivateSubscriptionResponse{
		Message: "Subscription /projects/p1/subscription/sub1 activated",
	}, s)
	suite.Nil(e)

	lw, _ := ps.PushWorkers["/projects/p1/subscription/sub1"]
	suite.Equal(&sub, lw.Subscription())
}

// TestActivateSubscriptionNIL tests the case where the provided subscription is invalid
func (suite *ServerTestSuite) TestActivateSubscriptionInvalidArgument() {

	ps := NewPushService(config.NewMockConfig())

	// invalid argument through nil subscription
	s, e := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{Subscription: nil})

	suite.Equal(status.Error(codes.InvalidArgument, "Empty subscription"), e)

	suite.Nil(s)

	// invalid argument through unimplemented worker type
	s1, e1 := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{
		Subscription: &amsPb.Subscription{
			PushConfig: &amsPb.PushConfig{
				PushEndpoint: "https://example.com",
				RetryPolicy: &amsPb.RetryPolicy{
					Type: "unknown",
				},
			},
		}})

	suite.Equal(status.Error(codes.InvalidArgument, "Invalid argument, worker unknown not yet implemented"), e1)

	suite.Nil(s1)
}

// TestActivateSubscriptionCONFLICT tests the case where the subscription is already activated and a conflict is produced
func (suite *ServerTestSuite) TestActivateSubscriptionCONFLICT() {

	ps := NewPushService(config.NewMockConfig())
	ps.PushWorkers["conflict_sub"] = new(push.MockWorker)
	sub := amsPb.Subscription{
		FullName: "conflict_sub",
		PushConfig: &amsPb.PushConfig{
			RetryPolicy: &amsPb.RetryPolicy{},
		}}
	s, e := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{Subscription: &sub})

	suite.Equal(status.Error(codes.AlreadyExists, "Subscription conflict_sub is already activated"), e)

	suite.Nil(s)
}

func (suite *ServerTestSuite) TestSubscriptionStatus() {

	ps := NewPushService(config.NewMockConfig())
	sub := amsPb.Subscription{
		FullName: "sub1",
		PushConfig: &amsPb.PushConfig{
			RetryPolicy: &amsPb.RetryPolicy{},
		}}
	// not found case
	s, e := ps.SubscriptionStatus(context.Background(), &amsPb.SubscriptionStatusRequest{FullName: "sub1"})

	suite.Equal(status.Error(codes.NotFound, "Subscription sub1 is not active"), e)

	suite.Nil(s)

	// normal case
	ps.PushWorkers["sub1"] = &push.MockWorker{
		Sub:       sub,
		SubStatus: "ok",
	}

	s2, e2 := ps.SubscriptionStatus(context.Background(), &amsPb.SubscriptionStatusRequest{FullName: "sub1"})

	suite.Equal(&amsPb.SubscriptionStatusResponse{
		Status: "ok",
	}, s2)

	suite.Nil(e2)
}

// TestIsSubActive tests the IsSubActive method of PushService for both true and false cases
func (suite *ServerTestSuite) TestIsSubActive() {

	ps := NewPushService(config.NewMockConfig())

	ps.PushWorkers["sub1"] = new(push.MockWorker)

	suite.True(ps.IsSubActive("sub1"))

	suite.False(ps.IsSubActive("not_active"))
}

// TestHandleDeactivateChannel tests the that when a cancelable error is send to the channel
// the respective subscription is deleted
func (suite *ServerTestSuite) TestHandleDeactivateChannel() {

	ps := NewPushService(config.NewMockConfig())

	ps.PushWorkers["sub1"] = new(push.MockWorker)

	ps.deactivateChan <- consumers.CancelableError{
		ErrMsg:   "cancel",
		Resource: "sub1",
	}

	_, found := ps.PushWorkers["sub1"]

	suite.False(found)

}

func (suite *ServerTestSuite) TestDeactivateSubscription() {

	ps := NewPushService(config.NewMockConfig())
	mw := new(push.MockWorker)
	ps.PushWorkers["sub1"] = mw

	e1 := ps.deactivateSubscription("sub1")
	_, found := ps.PushWorkers["sub1"]

	// test normal case(delete entry from map, deactivate worker)
	suite.Equal("stopped", mw.Status())
	suite.False(found)
	suite.Nil(e1)

	e2 := ps.deactivateSubscription("unknown")

	// test the case where the sub is not active
	suite.Equal("Subscription unknown is not active", e2.Error())
}

// TestNewPushService tests the NewPushService function that returns a *PushService and that its fields are set properly
func (suite *ServerTestSuite) TestNewPushService() {

	ps := NewPushService(config.NewMockConfig())

	suite.IsType(&PushService{}, ps)

	// make sure the map containing the subscriptions is initialised
	suite.NotNil(ps.PushWorkers)
}

func (suite *ServerTestSuite) TestDeactivateSubscriptionRequest() {

	ps := NewPushService(config.NewMockConfig())

	ps.PushWorkers["sub1"] = new(push.MockWorker)

	s, e := ps.DeactivateSubscription(context.Background(), &amsPb.DeactivateSubscriptionRequest{FullName: "sub1"})

	_, ok := ps.PushWorkers["sub1"]
	suite.Equal(&amsPb.DeactivateSubscriptionResponse{Message: "Subscription sub1 deactivated"}, s)

	suite.False(ok)
	suite.Nil(e)
}

// TestDeactivateSubscriptionRequestNOTFOUND tests the case where the subscription is not yet activated
func (suite *ServerTestSuite) TestDeactivateSubscriptionRequestNOTFOUND() {

	ps := NewPushService(config.NewMockConfig())

	s, e := ps.DeactivateSubscription(context.Background(), &amsPb.DeactivateSubscriptionRequest{FullName: "not_found"})

	suite.Equal(status.Error(codes.NotFound, "Subscription not_found is not active"), e)
	suite.Nil(s)
}

func (suite *ServerTestSuite) TestNewGRPCServer() {

	srv := NewGRPCServer(config.NewMockConfig())

	suite.IsType(&grpc.Server{}, srv)
}

func (suite *ServerTestSuite) TestLoadSubscriptions() {

	ps := NewPushService(config.NewMockConfig())
	client := &http.Client{
		Transport: new(ams.MockAmsRoundTripper),
	}
	ps.Client = client
	amsClient := ams.NewClient("", "", "", 443, client)
	ps.AmsClient = amsClient

	ps.loadSubscriptions()

	// since there was no problem retrieving the ams user, status should be ok
	suite.Equal("ok", ps.status)

	// normal case, sub1 is push enabled and it should be activated successfully
	_, sub1Found := ps.PushWorkers["/projects/push1/subscriptions/sub1"]
	suite.True(sub1Found)

	// normal case, sub4 is push enabled and it should be activated successfully
	_, sub4Found := ps.PushWorkers["/projects/push2/subscriptions/sub4"]
	suite.True(sub4Found)

	// normal case, sub5 is mattermost push enabled and it should be activated successfully
	_, sub5Found := ps.PushWorkers["/projects/push2/subscriptions/sub5"]
	suite.True(sub5Found)

	// error case, the subscription should not have been activated
	_, errorSubFound := ps.PushWorkers["/projects/push1/subscriptions/errorsub"]
	suite.False(errorSubFound)

	// the subscription is not push enabled case, the subscription should not have been activated
	_, sub3SubFound := ps.PushWorkers["/projects/push2/subscriptions/sub3"]
	suite.False(sub3SubFound)
}

func TestServerTestSuite(t *testing.T) {
	logrus.SetOutput(io.Discard)
	suite.Run(t, new(ServerTestSuite))
}
