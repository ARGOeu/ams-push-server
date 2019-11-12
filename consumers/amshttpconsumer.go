package consumers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

const (
	ApplicationJson      = "application/json"
	ProjectNotFound      = "project doesn't exist"
	SubscriptionNotFound = "Subscription doesn't exist"
)

type PushStatus struct {
	PushStatus string `json:"push_status"`
}

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

// AmsHttpConsumer is a consumer that helps us interface with AMS through its rest api
type AmsHttpConsumer struct {
	client   *http.Client
	endpoint string
	fullSub  string
	token    string
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
func NewAmsHttpConsumer(endpoint, fullSub, token string, client *http.Client) *AmsHttpConsumer {
	ahc := new(AmsHttpConsumer)
	ahc.client = client
	ahc.endpoint = endpoint
	ahc.fullSub = fullSub
	ahc.token = token
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
	return fmt.Sprintf("subscription %v from %v", ahc.fullSub, ahc.endpoint)
}

// Consume consumes messages from an subscription
func (ahc *AmsHttpConsumer) Consume(ctx context.Context, numberOfMessages int64) (ReceivedMessagesList, error) {

	url := fmt.Sprintf("https://%v/v1%v:pull?key=%v", ahc.endpoint, ahc.fullSub, ahc.token)

	pullOptions := PullOptions{
		MaxMessages:       strconv.FormatInt(numberOfMessages, 10),
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
	log.WithFields(
		log.Fields{
			"type":     "service_log",
			"resource": ahc.ResourceInfo(),
		},
	).Debug("Trying to consume message")

	t1 := time.Now()

	resp, err := ahc.client.Do(req.WithContext(ctx))
	if err != nil {
		return ReceivedMessagesList{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		return ReceivedMessagesList{}, errors.New(buf.String())
	}

	reqList := ReceivedMessagesList{}
	err = json.NewDecoder(resp.Body).Decode(&reqList)
	if err != nil {
		return reqList, err
	}

	if reqList.IsEmpty() {
		return ReceivedMessagesList{}, errors.New("no new messages")
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
		err = fmt.Errorf("an error occurred while trying to acknowledge message with ackId %v from %v, %v", ackId, ahc.ResourceInfo(), buf.String())
		return err
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

// UpdateResourceStatus updates the subscription's status
func (ahc *AmsHttpConsumer) UpdateResourceStatus(ctx context.Context, status string) error {

	pushStatus := PushStatus{
		PushStatus: status,
	}

	pushB, err := json.Marshal(pushStatus)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%v/v1%v:modifyPushStatus?key=%v", ahc.endpoint, ahc.fullSub, ahc.token)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(pushB))
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
		return errors.New(buf.String())
	}

	log.WithFields(
		log.Fields{
			"type":            "performance",
			"status":          status,
			"resource":        ahc.ResourceInfo(),
			"processing_time": time.Since(t1).String(),
		},
	).Info("Resource status updated")

	return nil
}
