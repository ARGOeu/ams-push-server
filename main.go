package main

import (
	amsgRPC "github.com/ARGOeu/ams-push-server/api/v1/grpc"
	LOGGER "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"log/syslog"
	"net"
)

func init() {
	LOGGER.SetFormatter(&LOGGER.TextFormatter{FullTimestamp: true, DisableColors: true})
	hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		LOGGER.AddHook(hook)
	}
}

func main() {

	listener, err := net.Listen("tcp", ":5555")
	if err != nil {
		LOGGER.Fatalf("Could not listen, %v", err.Error())
	}

	srv := amsgRPC.NewGRPCServer()

	err = srv.Serve(listener)
	if err != nil {
		LOGGER.Fatalf("Could not serve, %v", err.Error())
	}

}
