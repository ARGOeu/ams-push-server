package senders

import (
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
	s1, e1 := New(HttpSenderType, "example.com", &http.Client{})
	suite.IsType(&HttpSender{}, s1)
	suite.Nil(e1)

	// unimplemented sender
	_, e2 := New("unknown", "", nil)
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
