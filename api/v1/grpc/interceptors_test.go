package grpc

import (
	"context"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type InterceptorsTestSuite struct {
	suite.Suite
}

func (suite *InterceptorsTestSuite) TestStatusInterceptor() {

	s := &PushService{}

	// check that the interceptor is blocking
	// any requests when the status is not ok
	s.status = "not ok"

	interceptor := StatusInterceptor(s)

	r, err := interceptor(
		context.Background(),
		"i1",
		&grpc.UnaryServerInfo{FullMethod: "/PushService/Random"},
		MockUnaryHandler)

	suite.Nil(r)
	suite.Equal(status.Error(codes.Internal, "The push service is currently unable to handle any requests"), err)

	// normal case where status is ok
	s.status = "ok"
	interceptor2 := StatusInterceptor(s)

	r2, err2 := interceptor2(
		context.Background(),
		"i2",
		&grpc.UnaryServerInfo{FullMethod: "/PushService/Random"},
		MockUnaryHandler)

	suite.Equal("i2", r2.(string))
	suite.Nil(err2)

	//  status not ok but the request asks for the status api call
	s.status = "not ok"
	interceptor3 := StatusInterceptor(s)

	r3, err3 := interceptor3(
		context.Background(),
		"i3",
		&grpc.UnaryServerInfo{FullMethod: "/PushService/Status"},
		MockUnaryHandler)

	suite.Equal("i3", r3.(string))
	suite.Nil(err3)

}

func MockUnaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func TestInterceptorsTestSuite(t *testing.T) {
	suite.Run(t, new(InterceptorsTestSuite))
}
