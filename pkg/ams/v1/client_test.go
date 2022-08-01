package v1

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type ClientTestSuite struct {
	suite.Suite
}

func (suite *ClientTestSuite) TestNewAmsClient() {
	client := new(http.Client)
	amsClient := NewClient("https", "localhost", "token", 443, client)

	suite.NotNil(amsClient.UserService)
	suite.NotNil(amsClient.SubscriptionService)

	// check the proper initialization of the user service
	suite.NotNil(amsClient.UserService.client)
	suite.Equal("https", amsClient.UserService.Scheme)
	suite.Equal("localhost:443", amsClient.UserService.Host)
	suite.Equal(map[string]string{
		"Content-type": "application/json",
		"x-api-key":    "token",
	}, amsClient.UserService.Headers)

	// check the proper initialization of the subscription service
	suite.NotNil(amsClient.SubscriptionService.client)
	suite.Equal("https", amsClient.SubscriptionService.Scheme)
	suite.Equal("localhost:443", amsClient.SubscriptionService.Host)
	suite.Equal(map[string]string{
		"Content-type": "application/json",
		"x-api-key":    "token",
	}, amsClient.SubscriptionService.Headers)

	// check the proper initialization of the message service
	suite.NotNil(amsClient.MessageService.client)
	suite.Equal("https", amsClient.MessageService.Scheme)
	suite.Equal("localhost:443", amsClient.MessageService.Host)
	suite.Equal(map[string]string{
		"Content-type": "application/json",
		"x-api-key":    "token",
	}, amsClient.MessageService.Headers)
}

func (suite *ClientTestSuite) TestHost() {
	amsClient := NewClient("https", "localhost", "token", 443, new(http.Client))
	suite.Equal("localhost:443", amsClient.Host())
}

func (suite *ClientTestSuite) TestAmsRequestExecute() {

	mrt := new(mockRoundTripper)

	client := &http.Client{
		Transport: mrt,
	}

	rBody := mockStruct{
		FieldOne: "one",
		FieldTwo: 22,
	}

	amsRequest := AmsRequest{
		ctx:    context.Background(),
		method: http.MethodGet,
		url:    "https://localhost:443/api",
		body:   rBody,
		headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    "token",
		},
		Client: client,
	}

	amsRequest100 := AmsRequest{
		ctx:    context.Background(),
		method: http.MethodGet,
		url:    "https://localhost:443/api-100",
		body:   rBody,
		headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    "token",
		},
		Client: client,
	}

	amsRequest300 := AmsRequest{
		ctx:    context.Background(),
		method: http.MethodGet,
		url:    "https://localhost:443/api-300",
		body:   rBody,
		headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    "token",
		},
		Client: client,
	}

	// check successful response and execution
	resp, err := amsRequest.execute()
	suite.NotNil(resp)
	suite.Nil(err)
	// check that the body has been marshaled properly and headers assigned properly
	b, _ := json.Marshal(rBody)
	suite.Equal(b, mrt.requestBody)
	suite.Equal("token", mrt.header.Get("x-api-key"))
	suite.Equal("application/json", mrt.header.Get("Content-type"))
	// status code < 200
	_, err100 := amsRequest100.execute()
	suite.Equal("error-100", err100.Error())
	// status code > 300
	_, err300 := amsRequest300.execute()
	suite.Equal("error-300", err300.Error())
}

func TestClientTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(ClientTestSuite))
}

type mockRoundTripper struct {
	requestBody []byte
	header      http.Header
}

type mockStruct struct {
	FieldOne string `json:"field_one"`
	FieldTwo int    `json:"field_two"`
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var resp *http.Response

	header := make(http.Header)
	header.Set("Content-type", "application/json")

	m.requestBody, _ = ioutil.ReadAll(r.Body)
	m.header = r.Header.Clone()

	switch r.URL.Path {

	case "/api":

		resp = &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("text")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/api-100":

		resp = &http.Response{
			StatusCode: 100,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("error-100")),
			// Must be set to non-nil value or it panics
			Header: header,
		}

	case "/api-300":

		resp = &http.Response{
			StatusCode: 300,
			// Send response to be tested
			Body: ioutil.NopCloser(strings.NewReader("error-300")),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	}

	return resp, nil
}
