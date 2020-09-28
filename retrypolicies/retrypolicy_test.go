package retrypolicies

import (
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/stretchr/testify/suite"
	"testing"

	"time"
)

type RetryPolicyTestSuite struct {
	suite.Suite
}

// TestNew tests the functionality of the constructor New for retry policies
func (suite *RetryPolicyTestSuite) TestNew() {

	// normal case of Linear retry policy
	pbR := &amsPb.RetryPolicy{
		Period: 500,
		Type:   LinearRetryPolicy,
	}

	r1, e1 := New(pbR)
	lr1 := r1.(*Linear)

	suite.Equal(time.Duration(500*time.Millisecond), lr1.period)
	suite.Nil(e1)

	// normal case of Slowstart retry policy
	pbR2 := &amsPb.RetryPolicy{
		Type: SlowStartRetryPolicy,
	}

	r2, e2 := New(pbR2)
	lr2 := r2.(*Slowstart)

	suite.Equal(time.Duration(SlowStartInitialInterval), lr2.previousRestartInterval)
	suite.False(lr2.previousError)
	suite.Nil(e2)

	// error case
	pbR.Type = "unknown"
	_, e3 := New(pbR)
	suite.Equal("not implemented", e3.Error())
}

func TestRetryPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(RetryPolicyTestSuite))
}
