package retrypolicies

import (
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/pkg/errors"
	"time"
)

// RetryPolicyType represents the available retry policies that the service supports
type RetryPolicyType string

const (
	LinearRetryPolicy    = "linear"
	SlowStartRetryPolicy = "slowstart"
)

// RetryPolicy provides a worker the time intervals it will need in order to perform the push cycle
type RetryPolicy interface {
	// Reset resets the Timer in order to provide the next time event
	// Some retry policies might take into consideration the provided error in order to compute the next time event
	Reset(err string)
	// Timer returns the timer used by the respective retry policy
	Timer() *time.Timer
}

// New transforms the registered retry policy of a subscription
// to an the internal corresponding one that will be used by the respective worker
func New(rp *amsPb.RetryPolicy) (RetryPolicy, error) {

	switch rp.Type {

	case LinearRetryPolicy:

		return &Linear{
			period: time.Duration(rp.Period) * time.Millisecond,
			timer:  time.NewTimer(0),
		}, nil

	case SlowStartRetryPolicy:

		return &Slowstart{
			timer: time.NewTimer(SlowStartInitialInterval),
			previousRestartInterval: SlowStartInitialInterval,
			previousError:           false,
		}, nil

	}

	return nil, errors.New("not implemented")
}
