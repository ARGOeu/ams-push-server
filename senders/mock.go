package senders

import (
	"context"
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
	return "mock desstination"
}

func (s *MockSender) Send(ctx context.Context, msg PushMsg) error {

	switch s.SendStatus {

	case "error_send":
		return errors.New("error while sending")
	}

	s.PushMessages = append(s.PushMessages, msg)

	return nil
}

type MockSenderRoundTripper struct{}

func (m *MockSenderRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", ApplicationJson)

	switch r.URL.Path {

	case "/receive_here_200":
		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("")),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	case "/receive_here_201":
		resp = &http.Response{
			StatusCode: 201,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("")),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	case "/receive_here_204":
		resp = &http.Response{
			StatusCode: 204,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("")),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	case "/receive_here_102":
		resp = &http.Response{
			StatusCode: 102,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("")),
			// Must be set to non-nil value or it panics
			Header: header,
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
