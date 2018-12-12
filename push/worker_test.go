package push

import (
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
	"github.com/ARGOeu/ams-push-server/senders"
	"github.com/stretchr/testify/suite"
	"testing"
)

type WorkerTestSuite struct {
	suite.Suite
}

// TestNew tests that the worker factory behaves properly
func (suite *WorkerTestSuite) TestNew() {

	s1 := &amsPb.Subscription{
		PushConfig: &amsPb.PushConfig{
			RetryPolicy: &amsPb.RetryPolicy{
				Type: "linear",
			},
		},
	}

	// normal creation
	w1, err1 := New(s1, &consumers.MockConsumer{}, &senders.MockSender{})
	suite.IsType(&LinearWorker{}, w1)
	suite.Nil(err1)

	// unimplemented worker type
	s1.PushConfig.RetryPolicy.Type = "unknown"
	w2, err2 := New(s1, nil, nil)
	suite.Equal("worker unknown not yet implemented", err2.Error())
	suite.Nil(w2)
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
