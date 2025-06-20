package senders

import (
	"context"
	v1 "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"testing"
)

type MattermostSenderTestSuite struct {
	suite.Suite
}

// TestNewMattermostSender tests the proper initialisation of a mattermost sender
func (suite *HttpSenderTestSuite) TestNewMattermostSender() {

	m := NewMattermostSender("https://example.com:8080/webhook-1",
		"mattermost", "ops", new(http.Client))

	suite.Equal("https://example.com:8080/webhook-1", m.webhookUrl)
	suite.Equal("mattermost", m.username)
	suite.Equal("ops", m.channel)
	suite.Equal(new(http.Client), m.client)
}

// TestSend tests the send functionality
func (suite *MattermostSenderTestSuite) TestSend() {

	mmrt := new(MockMattermostRoundTripper)

	client := &http.Client{
		Transport: mmrt,
	}

	// test the normal case of 200
	m := NewMattermostSender("https://example.com/webhook", "mattermost", "ops", client)
	m1 := PushMsg{Sub: "sub", Msg: v1.Message{
		Data: "ops-data",
	}}
	m1s := PushMsgs{Messages: []PushMsg{m1}}
	e := m.Send(context.Background(), m1s, SingleMessageFormat)
	suite.Nil(e)
	suite.Equal("mattermost", mmrt.Message.Username)
	suite.Equal("ops", mmrt.Message.Channel)
	suite.Equal("ops-data", mmrt.Message.Text)

	// case with mattermost error
	m2 := NewMattermostSender("https://example.com/mattermost-error", "mattermost", "ops", client)
	e2 := m2.Send(context.Background(), m1s, SingleMessageFormat)
	suite.Equal("Couldn't find the channel.", e2.Error())

	// case with generic error
	m3 := NewMattermostSender("https://example.com/generic-error", "mattermost", "ops", client)
	e3 := m3.Send(context.Background(), m1s, SingleMessageFormat)
	suite.Equal("generic-error", e3.Error())
}

func (suite *MattermostSenderTestSuite) TestMattermostError() {
	e1 := MattermostError{
		Message:       "message",
		DetailedError: "error",
	}
	suite.Equal("message.error", e1.Error())

	e2 := MattermostError{
		Message: "message",
	}
	suite.Equal("message", e2.Error())
}

func (suite *MattermostSenderTestSuite) TestDestination() {
	m := NewMattermostSender("https://example.com/webhook", "mattermost", "ops", nil)
	suite.Equal("https://example.com/webhook", m.Destination())
}

func TestMattermostSenderTestSuite(t *testing.T) {
	logrus.SetOutput(io.Discard)
	suite.Run(t, new(MattermostSenderTestSuite))
}
