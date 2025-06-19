package main

import (
	"bytes"
	"flag"
	"fmt"
	amsgRPC "github.com/ARGOeu/ams-push-server/api/v1/grpc"
	"github.com/ARGOeu/ams-push-server/config"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
)

func init() {
	fmter := &log.TextFormatter{FullTimestamp: true, DisableColors: true}
	log.SetFormatter(fmter)
}

func main() {

	// Retrieve configuration file location from a cli argument
	cfgPath := flag.String("config", "/etc/ams-push-server/conf.d/ams-push-server-config.json", "Path for the required configuration file.")
	flag.Parse()

	bCfg, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":  "error_log",
				"path":  *cfgPath,
				"error": err.Error(),
			},
		).Fatal("Could not read configuration file")
	}

	// initialize the config
	cfg := new(config.Config)
	err = cfg.LoadFromJson(bytes.NewReader(bCfg))
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":  "error_log",
				"error": err.Error(),
			},
		).Fatal("Could not load configuration file")
	}

	log.SetLevel(cfg.GetLogLevel())

	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", cfg.BindIp, cfg.BindPort))
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":  "error_log",
				"error": err.Error(),
			},
		).Fatal("Could not listen")
	}

	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("API is ready to start serving")

	srv := amsgRPC.NewGRPCServer(cfg)

	defer func() {
		listener.Close()
		srv.GracefulStop()
	}()

	err = srv.Serve(listener)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":  "error_log",
				"error": err.Error(),
			},
		).Fatal("Could not serve")
	}
}
