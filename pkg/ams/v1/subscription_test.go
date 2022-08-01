package v1

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type SubscriptionTestSuite struct {
	suite.Suite
}

func (suite *SubscriptionTestSuite) TestNewSubscriptionService() {

	subscriptionService := NewSubscriptionService(AmsBaseInfo{
		Scheme: "https",
		Host:   "localhost",
		Headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    "token",
		}}, new(http.Client))

	suite.NotNil(subscriptionService.client)
	suite.Equal("https", subscriptionService.Scheme)
	suite.Equal("localhost", subscriptionService.Host)
	suite.Equal(map[string]string{
		"Content-type": "application/json",
		"x-api-key":    "token",
	}, subscriptionService.Headers)
}

func (suite *SubscriptionTestSuite) TestIsPushEnabled() {
	sub := Subscription{
		PushCfg: PushConfig{
			Pend: "remote.com",
			Type: HttpEndpointPushConfig,
		},
	}
	suite.True(sub.IsPushEnabled())

	sub2 := Subscription{}
	suite.False(sub2.IsPushEnabled())
}

func (suite *SubscriptionTestSuite) TestGetSubscription() {

	client := &http.Client{
		Transport: new(MockAmsRoundTripper),
	}
	amsClient := NewClient("", "", "", 443, client)

	rp := RetryPolicy{
		PolicyType: "linear",
		Period:     300,
	}

	authz := AuthorizationHeader{
		Value: "auth-header-1",
	}

	pc := PushConfig{
		Type:                HttpEndpointPushConfig,
		Pend:                "example.com:9999",
		AuthorizationHeader: authz,
		RetPol:              rp,
		Base64Decode:        true,
	}

	expectedSub := Subscription{
		FullName:  "/projects/push1/subscriptions/sub1",
		FullTopic: "/projects/push1/topics/t1",
		PushCfg:   pc,
	}

	sub1, err1 := amsClient.GetSubscription(context.TODO(), "/projects/push1/subscriptions/sub1")
	suite.Equal(expectedSub, sub1)
	suite.Nil(err1)

	// error case
	sub2, err2 := amsClient.GetSubscription(context.TODO(), "/projects/push1/subscriptions/errorsub")
	suite.Equal(Subscription{}, sub2)
	suite.Equal("server internal error", err2.Error())
}

func TestSubscriptionTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(SubscriptionTestSuite))
}
