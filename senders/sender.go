package senders

import (
	"context"
	"fmt"
	amsPb "github.com/ARGOeu/ams-push-server/api/v1/grpc/proto"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"net/http"
)

type senderType string

type pushMessageFormat string

const (
	HttpSenderType        senderType        = "http-sender"
	MattermostSenderType  senderType        = "mattermost"
	SingleMessageFormat   pushMessageFormat = "single"
	MultipleMessageFormat pushMessageFormat = "multi"
)

// Sender is responsible for delivering data to remote destinations
type Sender interface {
	// Send sends the data to a remote destination
	Send(ctx context.Context, msgs PushMsgs, format pushMessageFormat) error
	// Destination returns the target destination where the sender sends the data
	Destination() string
}

// New acts as a sender factory, creates and returns a new sender based on the provided type
func New(cfg amsPb.PushConfig, client *http.Client) (Sender, error) {

	switch cfg.Type {
	case amsPb.PushType_HTTP_ENDPOINT:
		return NewHttpSender(cfg.PushEndpoint, cfg.AuthorizationHeader, client), nil
	case amsPb.PushType_MATTERMOST:
		return NewMattermostSender(cfg.MattermostUrl, cfg.MattermostUsername, cfg.MattermostChannel, client), nil
	}

	return nil, fmt.Errorf("sender %v not yet implemented", cfg.Type)
}

// PushMsg holds data to be send to a remote endpoint
type PushMsg struct {
	// the actual message
	Msg ams.Message `json:"message"`
	// the source
	Sub string `json:"subscription"`
}

// PushMsgs holds data to be send to a remote endpoint(multiple messages grouped together)
type PushMsgs struct {
	// the actual messages
	Messages []PushMsg `json:"messages"`
}

// DetermineMessageFormat decides what message format should be used depending on the number of messages
func DetermineMessageFormat(numberOfMessages int64) pushMessageFormat {

	var f pushMessageFormat

	if numberOfMessages == 1 {
		f = SingleMessageFormat
	} else {
		f = MultipleMessageFormat
	}

	return f
}
