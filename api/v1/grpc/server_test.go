package grpc

import (
	"context"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
}

func (suite *ServerTestSuite) TestActivateSubscription() {

	ps := new(pushService)

	retry1 := amsPb.RetryPolicy{Type: "linear", Period: 30}
	pCfg1 := amsPb.PushConfig{PushEndpoint: "https://127.0.0.1:5000/receive_here", RetryPolicy: &retry1}
	sub1 := amsPb.Subscription{FullName: "projects/p1/subscription/sub1", FullTopic: "projects/p1/topics/topic1", PushConfig: &pCfg1}
	s1, e1 := ps.ActivateSubscription(context.Background(), &amsPb.ActivateSubscriptionRequest{Subscription: &sub1})

	suite.Equal(s1, &amsPb.Status{})

	suite.Nil(e1)
}

func (suite *ServerTestSuite) TestDeactivateSubscriptionRequest() {

	ps := new(pushService)

	s1, e1 := ps.DeactivateSubscription(context.Background(), &amsPb.DeactivateSubscriptionRequest{FullName: "projects/p1/subscription/sub1"})

	suite.Equal(s1, &amsPb.Status{})

	suite.Nil(e1)
}

func (suite *ServerTestSuite) TestNewGRPCServer() {

	srv := NewGRPCServer()

	suite.IsType(srv, &grpc.Server{})

}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))

}
