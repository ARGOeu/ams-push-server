package push

import (
	"context"
	"github.com/ARGOeu/ams-push-server/consumers"
	"github.com/ARGOeu/ams-push-server/senders"
	"time"

	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	log "github.com/sirupsen/logrus"
)

// LinearWorker implements a worker that follows a linear strategy when trying to push
type LinearWorker struct {
	sub      *amsPb.Subscription
	consumer consumers.Consumer
	sender   senders.Sender
	cancel   context.CancelFunc
	ctx      context.Context
}

// NewLinearWorker initialises and configures a new linear worker
func NewLinearWorker(sub *amsPb.Subscription, c consumers.Consumer, s senders.Sender) *LinearWorker {
	lw := new(LinearWorker)

	parentCtx := context.TODO()
	ctx, cancel := context.WithCancel(parentCtx)

	lw.sub = sub
	lw.consumer = c
	lw.sender = s
	lw.ctx = ctx
	lw.cancel = cancel

	return lw
}

// Subscription returns the currently active subscription inside the linear worker
func (w *LinearWorker) Subscription() *amsPb.Subscription {
	return w.sub
}

// Start starts the push functionality for the worker
func (w *LinearWorker) Start() {

	timer := time.NewTimer(0)
	rate := time.Duration(w.sub.PushConfig.RetryPolicy.Period)

Loop:
	for {
		select {
		case <-timer.C:

			err := w.push()
			if err != nil {
				log.Error(err.Error())
			}

		case <-w.ctx.Done():

			canceled := timer.Stop()

			if !canceled {
				<-timer.C
			}

			break Loop
		}

		timer.Reset(rate * time.Millisecond)
	}
}

// push executes the push cycle of consume -> send -> ack
func (w *LinearWorker) push() error {

	rml, err := w.consumer.Consume(w.ctx)
	if err != nil {
		return err
	}

	if rml.IsEmpty() {
		log.WithFields(
			log.Fields{
				"type":     "service_log",
				"resource": w.consumer.ResourceInfo(),
			},
		).Debug("No new messages")
		return nil
	}

	pm := senders.PushMsg{
		Sub: w.consumer.ResourceInfo(),
		Msg: rml.RecMsgs[0].Msg,
	}

	err = w.sender.Send(w.ctx, pm)
	if err != nil {
		return err
	}

	err = w.consumer.Ack(w.ctx, rml.RecMsgs[0].AckID)
	if err != nil {
		return err
	}

	return nil
}

// Stop stops the push worker's functionality
func (w *LinearWorker) Stop() {
	w.cancel()
}
