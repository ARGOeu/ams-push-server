package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// requires the full subscription path
	// .e.g. /projects/project_one/subscriptions/sub_one
	getSubscriptionPath = "/v1%s"
)

type Subscription struct {
	FullName   string     `json:"name"`
	FullTopic  string     `json:"topic"`
	PushCfg    PushConfig `json:"pushConfig"`
	PushStatus string     `json:"push_status"`
}

// PushConfig holds optional configuration for push operations
type PushConfig struct {
	Pend                string              `json:"pushEndpoint"`
	AuthorizationHeader AuthorizationHeader `json:"authorization_header"`
	MaxMessages         int64               `json:"maxMessages"`
	RetPol              RetryPolicy         `json:"retryPolicy"`
}

// AuthorizationHeader holds an optional value to be supplied as an Authorization header to push requests
type AuthorizationHeader struct {
	Value string `json:"value"`
}

// RetryPolicy holds information on retry policies
type RetryPolicy struct {
	PolicyType string `json:"type,omitempty"`
	Period     uint32 `json:"period,omitempty"`
}

func (s *Subscription) IsPushEnabled() bool {
	return s.PushCfg != PushConfig{}
}

type SubscriptionService struct {
	client *http.Client
	AmsBaseInfo
}

// NewSubscriptionService properly initialises a subscription service
func NewSubscriptionService(info AmsBaseInfo, client *http.Client) *SubscriptionService {
	return &SubscriptionService{
		AmsBaseInfo: info,
		client:      client,
	}
}

// GetSubscription retrieves the respective subscription info.
// Requires the full subscription path
// .e.g. /projects/project_one/subscriptions/sub_one
func (s *SubscriptionService) GetSubscription(ctx context.Context, subscription string) (Subscription, error) {

	u := url.URL{
		Host:   s.AmsBaseInfo.Host,
		Scheme: s.AmsBaseInfo.Scheme,
		Path:   fmt.Sprintf(getSubscriptionPath, subscription),
	}

	req := AmsRequest{
		ctx:     ctx,
		method:  http.MethodGet,
		url:     u.String(),
		body:    nil,
		headers: s.AmsBaseInfo.Headers,
		Client:  s.client,
	}

	resp, err := req.execute()
	if err != nil {
		return Subscription{}, err
	}

	defer resp.Body.Close()

	sub := Subscription{}

	err = json.NewDecoder(resp.Body).Decode(&sub)
	if err != nil {
		return Subscription{}, err
	}

	return sub, nil
}
