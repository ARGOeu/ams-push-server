package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatusInterceptor is used in order to check, depending on the service's status if the call should be continued or not
func StatusInterceptor(srv *PushService) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		// if a request tries to access any other api call rather than the Status call
		// while the service's status is not ok, block the request
		if info.FullMethod != "/PushService/Status" && srv.status != "ok" {
			return nil, status.Error(codes.Internal, ServiceUnavailable)
		}

		return handler(ctx, req)
	}
}
