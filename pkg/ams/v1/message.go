package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// requires the full subscription path
	// .e.g. /projects/project_one/subscriptions/sub_one::pull
	pullMessagePath = "/v1%s:pull"
	// requires the full subscription path
	// .e.g. /projects/project_one/subscriptions/sub_one::acknowledge
	ackMessagePath = "/v1%s:acknowledge"
)

// Attributes is key/value pairs of extra data
type Attributes map[string]string

// PullOptions holds information on how we want to pull messages
type PullOptions struct {
	// amount of messages to be pulled at once
	MaxMessages string `json:"maxMessages"`
	// whether or not it should wait until it gathers
	// the requested amount of maxMessages
	// to be pulled or return with what is available
	ReturnImmediately string `json:"returnImmediately"`
}

// Message struct used to hold message information
type Message struct {
	// message id
	ID string `json:"messageId,omitempty"`
	// used to hold attribute key/value store
	Attr Attributes `json:"attributes,omitempty"`
	// base64 encoded data payload
	Data string `json:"data"`
	// publish time date of message
	PubTime string `json:"publishTime,omitempty"`
}

// ReceivedMessage holds info for a received message
type ReceivedMessage struct {
	// id to be used for acknowledgement
	AckID string `json:"ackId,omitempty"`
	// the message itself
	Msg Message `json:"message"`
}

// ReceivedMessagesList holds the array of the receivedMessages - subscription related
type ReceivedMessagesList struct {
	RecMsgs []ReceivedMessage `json:"receivedMessages"`
}

// IsEmpty returns whether or not a received message list is empty
func (r *ReceivedMessagesList) IsEmpty() bool {
	return len(r.RecMsgs) <= 0
}

// Last returns the last ReceivedMessage of the slice
func (r *ReceivedMessagesList) Last() ReceivedMessage {
	return r.RecMsgs[len(r.RecMsgs)-1]
}

// AckMsgs the ack ids for the messages we want to acknowledge
type AckMsgs struct {
	AckIDS []string `json:"ackIds"`
}

type MessageService struct {
	client *http.Client
	AmsBaseInfo
}

// NewMessageService properly initialises a message service
func NewMessageService(info AmsBaseInfo, client *http.Client) *MessageService {
	return &MessageService{
		AmsBaseInfo: info,
		client:      client,
	}
}

// Pull consumes messages from an subscription.
// Requires the full subscription path
// .e.g. /projects/project_one/subscriptions/sub_one
func (s *MessageService) Pull(ctx context.Context, subscription string, numberOfMessages int64, returnImmediately bool) (ReceivedMessagesList, error) {

	u := url.URL{
		Host:   s.AmsBaseInfo.Host,
		Scheme: s.AmsBaseInfo.Scheme,
		Path:   fmt.Sprintf(pullMessagePath, subscription),
	}

	req := AmsRequest{
		ctx:    ctx,
		method: http.MethodPost,
		url:    u.String(),
		body: PullOptions{
			MaxMessages:       strconv.FormatInt(numberOfMessages, 10),
			ReturnImmediately: strconv.FormatBool(returnImmediately),
		},
		headers: s.AmsBaseInfo.Headers,
		Client:  s.client,
	}

	resp, err := req.execute()
	if err != nil {
		return ReceivedMessagesList{}, err
	}

	defer resp.Body.Close()

	reqList := ReceivedMessagesList{}
	err = json.NewDecoder(resp.Body).Decode(&reqList)
	if err != nil {
		return reqList, err
	}

	return reqList, nil
}

// Ack acknowledges that an ams message has been consumed and processed
// Requires the full subscription path
// .e.g. /projects/project_one/subscriptions/sub_one
func (s *MessageService) Ack(ctx context.Context, subscription string, ackId string) error {

	u := url.URL{
		Host:   s.AmsBaseInfo.Host,
		Scheme: s.AmsBaseInfo.Scheme,
		Path:   fmt.Sprintf(ackMessagePath, subscription),
	}

	req := AmsRequest{
		ctx:     ctx,
		method:  http.MethodPost,
		url:     u.String(),
		body:    AckMsgs{AckIDS: []string{ackId}},
		headers: s.AmsBaseInfo.Headers,
		Client:  s.client,
	}

	resp, err := req.execute()
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
