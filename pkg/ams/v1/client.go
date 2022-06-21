package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Client encapsulates all the possible api calls that a client can use to interface with the ams service
type Client struct {
	*UserService
	*SubscriptionService
	*MessageService
}

// AmsBaseInfo holds basic information to interact with the ams service
type AmsBaseInfo struct {
	Scheme  string
	Host    string
	Headers map[string]string
}

// NewClient properly initialises a new ams client
func NewClient(scheme, host, token string, port int, client *http.Client) *Client {
	baseInfo := AmsBaseInfo{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%v", host, port),
		Headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    token,
		},
	}

	c := new(Client)
	c.UserService = NewUserService(baseInfo, client)
	c.SubscriptionService = NewSubscriptionService(baseInfo, client)
	c.MessageService = NewMessageService(baseInfo, client)
	return c
}

func (c *Client) Host() string {
	return c.UserService.Host
}

// AmsRequest contains the necessary data for an ams request to be executed
type AmsRequest struct {
	ctx     context.Context
	method  string
	url     string
	body    interface{}
	headers map[string]string
	*http.Client
}

func (a *AmsRequest) execute() (*http.Response, error) {

	req, err := http.NewRequestWithContext(a.ctx, a.method, a.url, a.marshalRequestBody())
	if err != nil {
		return &http.Response{}, err
	}

	// headers
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := bytes.Buffer{}
		_, _ = buf.ReadFrom(resp.Body)
		_ = resp.Body.Close()
		return &http.Response{}, errors.New(buf.String())
	}

	return resp, nil
}

func (a *AmsRequest) marshalRequestBody() io.Reader {
	if a.body == nil {
		return nil
	}
	b, _ := json.Marshal(a.body)
	return bytes.NewReader(b)
}
