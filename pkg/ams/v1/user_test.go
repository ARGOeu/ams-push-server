package v1

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type UserTestSuite struct {
	suite.Suite
}

func (suite *UserTestSuite) TestNewUserService() {

	userService := NewUserService(AmsBaseInfo{
		Scheme: "https",
		Host:   "localhost",
		Headers: map[string]string{
			"Content-type": "application/json",
			"x-api-key":    "token",
		}}, new(http.Client))

	suite.NotNil(userService.client)
	suite.Equal("https", userService.Scheme)
	suite.Equal("localhost", userService.Host)
	suite.Equal(map[string]string{
		"Content-type": "application/json",
		"x-api-key":    "token",
	}, userService.Headers)
}

func (suite *UserTestSuite) TestGetUserByToken() {

	client := &http.Client{
		Transport: new(MockAmsRoundTripper),
	}

	userService := UserService{
		client: client,
		AmsBaseInfo: AmsBaseInfo{
			Scheme: "https",
			Host:   "localhost",
			Headers: map[string]string{
				"Content-type": "application/json",
				"x-api-key":    "sometoken",
			},
		},
	}

	u1, err1 := userService.GetUserByToken(context.TODO(), "sometoken")

	p1 := Project{
		Project:       "push1",
		Subscriptions: []string{"sub1", "errorsub"},
	}

	p2 := Project{
		Project:       "push2",
		Subscriptions: []string{"sub3", "sub4", "sub5"},
	}

	expectedUserInfo := UserInfo{
		Name:     "worker",
		Projects: []Project{p1, p2},
	}

	suite.Equal(expectedUserInfo, u1)
	suite.Nil(err1)

	// error case
	u2, err2 := userService.GetUserByToken(context.TODO(), "errortoken")
	suite.Equal(UserInfo{}, u2)
	suite.Equal("server internal error", err2.Error())
}

func TestUserServiceTestSuite(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	suite.Run(t, new(UserTestSuite))
}
