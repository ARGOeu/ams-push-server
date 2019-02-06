package main

import (
	"bytes"
	"flag"
	"fmt"
	amsgRPC "github.com/ARGOeu/ams-push-server/api/v1/grpc"
	"github.com/ARGOeu/ams-push-server/config"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"io/ioutil"
	"log/syslog"
	"net"
)

func init() {

	fmter := &log.TextFormatter{FullTimestamp: true, DisableColors: true}
	hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")

	log.SetFormatter(fmter)

	if err == nil {
		log.AddHook(hook)
	}
}

func main() {

	// Retrieve configuration file location from a cli argument
	cfgPath := flag.String("config", "/etc/ams-push-server/conf.d/ams-push-server-config.json", "Path for the required configuration file.")
	flag.Parse()

	bCfg, err := ioutil.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("Error while reading file, %v", err.Error())
	}

	// initialize the config
	cfg := new(config.Config)
	err = cfg.LoadFromJson(bytes.NewReader(bCfg))
	if err != nil {
		log.Fatalf("Error while loading configuration, %v", err.Error())
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.ServicePort))
	if err != nil {
		log.Fatalf("Could not listen, %v", err.Error())
	}

	srv := amsgRPC.NewGRPCServer(cfg)

	log.Infof("API    gRPC Server will serve on port: %v", cfg.ServicePort)

	defer func() {
		listener.Close()
		srv.GracefulStop()
	}()

	err = srv.Serve(listener)
	if err != nil {
		log.Fatalf("Could not serve, %v", err.Error())
	}
}
