package senders

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type MockSender struct {
	SendStatus   string
	PushMessages []PushMsg
}

func (s *MockSender) Destination() string {
	return "mock destination"
}

func (s *MockSender) Send(ctx context.Context, msgs PushMsgs, format pushMessageFormat) error {

	switch s.SendStatus {

	case "error_send":
		return errors.New("error while sending")
	}

	if format == SingleMessageFormat {
		s.PushMessages = append(s.PushMessages, msgs.Messages[0])
	} else if format == MultipleMessageFormat {
		for _, msg := range msgs.Messages {
			s.PushMessages = append(s.PushMessages, msg)
		}
	}

	return nil
}

type MockSenderRoundTripper struct {
	RequestBodyBytes []byte
}

func (m *MockSenderRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", ApplicationJson)

	m.RequestBodyBytes, _ = ioutil.ReadAll(r.Body)

	switch r.URL.Path {

	case "/receive_here_200":
		if r.Header.Get("authorization") == "auth-header-1" {
			resp = &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(strings.NewReader("")),
				// Must be set to non-nil value or it panics
				Header: header,
			}
		}
	case "/receive_here_201":
		if r.Header.Get("authorization") == "" {
			resp = &http.Response{
				StatusCode: 201,
				// Send response to be tested
				Body: ioutil.NopCloser(strings.NewReader("")),
				// Must be set to non-nil value or it panics
				Header: header,
			}
		}
	case "/receive_here_204":
		if r.Header.Get("authorization") == "auth-header-1" {
			resp = &http.Response{
				StatusCode: 204,
				// Send response to be tested
				Body: ioutil.NopCloser(strings.NewReader("")),
				// Must be set to non-nil value or it panics
				Header: header,
			}
		}
	case "/receive_here_102":
		if r.Header.Get("authorization") == "auth-header-1" {
			resp = &http.Response{
				StatusCode: 102,
				// Send response to be tested
				Body: ioutil.NopCloser(strings.NewReader("")),
				// Must be set to non-nil value or it panics
				Header: header,
			}
		}
	case "/receive_here_error":

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

type MockMattermostRoundTripper struct {
	RequestBodyBytes []byte
	Message          MattermostMessage
}

func (m *MockMattermostRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", ApplicationJson)

	m.RequestBodyBytes, _ = ioutil.ReadAll(r.Body)
	switch r.URL.Path {

	case "/webhook":

		_ = json.Unmarshal(m.RequestBodyBytes, &m.Message)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("ok")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/mattermost-error":
		err := `{
    				"id": "web.incoming_webhook.channel.app_error",
    				"message": "Couldn't find the channel.",
    				"detailed_error": "",
    				"request_id": "3bxwn99zx7dfxdjy5xkoh5xnmo",
    				"status_code": 404
				}`

		resp = &http.Response{
			StatusCode: 404,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/generic-error":
		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("generic-error")),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	}

	return resp, nil
}
