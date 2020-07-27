package main

import (
	"log"

	"go.uber.org/zap"
)

const (
	httpAddress = "0.0.0.0:8080"
	grpcaddress = "0.0.0.0:9090"
)

var (
	connectionString = ""
	//Log centralized entry point for logging
	Log *zap.Logger
)

func main() {
	go func() {
		log.Fatal(launchpprof())
	}()
	go func() {
		log.Fatal(launchGRPC())
	}()
	log.Fatal(launchHTTP())
}
