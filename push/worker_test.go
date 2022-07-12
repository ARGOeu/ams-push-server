package push

import (
	"context"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/ARGOeu/ams-push-server/retrypolicies"
	"github.com/ARGOeu/ams-push-server/senders"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

type WorkerTestSuite struct {
	suite.Suite
}

// TestNew tests that the worker factory behaves properly
func (suite *WorkerTestSuite) TestNew() {

	c := consumers.NewAmsHttpConsumer("", &ams.Client{})
	s := senders.NewHttpSender("", "", &http.Client{})
	sub := &amsPb.Subscription{
		PushConfig: &amsPb.PushConfig{
			Type:        amsPb.PushType_HTTP_ENDPOINT,
			MaxMessages: 1,
			RetryPolicy: &amsPb.RetryPolicy{
				Period: 300,
				Type:   retrypolicies.LinearRetryPolicy,
			},
		},
	}

	// normal creation

	w, err1 := New(sub, c, s, make(chan consumers.CancelableError))

	w1 := w.(*worker)
	suite.Equal(sub, w1.sub)
	suite.Equal(c, w1.consumer)
	suite.Equal(s, w1.sender)
	suite.IsType(&retrypolicies.Linear{}, w1.retryPolicy)
	suite.NotNil(w1.cancel)
	suite.NotNil(w1.ctx)
	suite.IsType(&worker{}, w1)
	suite.Nil(err1)

	// unimplemented worker type
	sub.PushConfig.RetryPolicy.Type = "unknown"
	w2, err2 := New(sub, nil, nil, nil)
	suite.Equal("worker unknown not yet implemented", err2.Error())
	suite.Nil(w2)
}

// TestStartStopCycle checks the correct functionality of starting and stopping a push worker
func (suite *WorkerTestSuite) TestStartStopCycle() {

	// how often does a push cycle occurs in milliseconds
	// a full push cycle: consume -> send -> ack
	rate := uint32(300)

	// for how long the push worker will perform in seconds
	workerLifeTime := 1

	ctx, cancel := context.WithCancel(context.TODO())
	sub := &amsPb.Subscription{
		PushConfig: &amsPb.PushConfig{
			Type:        amsPb.PushType_HTTP_ENDPOINT,
			MaxMessages: 1,
			RetryPolicy: &amsPb.RetryPolicy{
				Period: rate,
				Type:   retrypolicies.LinearRetryPolicy,
			},
		},
	}

	rp, _ := retrypolicies.New(sub.PushConfig.RetryPolicy)

	c := new(consumers.MockConsumer)
	c.SubStatus = "normal_sub"
	c.AckStatus = "normal_ack"
	s := new(senders.MockSender)

	lw := worker{
		sub:         sub,
		consumer:    c,
		sender:      s,
		retryPolicy: rp,
		ctx:         ctx,
		cancel:      cancel,
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

func (suite *WorkerTestSuite) TestPush() {

	rate := uint32(300)

	ctx, cancel := context.WithCancel(context.TODO())
	sub := &amsPb.Subscription{
		FullName: "sub1",
		PushConfig: &amsPb.PushConfig{
			Type:        amsPb.PushType_HTTP_ENDPOINT,
			MaxMessages: 1,
			RetryPolicy: &amsPb.RetryPolicy{
				Period: rate,
				Type:   retrypolicies.LinearRetryPolicy,
			},
		},
	}

	rp, _ := retrypolicies.New(sub.PushConfig.RetryPolicy)

	c := new(consumers.MockConsumer)
	c.SubStatus = "normal_sub"
	c.AckStatus = "normal_ack"
	s := new(senders.MockSender)

	lw := worker{
		sub:         sub,
		consumer:    c,
		sender:      s,
		retryPolicy: rp,
		ctx:         ctx,
		cancel:      cancel,
	}

	// no error - single message - without decode
	lw.push()
	suite.Equal(1, len(c.AckMessages))
	suite.Equal(1, len(c.GeneratedMessages))
	suite.Equal("c29tZSBkYXRh", s.PushMessages[0].Msg.Data)
	suite.Equal(1, len(s.PushMessages))
	suite.Equal("Subscription sub1 is currently active", lw.Status())

	//  with decode
	sub.PushConfig.Base_64Decode = true
	cD := new(consumers.MockConsumer)
	cD.SubStatus = "normal_sub"
	cD.AckStatus = "normal_ack"
	sD := new(senders.MockSender)
	lwDecode := worker{
		sub:         sub,
		consumer:    cD,
		sender:      sD,
		retryPolicy: rp,
		ctx:         ctx,
		cancel:      cancel,
	}
	lwDecode.push()
	suite.Equal(1, len(cD.AckMessages))
	suite.Equal(1, len(cD.GeneratedMessages))
	suite.Equal("some data", sD.PushMessages[0].Msg.Data)
	suite.Equal(1, len(sD.PushMessages))
	suite.Equal("Subscription sub1 is currently active", lwDecode.Status())

	sub2 := &amsPb.Subscription{
		FullName: "sub1",
		PushConfig: &amsPb.PushConfig{
			Type:        amsPb.PushType_HTTP_ENDPOINT,
			MaxMessages: 3,
			RetryPolicy: &amsPb.RetryPolicy{
				Period: rate,
				Type:   retrypolicies.LinearRetryPolicy,
			},
		},
	}

	rp2, _ := retrypolicies.New(sub2.PushConfig.RetryPolicy)

	c2 := new(consumers.MockConsumer)
	c2.SubStatus = "normal_sub"
	c2.AckStatus = "normal_ack"
	s2 := new(senders.MockSender)

	lw2 := worker{
		sub:         sub2,
		consumer:    c2,
		sender:      s2,
		retryPolicy: rp2,
		ctx:         ctx,
		cancel:      cancel,
	}

	// no error - multiple messages
	lw2.push()
	suite.Equal(1, len(c2.AckMessages))
	suite.Equal(3, len(c2.GeneratedMessages))
	suite.Equal(3, len(s2.PushMessages))
	suite.Equal("Subscription sub1 is currently active", lw2.Status())

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

func (suite *WorkerTestSuite) TestConsumer() {

	mc := new(consumers.MockConsumer)
	lw := worker{
		consumer: mc,
	}

	suite.Equal(mc, lw.consumer)
}

func (suite *WorkerTestSuite) TestStatus() {

	lw := worker{
		sub: &amsPb.Subscription{
			FullName: "sub1",
		},
	}

	lw.pushErr = "error1"

	suite.Equal("error1", lw.Status())

	lw.pushErr = ""
	suite.Equal("Subscription sub1 is currently active", lw.Status())
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
