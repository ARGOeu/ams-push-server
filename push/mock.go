package push

import amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"

// MockWorker is to be used as a dummy worker when we want the push actual worker functionality
type MockWorker struct {
	Sub    amsPb.Subscription
	Status string
}

func (w *MockWorker) Subscription() *amsPb.Subscription {
	return new(amsPb.Subscription)
}

func (w *MockWorker) Start() {}

func (w *MockWorker) Stop() { w.Status = "stopped" }
