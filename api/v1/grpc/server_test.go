package grpc

import (
	"context"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
}

// TestActivateSubscriptionOK tests the normal case where a subscription is added successfully
func (suite *ServerTestSuite) TestActivateSubscriptionOK() {

	ps := NewPushService()

	retry := amsPb.RetryPolicy{
		Type:   "linear",
		Period: 30,
	}

	pCfg := amsPb.PushConfig{
		PushEndpoint: "https://127.0.0.1:5000/receive_here",
		RetryPolicy:  &retry,
	}

	sub := amsPb.Subscription{
		FullName:   "projects/p1/subscription/sub1",
		FullTopic:  "projects/p1/topics/topic1",
		PushConfig: &pCfg,
	}

	s, e := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{Subscription: &sub})

	suite.Equal(&amsPb.ActivateSubscriptionResponse{
		Message: "Subscription projects/p1/subscription/sub1 activated",
	}, s)

	suite.Equal(&sub, ps.Subscriptions["projects/p1/subscription/sub1"])
	suite.Nil(e)
}

// TestActivateSubscriptionCONFLICT tests the case where the subscription is already activated and a conflict is produced
func (suite *ServerTestSuite) TestActivateSubscriptionCONFLICT() {

	ps := NewPushService()
	ps.Subscriptions["conflict_sub"] = &amsPb.Subscription{}
	sub := amsPb.Subscription{FullName: "conflict_sub"}
	s, e := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{Subscription: &sub})

	suite.Equal(status.Error(codes.AlreadyExists, "Subscription conflict_sub is already activated"), e)

	suite.Nil(s)
}

// TestIsSubActive tests the IsSubActive method of PushService for both true and false cases
func (suite *ServerTestSuite) TestIsSubActive() {

	ps := NewPushService()

	ps.Subscriptions["sub1"] = &amsPb.Subscription{}

	suite.True(ps.IsSubActive("sub1"))

	suite.False(ps.IsSubActive("not_active"))
}

// TestNewPushService tests the NewPushService function that returns a *PushService and that its fields are set properly
func (suite *ServerTestSuite) TestNewPushService() {

	ps := NewPushService()

	suite.IsType(&PushService{}, ps)

	// make sure the map containing the subscriptions is initialised
	suite.NotNil(ps.Subscriptions)
}

// TestDeactivateSubscriptionRequest tests the case where a subscription is activated, and deactivates it
func (suite *ServerTestSuite) TestDeactivateSubscriptionRequest() {

	ps := NewPushService()

	ps.Subscriptions["sub1"] = &amsPb.Subscription{}

	s, e := ps.DeactivateSubscription(context.Background(), &amsPb.DeactivateSubscriptionRequest{FullName:"sub1"})

	_, ok := ps.Subscriptions["sub1"]

	suite.Equal(&amsPb.DeactivateSubscriptionResponse{Message:"Subscription sub1 deactivated"}, s)

	suite.False(ok)

	suite.Nil(e)
}

// TestDeactivateSubscriptionRequestNOTFOUND tests the case where the subscription is not yet activated
func (suite *ServerTestSuite) TestDeactivateSubscriptionRequestNOTFOUND() {

	ps := NewPushService()

	s, e := ps.DeactivateSubscription(context.Background(), &amsPb.DeactivateSubscriptionRequest{FullName:"not_found"})

	suite.Equal(status.Error(codes.NotFound, "Subscription not_found is not active"), e)

	suite.Nil(s)
}

func (suite *ServerTestSuite) TestNewGRPCServer() {

	srv := NewGRPCServer()

	suite.IsType(srv, &grpc.Server{})
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
