package senders

import (
	"context"
	"fmt"
	"github.com/ARGOeu/ams-push-server/consumers"
	"net/http"
)

type senderType string

const (
	HttpSenderType senderType = "http-sender"
)

// Sender is responsible for delivering data to remote destinations
type Sender interface {
	// Send sends the data to a remote destination
	Send(ctx context.Context, msg PushMsg) error
}

// New acts as a sender factory, creates and returns a new sender based on the provided type
func New(sType senderType, endpoint string, client *http.Client) (Sender, error) {

	switch sType {
	case HttpSenderType:
		return NewHttpSender(endpoint, client), nil
	}

	return nil, fmt.Errorf("sender %v not yet implemented", sType)
}

// PushMsg holds data to be send to a remote endpoint
type PushMsg struct {
	// the actual message
	Msg consumers.Message `json:"message"`
	// the source
	Sub string `json:"subscription"`
}
