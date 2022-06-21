package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	getUserByTokenPath = "/v1/users:byToken/%s"
)

type UserInfo struct {
	Name     string    `json:"name"`
	Projects []Project `json:"projects"`
}

type Project struct {
	Project       string   `json:"project"`
	Subscriptions []string `json:"subscriptions"`
}

type UserService struct {
	client *http.Client
	AmsBaseInfo
}

// NewUserService properly initialises a user service
func NewUserService(info AmsBaseInfo, client *http.Client) *UserService {
	return &UserService{
		AmsBaseInfo: info,
		client:      client,
	}
}

// GetUserByToken uses the provided token to get the respective user profile
func (us *UserService) GetUserByToken(ctx context.Context, token string) (UserInfo, error) {

	u := url.URL{
		Host:   us.AmsBaseInfo.Host,
		Scheme: us.AmsBaseInfo.Scheme,
		Path:   fmt.Sprintf(getUserByTokenPath, token),
	}

	req := AmsRequest{
		ctx:     ctx,
		method:  http.MethodGet,
		url:     u.String(),
		body:    nil,
		headers: us.AmsBaseInfo.Headers,
		Client:  us.client,
	}

	resp, err := req.execute()
	if err != nil {
		return UserInfo{}, err
	}

	defer resp.Body.Close()

	userInfo := UserInfo{}

	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return UserInfo{}, err
	}
	return userInfo, nil
}
