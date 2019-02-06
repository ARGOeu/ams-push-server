package push

import (
	"context"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
	"github.com/ARGOeu/ams-push-server/senders"
	"github.com/stretchr/testify/suite"
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

	lw := NewLinearWorker(sub, c, s)

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
	e := lw.push()
	suite.Nil(e)

	// empty sub should not forward any message to ack
	c.SubStatus = "empty_sub"
	// clear slices
	c.GeneratedMessages = nil
	c.AckMessages = nil
	e1 := lw.push()
	suite.Nil(e1)
	suite.Equal(0, len(c.AckMessages))

	// receive consumer error
	c.SubStatus = "error_sub"
	e2 := lw.push()
	suite.Equal("error while consuming", e2.Error())

	// receive ack error
	c.SubStatus = "normal_sub"
	c.AckStatus = "timeout_ack"
	e3 := lw.push()
	suite.Equal("error while acknowledging", e3.Error())

	// send error
	c.AckStatus = "normal_ack"
	s.SendStatus = "error_send"
	e4 := lw.push()
	suite.Equal("error while sending", e4.Error())
}

func TestLinearWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(LinearWorkerTestSuite))
}
