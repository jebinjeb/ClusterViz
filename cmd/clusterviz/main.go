package main

import (
	logger "github.com/sirupsen/logrus"
	"clusterviz/internal/pkg/server"
	"os"
)

func main() {
	logger.SetFormatter(&logger.JSONFormatter{})
	logger.SetReportCaller(true)
	logger.SetLevel(logger.DebugLevel)
	logger.SetOutput(os.Stdout)
	logger.Infof("Server Starting...")
	s, err := server.New()
	if err != nil {
		logger.Fatalf("unable to create the server instance, gerror: %v", err)
	}
	logger.Infof("Server initialized successfully...")
	s.Start()
}
