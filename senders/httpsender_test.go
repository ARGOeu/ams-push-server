package senders

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"testing"
)

type HttpSenderTestSuite struct {
	suite.Suite
}

// TestNewHttpSender tests the proper initialisation of an http sender
func (suite *HttpSenderTestSuite) TestNewHttpSender() {

	s := NewHttpSender("example.com:443", "auth-header-1", new(http.Client))

	suite.Equal("example.com:443", s.endpoint)
	suite.Equal("auth-header-1", s.authZHeader)
	suite.Equal(new(http.Client), s.client)
}

// TestSend tests the send functionality
func (suite *HttpSenderTestSuite) TestSend() {

	msrt := new(MockSenderRoundTripper)

	client := &http.Client{
		Transport: msrt,
	}

	// test the normal case of 200
	s1 := NewHttpSender("https://example.com:8080/receive_here_200", "auth-header-1", client)
	m1 := PushMsg{Sub: "sub"}
	m1s := PushMsgs{Messages: []PushMsg{m1}}
	e1 := s1.Send(context.Background(), m1s, MultipleMessageFormat)

	// check that the format is of multiple messages
	// marshal the request body
	expP1 := PushMsgs{Messages: []PushMsg{}}
	json.Unmarshal(msrt.RequestBodyBytes, &expP1)
	suite.Equal(m1s, expP1)
	suite.Nil(e1)

	// check the format of single message
	msrt.RequestBodyBytes = []byte{}
	e1Single := s1.Send(context.Background(), m1s, SingleMessageFormat)
	expP1Single := PushMsg{Sub: "sub"}
	json.Unmarshal(msrt.RequestBodyBytes, &expP1Single)
	suite.Equal(m1, expP1Single)
	suite.Nil(e1Single)

	// check also that the "messages" field is not present
	m := make(map[string]interface{})
	json.Unmarshal(msrt.RequestBodyBytes, &m)
	_, ok := m["messages"]
	suite.False(ok)

	// test the normal case of 201
	s2 := NewHttpSender("https://example.com:8080/receive_here_201", "", client)
	e2 := s2.Send(context.Background(), PushMsgs{}, MultipleMessageFormat)
	suite.Nil(e2)

	// test the normal case of 204
	s3 := NewHttpSender("https://example.com:8080/receive_here_204", "auth-header-1", client)
	e3 := s3.Send(context.Background(), PushMsgs{}, MultipleMessageFormat)
	suite.Nil(e3)

	// test the normal case of 102
	s4 := NewHttpSender("https://example.com:8080/receive_here_102", "auth-header-1", client)
	e4 := s4.Send(context.Background(), PushMsgs{}, MultipleMessageFormat)
	suite.Nil(e4)

	// test the error case
	s5 := NewHttpSender("https://example.com:8080/receive_here_error", "", client)
	e5 := s5.Send(context.Background(), PushMsgs{}, MultipleMessageFormat)

	expOut := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`

	suite.Equal(expOut, e5.Error())
}

func (suite *HttpSenderTestSuite) TestDestination() {
	s := NewHttpSender("example.com:443", "auth-header-1", nil)
	suite.Equal("example.com:443", s.Destination())
}

func TestHttpSenderTestSuite(t *testing.T) {
	logrus.SetOutput(io.Discard)
	suite.Run(t, new(HttpSenderTestSuite))
}
