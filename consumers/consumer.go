package consumers

import (
	"context"
	"fmt"
	"github.com/ARGOeu/ams-push-server/config"
	"net/http"
)

type consumerType string

const (
	AmsHttpConsumerType consumerType = "ams-http-consumer"
)

// Consumer is used to consume data from a source.
type Consumer interface {
	// Consume pulls data from the source
	Consume(ctx context.Context, numberOfMessages int64) (ReceivedMessagesList, error)
	// Ack acknowledges that a data have been successfully pulled and send
	Ack(ctx context.Context, ackId string) error
	// ResourceInfo returns returns a string representation of the data source
	ResourceInfo() string
	// ToCancelableError checks whether or not an error represents a cancelable error
	// for the respective consumer, if it does, it formats it to a cancelable error
	ToCancelableError(error error) (CancelableError, bool)
	// UpdateResourceStatus updates the status of the resource
	UpdateResourceStatus(ctx context.Context, status string) error
}

// There are specific errors that if they are faced they indicate that the consumption should stop
// CancelableError is used as a special error form that indicates that the push worker should stop
// its functionality in case it occurs
type CancelableError struct {
	// string representation of the occurred error
	ErrMsg string
	// the resource that the error relates to
	Resource string
}

func NewCancelableError(errMSg string, resource string) CancelableError {
	return CancelableError{
		ErrMsg:   errMSg,
		Resource: resource,
	}
}

// New acts as consumer factory, creates and returns a new consumer based on the provided type
func New(cType consumerType, fullSub string, cfg *config.Config, client *http.Client) (Consumer, error) {

	switch cType {
	case AmsHttpConsumerType:

		amsEndpoint := fmt.Sprintf("%v:%v", cfg.AmsHost, cfg.AmsPort)

		return NewAmsHttpConsumer(amsEndpoint, fullSub, cfg.AmsToken, client), nil
	}

	return nil, fmt.Errorf("consumer %v not yet implemented", cType)
}
