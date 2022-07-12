package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/pkg/errors"
	"time"
)

type MockConsumer struct {
	GeneratedMessages     []ams.ReceivedMessage
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

func (m *MockConsumer) Consume(ctx context.Context, numberOfMessages int64) (ams.ReceivedMessagesList, error) {

	switch m.SubStatus {

	case "normal_sub":

		rm := ams.ReceivedMessage{
			AckID: fmt.Sprintf("ackid_%v", len(m.GeneratedMessages)),
			Msg: ams.Message{
				Data:    "c29tZSBkYXRh", // 'some data' literal encoded in b64
				ID:      fmt.Sprintf("id_%v", len(m.GeneratedMessages)),
				PubTime: time.Now().UTC().Format(time.StampNano),
			},
		}

		rml := ams.ReceivedMessagesList{RecMsgs: []ams.ReceivedMessage{}}

		for i := 1; i <= int(numberOfMessages); i++ {
			rml.RecMsgs = append(rml.RecMsgs, rm)
			m.GeneratedMessages = append(m.GeneratedMessages, rm)
		}

		return rml, nil

	case "empty_sub":

		rml := ams.ReceivedMessagesList{
			RecMsgs: make([]ams.ReceivedMessage, 0),
		}

		return rml, nil

	case "error_sub":

		return ams.ReceivedMessagesList{}, errors.New("error while consuming")

	case "error_sub_no_project":
		err := `{
		 "error": {
			"code": 404,
			"message": "project doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`
		return ams.ReceivedMessagesList{}, errors.New(err)

	case "error_sub_no_sub":
		err := `{
		 "error": {
			"code": 404,
			"message": "Subscription doesn't exist",
			"status": "NOT_FOUND"
		 }
		}`
		return ams.ReceivedMessagesList{}, errors.New(err)

	}

	return ams.ReceivedMessagesList{}, nil
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
