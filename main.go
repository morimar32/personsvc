package main

import (
	"fmt"
	"log"

	env "github.com/morimar32/helpers/environment"
)

const (
	httpAddress = "0.0.0.0:8080"
	grpcaddress = "0.0.0.0:9090"
)

var (
	connectionString = ""
)

func init() {
	if err := env.LoadEnvironmentFile(); err != nil {
		log.Fatal(err)
	}
	connPart, err := env.GetEncryptedValue("connectionString")
	if err != nil {
		log.Fatal(err)
	}

	dbServer := env.GetValueWithDefault("DB_HOST", "localhost")
	connectionString = fmt.Sprintf(connPart, dbServer)
}

func main() {
	go func() {
		log.Fatal(launchGRPC())
	}()
	log.Fatal(launchHTTP())
}
