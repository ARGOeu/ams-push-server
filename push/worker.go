package push

import (
	"context"
	"encoding/base64"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	"github.com/ARGOeu/ams-push-server/consumers"
	v1 "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/ARGOeu/ams-push-server/retrypolicies"
	"github.com/ARGOeu/ams-push-server/senders"
	log "github.com/sirupsen/logrus"
	"time"
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
	// Status returns the status of the worker
	Status() string
}

// New acts as a worker factory, creates and returns a new worker based on the provided type
func New(sub *amsPb.Subscription, c consumers.Consumer, s senders.Sender, ch chan<- consumers.CancelableError) (Worker, error) {

	rp, err := retrypolicies.New(sub.PushConfig.RetryPolicy)
	if err != nil {
		return nil, fmt.Errorf("worker %v not yet implemented", sub.PushConfig.RetryPolicy.Type)

	}

	w := new(worker)

	parentCtx := context.TODO()
	ctx, cancel := context.WithCancel(parentCtx)

	w.sub = sub
	w.consumer = c
	w.sender = s
	w.retryPolicy = rp
	w.ctx = ctx
	w.cancel = cancel
	w.deactivationChan = ch

	return w, nil

}

// worker implements the Worker interface
type worker struct {
	sub              *amsPb.Subscription
	consumer         consumers.Consumer
	sender           senders.Sender
	cancel           context.CancelFunc
	ctx              context.Context
	retryPolicy      retrypolicies.RetryPolicy
	deactivationChan chan<- consumers.CancelableError
	pushErr          string
}

// Consumer returns the currently in use consumer
func (w *worker) Consumer() consumers.Consumer {
	return w.consumer
}

// Status returns whether or not the worker is experiencing any error handling its assigned subscription
func (w *worker) Status() string {
	if w.pushErr == "" {
		return fmt.Sprintf("Subscription %v is currently active", w.sub.FullName)
	}

	return w.pushErr
}

// Subscription returns the currently active subscription inside the worker
func (w *worker) Subscription() *amsPb.Subscription {
	return w.sub
}

// Start starts the push functionality for the worker
func (w *worker) Start() {

Loop:
	for {
		select {
		case <-w.retryPolicy.Timer().C:
			w.push()
		case <-w.ctx.Done():
			canceled := w.retryPolicy.Timer().Stop()

			if !canceled {
				<-w.retryPolicy.Timer().C
			}

			break Loop
		}

		w.retryPolicy.Reset(w.pushErr)
	}
}

// push executes the push cycle of consume -> send -> ack
func (w *worker) push() {

	rml, err := w.consumer.Consume(w.ctx, w.sub.PushConfig.MaxMessages)
	if err != nil {

		ce, ok := w.consumer.ToCancelableError(err)
		if ok {
			w.deactivationChan <- ce
			return
		}

		if err.Error() == "no new messages" {
			log.WithFields(
				log.Fields{
					"type":     "service_log",
					"resource": w.consumer.ResourceInfo(),
				},
			).Debug("No new messages")
			return
		}

		log.WithFields(
			log.Fields{
				"type":     "service_log",
				"resource": w.consumer.ResourceInfo(),
				"error":    err.Error(),
			},
		).Error("Could not consume message")

		w.pushErr = fmt.Sprintf(
			"%v - %v, %v",
			time.Now().UTC().Format("2006-01-02T15:04:05"),
			"Could not consume message",
			err.Error(),
		)

		return
	}

	pms := senders.PushMsgs{}

	for _, rm := range rml.RecMsgs {

		msgData := ""
		// try to decode base64 payload of message
		// fallback to original content of the message if it fails
		if w.sub.PushConfig.Base_64Decode {
			decodedMessageBytes, err := base64.StdEncoding.DecodeString(rm.Msg.Data)
			if err != nil {
				log.WithFields(
					log.Fields{
						"type":         "service_log",
						"subscription": w.sub.FullName,
						"message_id":   rm.Msg.ID,
						"error":        err.Error(),
					},
				).Error("Could not decode message")
				msgData = rm.Msg.Data
			} else {
				msgData = string(decodedMessageBytes)
			}
		} else {
			msgData = rm.Msg.Data
		}

		msg := senders.PushMsg{
			Sub: w.consumer.ResourceInfo(),
			Msg: v1.Message{
				ID:      rm.Msg.ID,
				Attr:    rm.Msg.Attr,
				Data:    msgData,
				PubTime: rm.Msg.PubTime,
			},
		}

		pms.Messages = append(pms.Messages, msg)
	}

	err = w.sender.Send(w.ctx, pms, senders.DetermineMessageFormat(w.sub.PushConfig.MaxMessages))
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":     "service_log",
				"endpoint": w.sender.Destination(),
				"error":    err.Error(),
			},
		).Error("Could not send message")

		w.pushErr = fmt.Sprintf(
			"%v - %v, %v",
			time.Now().UTC().Format("2006-01-02T15:04:05"),
			"Could not send message",
			err.Error(),
		)

		return
	}

	err = w.consumer.Ack(w.ctx, rml.Last().AckID)
	if err != nil {

		log.WithFields(
			log.Fields{
				"type":  "service_log",
				"error": err.Error(),
			},
		).Error("Could not acknowledge message")

		w.pushErr = fmt.Sprintf(
			"%v - %v, %v",
			time.Now().UTC().Format("2006-01-02T15:04:05"),
			"Could not acknowledge message",
			err.Error(),
		)

		return
	}

	// if no errors occurred during the push cycle make sure that there is no error registered
	w.pushErr = ""
}

// Stop stops the push worker's functionality
func (w *worker) Stop() {
	w.cancel()
}
