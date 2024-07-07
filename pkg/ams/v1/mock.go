package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type MockAmsRoundTripper struct{}

func (m *MockAmsRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", "application/json")

	switch r.URL.Path {

	case "/v1/users:byToken/sometoken":

		p1 := Project{
			Project:       "push1",
			Subscriptions: []string{"sub1", "errorsub"},
		}

		p2 := Project{
			Project:       "push2",
			Subscriptions: []string{"sub3", "sub4", "sub5"},
		}

		userInfo := UserInfo{
			Name:     "worker",
			Projects: []Project{p1, p2},
		}

		b, _ := json.Marshal(userInfo)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewReader(b)),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	case "/v1/users:byToken/errortoken":

		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: io.NopCloser(strings.NewReader("server internal error")),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	case "/v1/projects/push1/subscriptions/sub1":

		rp := RetryPolicy{
			PolicyType: "linear",
			Period:     300,
		}
		authz := AuthorizationHeader{
			Value: "auth-header-1",
		}

		pc := PushConfig{
			Pend:                "example.com:9999",
			Type:                HttpEndpointPushConfig,
			AuthorizationHeader: authz,
			RetPol:              rp,
			Base64Decode:        true,
		}

		s := Subscription{
			FullName:  "/projects/push1/subscriptions/sub1",
			FullTopic: "/projects/push1/topics/t1",
			PushCfg:   pc,
		}

		sb, _ := json.Marshal(s)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewReader(sb)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/projects/push2/subscriptions/sub5":

		rp := RetryPolicy{
			PolicyType: "linear",
			Period:     300,
		}

		pc := PushConfig{
			Type:               MattermostPushConfig,
			RetPol:             rp,
			MattermostChannel:  "channel",
			MattermostUsername: "mattermost",
			MattermostUrl:      "webhook.com",
			Base64Decode:       false,
		}

		s := Subscription{
			FullName:  "/projects/push2/subscriptions/sub5",
			FullTopic: "/projects/push2/topics/t1",
			PushCfg:   pc,
		}

		sb, _ := json.Marshal(s)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewReader(sb)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/projects/push1/subscriptions/errorsub":

		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: io.NopCloser(strings.NewReader("server internal error")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/projects/push2/subscriptions/sub3":

		s := Subscription{
			FullName:  "/projects/push2/subscriptions/sub3",
			FullTopic: "/projects/push2/topics/t1",
		}

		sb, _ := json.Marshal(s)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewReader(sb)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/projects/push2/subscriptions/sub4":

		rp := RetryPolicy{
			PolicyType: "linear",
			Period:     300,
		}

		pc := PushConfig{
			Type:   HttpEndpointPushConfig,
			Pend:   "example.com:9999",
			RetPol: rp,
		}

		s := Subscription{
			FullName:  "/projects/push2/subscriptions/sub4",
			FullTopic: "/projects/push2/topics/t1",
			PushCfg:   pc,
		}

		sb, _ := json.Marshal(s)

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewReader(sb)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	default:
		resp = &http.Response{
			StatusCode: 500,
			// Send response to be tested
			Body: io.NopCloser(strings.NewReader("unexpected outcome")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	}

	return resp, nil
}

type MockConsumeRoundTripper struct {
	RequestBodyBytes []byte
}

func (m *MockConsumeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", "application/json")

	m.RequestBodyBytes, _ = io.ReadAll(r.Body)

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
			Body: io.NopCloser(bytes.NewReader(b)),
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
			Body: io.NopCloser(bytes.NewReader(b)),
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
			Body: io.NopCloser(strings.NewReader(err)),
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
			Body: io.NopCloser(strings.NewReader(err)),
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
			Body: io.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/normal_sub:acknowledge":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(strings.NewReader(`{}`)),
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
			Body: io.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/v1/normal_sub:modifyPushStatus":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(strings.NewReader("")),
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
			Body: io.NopCloser(strings.NewReader(err)),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	}

	return resp, nil
}
