package consumers

import (
	"context"
	"errors"
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

	ahc := NewAmsHttpConsumer("example.com:443", "/projects/p1/subscriptions/s1", "some_token", new(http.Client))

	suite.Equal("example.com:443", ahc.endpoint)
	suite.Equal("/projects/p1/subscriptions/s1", ahc.fullSub)
	suite.Equal("some_token", ahc.token)
	suite.Equal(new(http.Client), ahc.client)
}

// TestConsume tests various behaviors of the consume functionality
func (suite *AmsHttpConsumerTestSuite) TestConsume() {

	client := &http.Client{
		Transport: new(MockConsumeRoundTripper),
	}

	// test the normal case, where the consume method will return new messages
	acl := NewAmsHttpConsumer("", "/normal_sub", "", client)

	m1, e1 := acl.Consume(context.Background())

	suite.Equal("some_data", m1.RecMsgs[0].Msg.Data)
	suite.Equal("some_id", m1.RecMsgs[0].Msg.ID)
	suite.Equal("some_ack_id", m1.RecMsgs[0].AckID)
	suite.Nil(e1)

	// test the case where there are no new messages
	acl2 := NewAmsHttpConsumer("", "/empty_sub", "", client)

	m2, e2 := acl2.Consume(context.Background())

	suite.Equal(0, len(m2.RecMsgs))
	suite.Equal("no new messages", e2.Error())

	// test the case where an error occurred while interacting with ams
	acl3 := NewAmsHttpConsumer("e1", "/error_sub", "", client)

	_, e3 := acl3.Consume(context.Background())

	expOut := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`

	suite.Equal(expOut, e3.Error())

	// test the case where an error occurred while interacting with ams(project not found)
	acl4 := NewAmsHttpConsumer("e1", "/error_sub_no_project", "", client)

	_, e4 := acl4.Consume(context.Background())

	expOut2 := `{
		 "error": {
			"code": 404,
			"message": "project doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`

	suite.Equal(expOut2, e4.Error())

	// test the case where an error occurred while interacting with ams(subscription not found)
	acl5 := NewAmsHttpConsumer("e1", "/error_sub_no_sub", "", client)

	_, e5 := acl5.Consume(context.Background())

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

	ahc := NewAmsHttpConsumer("example.com:443", "/projects/p1/subscriptions/s1", "some_token", new(http.Client))

	suite.Equal("subscription /projects/p1/subscriptions/s1 from example.com:443", ahc.ResourceInfo())
}

// TestAck tests the proper Ack functionality
func (suite *AmsHttpConsumerTestSuite) TestAck() {

	client := &http.Client{
		Transport: new(MockConsumeRoundTripper),
	}

	// test the normal case, where the acknowledgement is successful
	acl := NewAmsHttpConsumer("", "/normal_sub", "", client)

	e1 := acl.Ack(context.Background(), "ackid")

	suite.Nil(e1)

	// test the normal case, where the acknowledgement has timed out
	acl2 := NewAmsHttpConsumer("e1", "/timeout_sub", "", client)

	e2 := acl2.Ack(context.Background(), "ackid-15")

	expOut := `an error occurred while trying to acknowledge message with ackId ackid-15 from subscription /timeout_sub from e1, {
		 "error": {
			"code": 408,
			"message": "ack timeout",
			"status": "TIMEOUT"
		 }
		}`

	suite.Equal(expOut, e2.Error())
}

func (suite *AmsHttpConsumerTestSuite) TestToCancelableError() {

	c := NewAmsHttpConsumer("", "normal_sub", "", nil)

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

func (suite *AmsHttpConsumerTestSuite) TestUpdateResourceStatus() {

	client := &http.Client{
		Transport: new(MockConsumeRoundTripper),
	}

	// normal case
	c := NewAmsHttpConsumer("e2", "/normal_sub", "", client)

	e1 := c.UpdateResourceStatus(context.Background(), "")

	suite.Nil(e1)

	// test the case where an error occurred while interacting with ams
	c = NewAmsHttpConsumer("e2", "/error_sub", "", client)

	e2 := c.UpdateResourceStatus(context.Background(), "")

	expOut := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`

	suite.Equal(expOut, e2.Error())
}

func TestAmsHttpConsumerTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(AmsHttpConsumerTestSuite))
}
