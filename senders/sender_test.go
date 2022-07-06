package senders

import (
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type SenderTestSuite struct {
	suite.Suite
}

// TestNew tests that the sender factory behaves properly
func (suite *SenderTestSuite) TestNew() {

	// normal creation
	pushCFG := amsPb.PushConfig{
		Type:                amsPb.PushType_HTTP_ENDPOINT,
		PushEndpoint:        "example.com",
		AuthorizationHeader: "auth-header-1",
	}
	s1, e1 := New(pushCFG, &http.Client{})
	suite.IsType(&HttpSender{}, s1)
	suite.Nil(e1)

	// normal creation
	pushCFG2 := amsPb.PushConfig{
		Type:                amsPb.PushType_MATTERMOST,
		PushEndpoint:        "example.com",
		AuthorizationHeader: "auth-header-1",
	}
	s2, e2 := New(pushCFG2, &http.Client{})
	suite.IsType(&MattermostSender{}, s2)
	suite.Nil(e2)
}

// TestDetermineMessageFormat tests the DetermineMessageFormat functionality
func (suite *SenderTestSuite) TestDetermineMessageFormat() {
	suite.Equal(SingleMessageFormat, DetermineMessageFormat(1))
	suite.Equal(MultipleMessageFormat, DetermineMessageFormat(30))
}

func TestSenderTestSuite(t *testing.T) {
	suite.Run(t, new(SenderTestSuite))
}
