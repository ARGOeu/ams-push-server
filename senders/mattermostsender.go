package senders

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type MattermostMessage struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
	Channel  string `json:"channel,omitempty"`
}

type MattermostError struct {
	Id            string `json:"id"`
	Message       string `json:"message"`
	DetailedError string `json:"detailed_error"`
	RequestId     string `json:"request_id"`
	StatusCode    int    `json:"status_code"`
}

// HttpSender delivers data to any http endpoint
type MattermostSender struct {
	client     *http.Client
	webhookUrl string
	username   string
	channel    string
}

func (m *MattermostError) Error() string {
	if m.DetailedError != "" {
		return m.Message + "." + m.DetailedError
	}
	return m.Message
}

// NewMattermostSender initialises and returns a new mattermost sender
func NewMattermostSender(webhookUrl, username, channel string, client *http.Client) *MattermostSender {
	s := new(MattermostSender)
	s.client = client
	s.webhookUrl = webhookUrl
	s.username = username
	s.channel = channel
	return s
}

// Send delivers a message to a remote mattermost webhook url
func (s *MattermostSender) Send(ctx context.Context, msgs PushMsgs, format pushMessageFormat) error {

	if len(msgs.Messages) == 0 {
		return errors.New("no message")
	}

	message := MattermostMessage{
		Text:     msgs.Messages[0].Msg.Data,
		Channel:  s.channel,
		Username: s.username,
	}

	msgB, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookUrl, bytes.NewBuffer(msgB))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", ApplicationJson)

	log.WithFields(
		log.Fields{
			"type":        "service_log",
			"text":        msgs,
			"destination": s.webhookUrl,
		},
	).Debug("Trying to send")

	t1 := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		mattermostError := MattermostError{}

		errorB, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		} else {
			// try to parse the error response into mattermost structured error
			err := json.Unmarshal(errorB, &mattermostError)
			if err != nil {
				return errors.New(string(errorB))
			} else {
				log.WithFields(
					log.Fields{
						"type":           "service_log",
						"endpoint":       s.webhookUrl,
						"id":             mattermostError.Id,
						"message":        mattermostError.Message,
						"detailed_error": mattermostError.DetailedError,
						"request_id":     mattermostError.RequestId,
						"status_code":    mattermostError.StatusCode,
					},
				).Error("Could not deliver message to mattermost")
				return &mattermostError
			}
		}
	}

	log.WithFields(
		log.Fields{
			"type":            "performance_log",
			"message(s)":      msgs,
			"endpoint":        s.webhookUrl,
			"processing_time": time.Since(t1).String(),
		},
	).Info("Delivered successfully")

	return nil
}

// Destination returns the http webhook where data is being sent
func (s *MattermostSender) Destination() string {
	return s.webhookUrl
}
