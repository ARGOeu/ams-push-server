package senders

import (
	"github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
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
	pushCFG := proto.PushConfig{
		PushEndpoint:        "example.com",
		AuthorizationHeader: "auth-header-1",
	}
	s1, e1 := New(HttpSenderType, pushCFG, &http.Client{})
	suite.IsType(&HttpSender{}, s1)
	suite.Nil(e1)

	// unimplemented sender
	_, e2 := New("unknown", pushCFG, nil)
	suite.Equal("sender unknown not yet implemented", e2.Error())

}

// TestDetermineMessageFormat tests the DetermineMessageFormat functionality
func (suite *SenderTestSuite) TestDetermineMessageFormat() {
	suite.Equal(SingleMessageFormat, DetermineMessageFormat(1))
	suite.Equal(MultipleMessageFormat, DetermineMessageFormat(30))
}

func TestSenderTestSuite(t *testing.T) {
	suite.Run(t, new(SenderTestSuite))
}
