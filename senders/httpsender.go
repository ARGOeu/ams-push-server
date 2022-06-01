package senders

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const ApplicationJson = "application/json"

// HttpSender delivers data to any http endpoint
type HttpSender struct {
	client      *http.Client
	endpoint    string
	authZHeader string
}

// NewHttpSender initialises and returns a new http sender
func NewHttpSender(endpoint, authz string, client *http.Client) *HttpSender {
	s := new(HttpSender)
	s.client = client
	s.endpoint = endpoint
	s.authZHeader = authz
	return s
}

// Send delivers a message to remote http endpoint
func (s *HttpSender) Send(ctx context.Context, msgs PushMsgs, format pushMessageFormat) error {

	var msgB []byte
	var err error

	if format == SingleMessageFormat {
		msgB, err = json.Marshal(msgs.Messages[0])
		if err != nil {
			return err
		}
	} else if format == MultipleMessageFormat {
		msgB, err = json.Marshal(msgs)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewBuffer(msgB))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", ApplicationJson)
	if s.authZHeader != "" {
		req.Header.Set("Authorization", s.authZHeader)
	}

	log.WithFields(
		log.Fields{
			"type":        "service_log",
			"message(s)":  msgs,
			"destination": s.endpoint,
		},
	).Debug("Trying to send")

	t1 := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusProcessing {
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		return errors.New(buf.String())
	}

	log.WithFields(
		log.Fields{
			"type":            "performance_log",
			"message(s)":      msgs,
			"endpoint":        s.endpoint,
			"processing_time": time.Since(t1).String(),
		},
	).Info("Delivered successfully")

	return nil
}

// Destination returns the http endpoint where data is being sent
func (s *HttpSender) Destination() string {
	return s.endpoint
}
