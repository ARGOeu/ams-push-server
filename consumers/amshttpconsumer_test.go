package consumers

import (
	"context"
	"encoding/json"
	"errors"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type AmsHttpConsumerTestSuite struct {
	suite.Suite
}

// TestNewAmsHttpConsumer tests the proper initialization of an ams http consumer
func (suite *AmsHttpConsumerTestSuite) TestNewAmsHttpConsumer() {

	ahc := NewAmsHttpConsumer("/projects/p1/subscriptions/s1",
		ams.NewClient("https", "example.com", "token", 443, new(http.Client)))

	suite.Equal("example.com:443", ahc.amsClient.Host())
	suite.Equal("/projects/p1/subscriptions/s1", ahc.fullSub)
	suite.NotNil(ahc.amsClient)
}

// TestConsume tests various behaviors of the consume functionality
func (suite *AmsHttpConsumerTestSuite) TestConsume() {

	mcrt := new(ams.MockConsumeRoundTripper)

	client := &http.Client{
		Transport: mcrt,
	}

	amsClient := ams.NewClient("https", "localhost", "token", 443, client)

	// test the normal case, where the consume method will return new messages
	acl := NewAmsHttpConsumer("/normal_sub", amsClient)

	m1, e1 := acl.Consume(context.Background(), 1)

	// check pull options(request body)
	po := ams.PullOptions{}
	json.Unmarshal(mcrt.RequestBodyBytes, &po)

	suite.Equal("some_data", m1.RecMsgs[0].Msg.Data)
	suite.Equal("some_id", m1.RecMsgs[0].Msg.ID)
	suite.Equal("some_ack_id", m1.RecMsgs[0].AckID)
	suite.Equal("1", po.MaxMessages)
	suite.Nil(e1)

	// check for multiple messages
	po2 := ams.PullOptions{}
	mcrt.RequestBodyBytes = []byte{}
	acl.Consume(context.Background(), 3)
	json.Unmarshal(mcrt.RequestBodyBytes, &po2)
	suite.Equal("3", po2.MaxMessages)

	// test the case where there are no new messages
	acl2 := NewAmsHttpConsumer("/empty_sub", amsClient)

	m2, e2 := acl2.Consume(context.Background(), 1)

	suite.Equal(0, len(m2.RecMsgs))
	suite.Equal("no new messages", e2.Error())

	// test the case where an error occurred while interacting with ams
	acl3 := NewAmsHttpConsumer("/error_sub", amsClient)

	_, e3 := acl3.Consume(context.Background(), 1)

	expOut := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`

	suite.Equal(expOut, e3.Error())

	// test the case where an error occurred while interacting with ams(project not found)
	acl4 := NewAmsHttpConsumer("/error_sub_no_project", amsClient)

	_, e4 := acl4.Consume(context.Background(), 1)

	expOut2 := `{
		 "error": {
			"code": 404,
			"message": "project doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`

	suite.Equal(expOut2, e4.Error())

	// test the case where an error occurred while interacting with ams(subscription not found)
	acl5 := NewAmsHttpConsumer("/error_sub_no_sub", amsClient)

	_, e5 := acl5.Consume(context.Background(), 1)

	expOut3 := `{
		 "error": {
			"code": 404,
			"message": "Subscription doesn't exist",
			"status": "NOT_FOUND"
		 } 
		}`

	suite.Equal(expOut3, e5.Error())
}

func (suite *AmsHttpConsumerTestSuite) TestResourceInfo() {

	ahc := NewAmsHttpConsumer("/projects/p1/subscriptions/s1",
		ams.NewClient("https", "example.com", "token", 443, new(http.Client)))

	suite.Equal("subscription /projects/p1/subscriptions/s1 from example.com:443", ahc.ResourceInfo())
}

// TestAck tests the proper Ack functionality
func (suite *AmsHttpConsumerTestSuite) TestAck() {

	client := &http.Client{
		Transport: new(ams.MockConsumeRoundTripper),
	}

	amsClient := ams.NewClient("https", "localhost", "token", 443, client)

	// test the normal case, where the acknowledgement is successful
	acl := NewAmsHttpConsumer("/normal_sub", amsClient)

	e1 := acl.Ack(context.Background(), "ackid")

	suite.Nil(e1)

	// test the normal case, where the acknowledgement has timed out
	acl2 := NewAmsHttpConsumer("/timeout_sub", amsClient)

	e2 := acl2.Ack(context.Background(), "ackid-15")

	expOut := `an error occurred while trying to acknowledge message with ackId ackid-15 from subscription /timeout_sub from localhost:443, {
		 "error": {
			"code": 408,
			"message": "ack timeout",
			"status": "TIMEOUT"
		 }
		}`

	suite.Equal(expOut, e2.Error())
}

func (suite *AmsHttpConsumerTestSuite) TestToCancelableError() {

	c := NewAmsHttpConsumer("normal_sub", nil)

	errorMsg1 := `{
		 "error": {
			"code": 404,
			"message": "Subscription doesn't exist",
			"status": "NOT_FOUND"
		 } 
		}`

	ce1, ok1 := c.ToCancelableError(errors.New(errorMsg1))

	// normal case with sub doesn't exist error
	suite.True(ok1)
	suite.Equal("Subscription doesn't exist", ce1.ErrMsg)
	suite.Equal("normal_sub", ce1.Resource)

	errorMsg2 := `{
		 "error": {
			"code": 404,
			"message": "project doesn't exist",
			"status": "NOT_FOUND"
		 } 
		}`

	ce2, ok2 := c.ToCancelableError(errors.New(errorMsg2))

	// normal case with project doesn't exist error
	suite.True(ok2)
	suite.Equal("project doesn't exist", ce2.ErrMsg)
	suite.Equal("normal_sub", ce2.Resource)

	// case where the error is not a cancelable one
	errorMsg3 := `some random error`
	ce3, ok3 := c.ToCancelableError(errors.New(errorMsg3))
	suite.False(ok3)
	suite.Equal(CancelableError{}, ce3)
}

func TestAmsHttpConsumerTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(AmsHttpConsumerTestSuite))
}
