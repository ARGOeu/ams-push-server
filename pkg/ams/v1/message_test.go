package v1

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type MessageTestSuite struct {
	suite.Suite
}

func (suite *MessageTestSuite) TestNewMessageService() {

	messageService := NewMessageService(AmsBaseInfo{
		Scheme: "https",
		Host:   "localhost",
		Headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    "token",
		}}, new(http.Client))

	suite.NotNil(messageService.client)
	suite.Equal("https", messageService.Scheme)
	suite.Equal("localhost", messageService.Host)
	suite.Equal(map[string]string{
		"Content-type": "application/json",
		"x-api-key":    "token",
	}, messageService.Headers)
}

func (suite *MessageTestSuite) TestPull() {

	mcrt := new(MockConsumeRoundTripper)

	client := &http.Client{
		Transport: mcrt,
	}

	amsClient := NewClient("https", "localhost", "token", 443, client)

	m1, e1 := amsClient.Pull(context.Background(), "/normal_sub", 1, true)

	// check pull options(request body)
	po := PullOptions{}
	json.Unmarshal(mcrt.RequestBodyBytes, &po)

	suite.Equal("some_data", m1.RecMsgs[0].Msg.Data)
	suite.Equal("some_id", m1.RecMsgs[0].Msg.ID)
	suite.Equal("some_ack_id", m1.RecMsgs[0].AckID)
	suite.Equal("1", po.MaxMessages)
	suite.Equal("true", po.ReturnImmediately)
	suite.Nil(e1)

	// check for multiple messages
	po2 := PullOptions{}
	mcrt.RequestBodyBytes = []byte{}
	amsClient.Pull(context.Background(), "/normal_sub", 3, true)
	json.Unmarshal(mcrt.RequestBodyBytes, &po2)
	suite.Equal("3", po2.MaxMessages)
}

func (suite *MessageTestSuite) TestAck() {

	client := &http.Client{
		Transport: new(MockConsumeRoundTripper),
	}

	amsClient := NewClient("https", "localhost", "token", 443, client)

	// test the normal case, where the acknowledgement is successful

	e1 := amsClient.Ack(context.Background(), "/normal_sub", "ackid")

	suite.Nil(e1)
}

func TestMessageTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(MessageTestSuite))
}
