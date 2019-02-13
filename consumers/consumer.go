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
	Consume(ctx context.Context) (ReceivedMessagesList, error)
	// Ack acknowledges that a data have been successfully pulled and send
	Ack(ctx context.Context, ackId string) error
	// ResourceInfo returns returns a string representation of the data source
	ResourceInfo() string
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