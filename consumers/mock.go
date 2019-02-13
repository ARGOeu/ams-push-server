package consumers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type MockConsumer struct {
	GeneratedMessages     []ReceivedMessage
	AckMessages           []string
	SubStatus             string
	AckStatus             string
	UpdStatus             string
	UpdatedStatusMessages []string
}

func (m *MockConsumer) UpdateResourceStatus(ctx context.Context, status string) error {

	switch m.UpdStatus {

	case "normal_upd":
		m.UpdatedStatusMessages = append(m.UpdatedStatusMessages, status)
		return nil

	case "error_upd":
		err := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`
		return errors.New(err)
	}

	return nil
}

func (m *MockConsumer) ToCancelableError(error error) (CancelableError, bool) {

	// check if the errMsg can be marshaled to an ams http error
	ahe := new(AmsHttpError)
	err := json.Unmarshal([]byte(error.Error()), ahe)
	if err != nil {
		return CancelableError{}, false
	}

	// check if the error is produced from a project or subscription that doesn't exist
	if ahe.Error.Message == ProjectNotFound {
		return NewCancelableError(ProjectNotFound, m.SubStatus), true
	}

	if ahe.Error.Message == SubscriptionNotFound {
		return NewCancelableError(SubscriptionNotFound, m.SubStatus), true
	}

	return CancelableError{}, false
}

func (m *MockConsumer) Consume(ctx context.Context) (ReceivedMessagesList, error) {

	switch m.SubStatus {

	case "normal_sub":

		rm := ReceivedMessage{
			AckID: fmt.Sprintf("ackid_%v", len(m.GeneratedMessages)),
			Msg: Message{
				Data:    "some data",
				ID:      fmt.Sprintf("id_%v", len(m.GeneratedMessages)),
				PubTime: time.Now().UTC().Format(time.StampNano),
			},
		}

		rml := ReceivedMessagesList{RecMsgs: []ReceivedMessage{rm}}
		m.GeneratedMessages = append(m.GeneratedMessages, rm)
		return rml, nil

	case "empty_sub":

		rml := ReceivedMessagesList{
			RecMsgs: make([]ReceivedMessage, 0),
		}

		return rml, nil

	case "error_sub":

		return ReceivedMessagesList{}, errors.New("error while consuming")

	case "error_sub_no_project":
		err := `{
		 "error": {
			"code": 404,
			"message": "project doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`
		return ReceivedMessagesList{}, errors.New(err)

	case "error_sub_no_sub":
		err := `{
		 "error": {
			"code": 404,
			"message": "Subscription doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`
		return ReceivedMessagesList{}, errors.New(err)

	}

	return ReceivedMessagesList{}, nil
}

func (m *MockConsumer) Ack(ctx context.Context, ackId string) error {

	switch m.AckStatus {

	case "normal_ack":

		m.AckMessages = append(m.AckMessages, ackId)
		return nil

	case "timeout_ack":

		return errors.New("error while acknowledging")
	}

	return nil
}

func (m *MockConsumer) ResourceInfo() string {
	return "mock-consumer"
}

type MockConsumeRoundTripper struct{}

func (m *MockConsumeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", ApplicationJson)

	switch r.URL.Path {

	case "/v1/normal_sub:pull":

		rm := ReceivedMessage{
			AckID: "some_ack_id",
			Msg: Message{
				ID:   "some_id",
				Data: "some_data",
			},
		}

		rml := ReceivedMessagesList{
			RecMsgs: []ReceivedMessage{rm},
		}

		b, _ := json.Marshal(rml)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewReader(b)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/empty_sub:pull":

		rml := ReceivedMessagesList{
			RecMsgs: make([]ReceivedMessage, 0),
		}

		b, _ := json.Marshal(rml)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewReader(b)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/error_sub:pull":

		err := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`

		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/error_sub_no_project:pull":

		err := `{
		 "error": {
			"code": 404,
			"message": "project doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`

		resp = &http.Response{
			StatusCode: 404,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/error_sub_no_sub:pull":

		err := `{
		 "error": {
			"code": 404,
			"message": "Subscription doesn't exist",
			"status": "NOT_FOUND"
		 } 
		}`

		resp = &http.Response{
			StatusCode: 404,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/normal_sub:acknowledge":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(`{}`)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/timeout_sub:acknowledge":

		err := `{
		 "error": {
			"code": 408,
			"message": "ack timeout",
			"status": "TIMEOUT"
		 }
		}`

		resp = &http.Response{
			StatusCode: 408,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/normal_sub:modifyPushStatus":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/error_sub:modifyPushStatus":

		err := `{
		 "error": {
			"code": 500,
			"message": "Internal error",
			"status": "INTERNAL_ERROR"
		 }
		}`

		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	}

	return resp, nil
}
