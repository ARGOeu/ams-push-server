package consumers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const ApplicationJson = "application/json"

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

// AckMsgs the ack ids for the messages we want to acknowledge
type AckMsgs struct {
	AckIDS []string `json:"ackIds"`
}

// AmsHttpConsumer is a consumer that helps us interface with AMS through its rest api
type AmsHttpConsumer struct {
	client   *http.Client
	endpoint string
	fullSub  string
	token    string
}

// NewAmsHttpConsumer initialises and returns a new ams http consumer
func NewAmsHttpConsumer(endpoint, fullSub, token string, client *http.Client) *AmsHttpConsumer {
	ahc := new(AmsHttpConsumer)
	ahc.client = client
	ahc.endpoint = endpoint
	ahc.fullSub = fullSub
	ahc.token = token
	return ahc
}

// ResourceInfo returns the ams subscription and the ams host it is on
func (ahc *AmsHttpConsumer) ResourceInfo() string {
	return fmt.Sprintf("subscription %v from %v", ahc.fullSub, ahc.endpoint)
}

// Consume consumes messages from an subscription
func (ahc *AmsHttpConsumer) Consume(ctx context.Context) (ReceivedMessagesList, error) {

	url := fmt.Sprintf("https://%v/v1%v:pull?key=%v", ahc.endpoint, ahc.fullSub, ahc.token)

	pullOptions := PullOptions{
		MaxMessages:       "1",
		ReturnImmediately: "true",
	}

	pullOptB, err := json.Marshal(pullOptions)
	if err != nil {
		return ReceivedMessagesList{}, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(pullOptB))
	if err != nil {
		return ReceivedMessagesList{}, err
	}

	req.Header.Set("Content-Type", ApplicationJson)
	log.Debugf("Trying to pull messages for %v", ahc.ResourceInfo())

	t1 := time.Now()

	resp, err := ahc.client.Do(req.WithContext(ctx))
	if err != nil {
		return ReceivedMessagesList{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		err = errors.New(fmt.Sprintf("an error occurred while trying to consume messages from %v, %v", ahc.ResourceInfo(), buf.String()))
		return ReceivedMessagesList{}, err
	}

	reqList := ReceivedMessagesList{}
	err = json.NewDecoder(resp.Body).Decode(&reqList)
	if err != nil {
		return reqList, err
	}

	log.Debugf("Messages %+v from %v consumed in: %v", reqList, ahc.ResourceInfo(), time.Since(t1).String())

	return reqList, nil
}

// Ack acknowledges that an ams message has been consumed and processed
func (ahc *AmsHttpConsumer) Ack(ctx context.Context, ackId string) error {

	ack := AckMsgs{AckIDS: []string{ackId}}

	ackB, err := json.Marshal(ack)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%v/v1%v:acknowledge?key=%v", ahc.endpoint, ahc.fullSub, ahc.token)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(ackB))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", ApplicationJson)

	t1 := time.Now()

	resp, err := ahc.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		err = errors.New(fmt.Sprintf("an error occurred while trying to acknowledge message with ackId %v from %v, %v", ackId, ahc.ResourceInfo(), buf.String()))
		return err
	}

	log.Debugf("Message with ackId %v for %v got acknowledged in %v", ackId, ahc.ResourceInfo(), time.Since(t1).String())

	return nil
}
