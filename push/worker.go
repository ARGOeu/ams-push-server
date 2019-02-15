package push

import (
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
	"github.com/ARGOeu/ams-push-server/senders"
)

type PushError struct {
	ErrMsg  string
	SubName string
}

type WorkerStatus string

// Worker encapsulates the flow of consuming, sending and acknowledging
type Worker interface {
	// Start starts the the push functionality based on the type of the worker
	Start()
	// Stop cancels the push functionality
	Stop()
	// Subscription returns the currently active subscription that is being handled by the worker
	Subscription() *amsPb.Subscription
	// Consumer returns the consumer that the worker is using
	Consumer() consumers.Consumer
}

// New acts as a worker factory, creates and returns a new worker based on the provided type
func New(sub *amsPb.Subscription, c consumers.Consumer, s senders.Sender, ch chan<- consumers.CancelableError) (Worker, error) {

	switch sub.PushConfig.RetryPolicy.Type {
	case "linear":
		return NewLinearWorker(sub, c, s, ch), nil
	}

	return nil, fmt.Errorf("worker %v not yet implemented", sub.PushConfig.RetryPolicy.Type)
}
