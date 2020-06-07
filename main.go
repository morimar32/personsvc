package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/morimar32/helpers/encryption"
)

const (
	httpAddress = "0.0.0.0:8080"
	grpcaddress = "0.0.0.0:9090"
)

var (
	connectionString = ""
)

func init() {
	env := os.Getenv("ENV")
	if len(env) <= 0 {
		env = "Dev"
	}
	envPath := fmt.Sprintf(".%s.env", env)
	godotenv.Load(envPath)
	connEnc := os.Getenv("connectionString")
	if len(connEnc) <= 0 {
		log.Fatalf("connectionString not defined in %s\n", envPath)
	}
	connPart, err := encryption.Decrypt(connEnc)
	if err != nil {
		log.Fatal(err)
	}
	if len(connPart) <= 0 {
		log.Fatalf("connectionString decrypted to empty in %s\n", envPath)
	}

	dbServer := os.Getenv("DB_HOST")
	if len(dbServer) <= 0 {
		dbServer = "localhost"
	}
	connectionString = fmt.Sprintf(connPart, dbServer)

}

func main() {
	go func() {
		log.Fatal(launchGRPC())
	}()
	log.Fatal(launchHTTP())
}
