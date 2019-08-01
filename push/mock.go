package push

import (
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
)

// MockWorker is to be used as a dummy worker when we want the push actual worker functionality
type MockWorker struct {
	Sub       amsPb.Subscription
	SubStatus string
	status    string
}

func (w *MockWorker) Status() string {
	return w.SubStatus
}

func (w *MockWorker) Consumer() consumers.Consumer {
	return new(consumers.MockConsumer)
}

func (w *MockWorker) Subscription() *amsPb.Subscription {
	return new(amsPb.Subscription)
}

func (w *MockWorker) Start() {}

func (w *MockWorker) Stop() {
	w.status = "stopped"
	w.SubStatus = "stopped"
}
