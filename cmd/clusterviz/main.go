package main

import (
    "clusterviz/internal/pkg/server"
    logger "github.com/sirupsen/logrus"
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
        logger.Fatalf("Unable to create the server instance, error: %v", err)
    }
    logger.Infof("Server initialized successfully...")
    s.Start()
}
