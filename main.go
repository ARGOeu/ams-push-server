package main

import (
	amsgRPC "github.com/ARGOeu/ams-push-server/api/v1/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"log/syslog"
	"net"
	"time"
)

var LOGGER = logrus.New()

func init() {
	LOGGER.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, DisableColors: true})
	hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		LOGGER.AddHook(hook)
	}
}

func main() {

	opts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}

	srvOpts := grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(LOGGER), opts...),
	)

	listener, err := net.Listen("tcp", ":5555")
	if err != nil {
		LOGGER.Fatalf("Could not listen, %v", err.Error())
	}

	srv := amsgRPC.NewGRPCServer(srvOpts)

	err = srv.Serve(listener)
	if err != nil {
		LOGGER.Fatalf("Could not serve, %v", err.Error())
	}

}
