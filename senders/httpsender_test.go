package senders

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type HttpSenderTestSuite struct {
	suite.Suite
}

// TestNewHttpSender tests the proper initialisation of an http sender
func (suite *HttpSenderTestSuite) TestNewHttpSender() {

	s := NewHttpSender("example.com:443", new(http.Client))

	suite.Equal("example.com:443", s.endpoint)
	suite.Equal(new(http.Client), s.client)
}

// TestSend tests the send functionality
func (suite *HttpSenderTestSuite) TestSend() {

	client := &http.Client{
		Transport: new(MockSenderRoundTripper),
	}

	// test the normal case of 200
	s1 := NewHttpSender("https://example.com:8080/receive_here_200", client)

	e1 := s1.Send(context.Background(), PushMsg{})

	suite.Nil(e1)

	// test the normal case of 201
	s2 := NewHttpSender("https://example.com:8080/receive_here_201", client)
	e2 := s2.Send(context.Background(), PushMsg{})
	suite.Nil(e2)

	// test the normal case of 204
	s3 := NewHttpSender("https://example.com:8080/receive_here_204", client)
	e3 := s3.Send(context.Background(), PushMsg{})
	suite.Nil(e3)

	// test the normal case of 102
	s4 := NewHttpSender("https://example.com:8080/receive_here_102", client)
	e4 := s4.Send(context.Background(), PushMsg{})
	suite.Nil(e4)

	// test the error case
	s5 := NewHttpSender("https://example.com:8080/receive_here_error", client)
	e5 := s5.Send(context.Background(), PushMsg{})

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
	s := NewHttpSender("example.com:443", nil)
	suite.Equal("example.com:443", s.Destination())
}

func TestHttpSenderTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(HttpSenderTestSuite))
}
