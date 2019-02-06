package senders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"

	"errors"
)

const ApplicationJson = "application/json"

// HttpSender delivers data to any http endpoint
type HttpSender struct {
	client   *http.Client
	endpoint string
}

// NewHttpSender initialises and returns a new http sender
func NewHttpSender(endpoint string, client *http.Client) *HttpSender {
	s := new(HttpSender)
	s.client = client
	s.endpoint = endpoint
	return s
}

// Send delivers a message to remote http endpoint
func (s *HttpSender) Send(ctx context.Context, msg PushMsg) error {

	msgB, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewBuffer(msgB))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", ApplicationJson)
	log.Infof("Trying to push message: %+v to: %v", msg, s.endpoint)

	t1 := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusProcessing {
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		err = errors.New(fmt.Sprintf("an error occurred while trying to send message to %v, %v", s.endpoint, buf.String()))
		return err
	}

	log.Infof("Message: %+v to endpoint: %v delivered in: %v", msg, s.endpoint, time.Since(t1).String())

	return nil
}
