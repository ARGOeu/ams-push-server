package consumers

import (
	"context"
	"fmt"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
)

type consumerType string

const (
	AmsHttpConsumerType consumerType = "ams-http-consumer"
)

// Consumer is used to consume data from a source.
type Consumer interface {
	// Consume pulls data from the source
	Consume(ctx context.Context, numberOfMessages int64) (ams.ReceivedMessagesList, error)
	// Ack acknowledges that a data have been successfully pulled and send
	Ack(ctx context.Context, ackId string) error
	// ResourceInfo returns returns a string representation of the data source
	ResourceInfo() string
	// ToCancelableError checks whether or not an error represents a cancelable error
	// for the respective consumer, if it does, it formats it to a cancelable error
	ToCancelableError(error error) (CancelableError, bool)
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
func New(cType consumerType, fullSub string, amsClient *ams.Client) (Consumer, error) {

	switch cType {
	case AmsHttpConsumerType:
		return NewAmsHttpConsumer(fullSub, amsClient), nil
	}

	return nil, fmt.Errorf("consumer %v not yet implemented", cType)
}
