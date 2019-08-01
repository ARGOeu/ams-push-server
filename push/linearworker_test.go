package push

import (
	"context"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
	"github.com/ARGOeu/ams-push-server/senders"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type LinearWorkerTestSuite struct {
	suite.Suite
}

// TestNewLinearWorker tests the new linear worker function
func (suite *LinearWorkerTestSuite) TestNewLinearWorker() {

	c := consumers.NewAmsHttpConsumer("", "", "", &http.Client{})
	s := senders.NewHttpSender("", &http.Client{})
	sub := &amsPb.Subscription{}

	lw := NewLinearWorker(sub, c, s, make(chan consumers.CancelableError))

	suite.Equal(sub, lw.sub)
	suite.Equal(c, lw.consumer)
	suite.Equal(s, lw.sender)
	suite.NotNil(lw.cancel)
	suite.NotNil(lw.ctx)

}

// TestStartStopCycle checks the correct functionality of starting and stopping a push worker
func (suite *LinearWorkerTestSuite) TestStartStopCycle() {

	// how often does a push cycle occurs in milliseconds
	// a full push cycle: consume -> send -> ack
	rate := uint32(300)

	// for how long the push worker will perform in seconds
	workerLifeTime := 1

	ctx, cancel := context.WithCancel(context.TODO())
	sub := &amsPb.Subscription{
		PushConfig: &amsPb.PushConfig{
			RetryPolicy: &amsPb.RetryPolicy{
				Period: rate,
			},
		},
	}

	c := new(consumers.MockConsumer)
	c.SubStatus = "normal_sub"
	c.AckStatus = "normal_ack"
	s := new(senders.MockSender)

	lw := LinearWorker{
		sub:      sub,
		consumer: c,
		sender:   s,
		ctx:      ctx,
		cancel:   cancel,
	}

	// signal the push worker to stop after our predefined lifetime duration
	time.AfterFunc(time.Duration(workerLifeTime)*time.Second, func() {
		lw.Stop()
	})

	// start the push worker
	lw.Start()

	// since the lifetime is 1 second and the push cycle is every 300ms
	// expect 4 elements, one at each timestamp
	// 0ms, 300ms, 600ms, 900ms
	suite.True(len(c.AckMessages) == 4, fmt.Sprintf("Expected 4 messages but got %v", len(c.AckMessages)))
	suite.Equal([]string{"ackid_0", "ackid_1", "ackid_2", "ackid_3"}, c.AckMessages)

	// for each pair of messages check if the time delta between them
	// is greater or equals with the rate that we have defined
	for idx, msg := range c.GeneratedMessages {

		// last element
		if idx == len(c.GeneratedMessages)-1 {
			break
		}

		msgTime, _ := time.Parse(time.StampNano, msg.Msg.PubTime)

		nextMsgTime, _ := time.Parse(time.StampNano, c.GeneratedMessages[idx+1].Msg.PubTime)

		timeDelta := nextMsgTime.Sub(msgTime)

		suite.True(timeDelta >= time.Duration(rate)*time.Millisecond)
	}

	// check that the lifetime was indeed no more than 1 second
	// by comparing the publish time of the first and last messages
	startTime, _ := time.Parse(time.StampNano, c.GeneratedMessages[0].Msg.PubTime)
	endTime, _ := time.Parse(time.StampNano, c.GeneratedMessages[len(c.GeneratedMessages)-1].Msg.PubTime)

	timeDelta := endTime.Sub(startTime)

	suite.True(timeDelta <= time.Duration(workerLifeTime)*time.Second)
}

func (suite *LinearWorkerTestSuite) TestPush() {

	rate := uint32(300)

	ctx, cancel := context.WithCancel(context.TODO())
	sub := &amsPb.Subscription{
		FullName: "sub1",
		PushConfig: &amsPb.PushConfig{
			RetryPolicy: &amsPb.RetryPolicy{
				Period: rate,
			},
		},
	}

	c := new(consumers.MockConsumer)
	c.SubStatus = "normal_sub"
	c.AckStatus = "normal_ack"
	s := new(senders.MockSender)

	lw := LinearWorker{
		sub:      sub,
		consumer: c,
		sender:   s,
		ctx:      ctx,
		cancel:   cancel,
	}

	// no error
	lw.push()
	suite.Equal(1, len(c.AckMessages))
	suite.Equal(1, len(c.GeneratedMessages))
	suite.Equal(1, len(s.PushMessages))
	suite.Equal("Subscription sub1 is currently active", lw.Status())

	// receive consumer error
	// no message available to send
	// no message available to ack
	c.SubStatus = "error_sub"
	c.GeneratedMessages = nil
	c.AckMessages = nil
	s.PushMessages = nil
	lw.push()
	suite.Equal(0, len(c.AckMessages))
	suite.Equal(0, len(c.GeneratedMessages))
	suite.Equal(0, len(s.PushMessages))
	suite.Regexp("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2} - Could not consume message, error while consuming", lw.Status())

	// receive ack error
	c.SubStatus = "normal_sub"
	c.AckStatus = "timeout_ack"
	c.GeneratedMessages = nil
	c.AckMessages = nil
	s.PushMessages = nil
	lw.push()
	suite.Equal(0, len(c.AckMessages))
	suite.Equal(1, len(c.GeneratedMessages))
	suite.Equal(1, len(s.PushMessages))
	suite.Regexp("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2} - Could not acknowledge message, error while acknowledging", lw.Status())

	// send error
	// no message available for ack
	c.AckStatus = "normal_ack"
	s.SendStatus = "error_send"
	c.GeneratedMessages = nil
	c.AckMessages = nil
	s.PushMessages = nil
	lw.push()
	suite.Equal(0, len(c.AckMessages))
	suite.Equal(1, len(c.GeneratedMessages))
	suite.Equal(0, len(s.PushMessages))
	pushErr1stCycle := lw.pushErr
	suite.Regexp("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2} - Could not send message, error while sending", pushErr1stCycle)
	suite.Regexp("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2} - Could not send message, error while sending", lw.Status())

	// consume error and cancel(project not found)
	c.SubStatus = "error_sub_no_project"
	c.GeneratedMessages = nil
	c.AckMessages = nil
	s.PushMessages = nil
	cancelCh := make(chan consumers.CancelableError, 1)
	lw.deactivationChan = cancelCh
	lw.push()
	suite.Equal(0, len(c.AckMessages))
	suite.Equal(0, len(c.GeneratedMessages))
	suite.Equal(0, len(s.PushMessages))
	suite.Equal(consumers.CancelableError{
		ErrMsg:   "project doesn't exist",
		Resource: "error_sub_no_project",
	}, <-cancelCh)

	// consume error and cancel(sub not found)
	c.SubStatus = "error_sub_no_sub"
	c.GeneratedMessages = nil
	c.AckMessages = nil
	s.PushMessages = nil
	cancelCh2 := make(chan consumers.CancelableError, 1)
	lw.deactivationChan = cancelCh2
	lw.push()
	suite.Equal(0, len(c.AckMessages))
	suite.Equal(0, len(c.GeneratedMessages))
	suite.Equal(0, len(s.PushMessages))
	suite.Equal(consumers.CancelableError{
		ErrMsg:   "Subscription doesn't exist",
		Resource: "error_sub_no_sub",
	}, <-cancelCh2)
}

func (suite *LinearWorkerTestSuite) TestConsumer() {

	mc := new(consumers.MockConsumer)
	lw := LinearWorker{
		consumer: mc,
	}

	suite.Equal(mc, lw.consumer)
}

func (suite *LinearWorkerTestSuite) TestStatus() {

	lw := LinearWorker{
		sub: &amsPb.Subscription{
			FullName: "sub1",
		},
	}

	lw.pushErr = "error1"

	suite.Equal("error1", lw.Status())

	lw.pushErr = ""
	suite.Equal("Subscription sub1 is currently active", lw.Status())
}

func TestLinearWorkerTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(LinearWorkerTestSuite))
}
