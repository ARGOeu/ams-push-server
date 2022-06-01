package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	ApplicationJson      = "application/json"
	ProjectNotFound      = "project doesn't exist"
	SubscriptionNotFound = "Subscription doesn't exist"
)

// AmsHttpConsumer is a consumer that helps us interface with AMS through its rest api
type AmsHttpConsumer struct {
	fullSub   string
	amsClient *ams.Client
}

// AmsHttpError represents the layout of an ams http api error
type AmsHttpError struct {
	Error amsErr `json:"error"`
}

// amsErr represents the "model" of an ams http api error
type amsErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewAmsHttpConsumer initialises and returns a new ams http consumer
func NewAmsHttpConsumer(fullSub string, amsClient *ams.Client) *AmsHttpConsumer {
	ahc := new(AmsHttpConsumer)
	ahc.fullSub = fullSub
	ahc.amsClient = amsClient
	return ahc
}

func (ahc *AmsHttpConsumer) ToCancelableError(error error) (CancelableError, bool) {

	// check if the errMsg can be marshaled to an ams http error
	ahe := new(AmsHttpError)
	err := json.Unmarshal([]byte(error.Error()), ahe)
	if err != nil {
		return CancelableError{}, false
	}

	// check if the error is produced from a project or subscription that doesn't exist
	if ahe.Error.Message == ProjectNotFound {
		return NewCancelableError(ProjectNotFound, ahc.fullSub), true
	}

	if ahe.Error.Message == SubscriptionNotFound {
		return NewCancelableError(SubscriptionNotFound, ahc.fullSub), true
	}

	return CancelableError{}, false
}

// ResourceInfo returns the ams subscription and the ams host it is on
func (ahc *AmsHttpConsumer) ResourceInfo() string {
	return fmt.Sprintf("subscription %v from %v", ahc.fullSub, ahc.amsClient.Host())
}

// Consume consumes messages from an subscription
func (ahc *AmsHttpConsumer) Consume(ctx context.Context, numberOfMessages int64) (ams.ReceivedMessagesList, error) {

	log.WithFields(
		log.Fields{
			"type":     "service_log",
			"resource": ahc.ResourceInfo(),
		},
	).Debug("Trying to consume message")

	t1 := time.Now()

	reqList, err := ahc.amsClient.Pull(ctx, ahc.fullSub, numberOfMessages, true)
	if err != nil {
		return ams.ReceivedMessagesList{}, err
	}

	if reqList.IsEmpty() {
		return ams.ReceivedMessagesList{}, errors.New("no new messages")
	}

	log.WithFields(
		log.Fields{
			"type":            "performance_log",
			"message":         reqList,
			"resource":        ahc.ResourceInfo(),
			"processing_time": time.Since(t1).String(),
		},
	).Info("Message consumed")

	return reqList, nil
}

// Ack acknowledges that an ams message has been consumed and processed
func (ahc *AmsHttpConsumer) Ack(ctx context.Context, ackId string) error {
	t1 := time.Now()
	err := ahc.amsClient.Ack(ctx, ahc.fullSub, ackId)
	if err != nil {
		return fmt.Errorf("an error occurred while trying to acknowledge message with ackId %v from %v, %v",
			ackId, ahc.ResourceInfo(), err.Error())
	}
	log.WithFields(
		log.Fields{
			"type":            "performance",
			"ackId":           ackId,
			"resource":        ahc.ResourceInfo(),
			"processing_time": time.Since(t1).String(),
		},
	).Debug("Message acknowledged")

	return nil
}
