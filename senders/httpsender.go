package senders

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

	log.WithFields(
		log.Fields{
			"type":        "service_log",
			"message":     msg,
			"destination": s.endpoint,
		},
	).Debug("Trying to push message")

	t1 := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":     "service_log",
				"endpoint": s.endpoint,
				"error":    err.Error(),
			},
		).Error("Could not send message")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusProcessing {
		buf := bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		log.WithFields(
			log.Fields{
				"type":     "service_log",
				"endpoint": s.endpoint,
				"error":    buf.String(),
			},
		).Error("Could not send message")
		err = errors.New(fmt.Sprintf("an error occurred while trying to send message to %v, %v", s.endpoint, buf.String()))
		return err
	}

	log.WithFields(
		log.Fields{
			"type":            "performance_log",
			"message":         msg,
			"endpoint":        s.endpoint,
			"processing_time": time.Since(t1).String(),
		},
	).Debug("Message delivered successfully")

	return nil
}
