package main

import (
	"context"
	"log"
	"time"

	person "personsvc/generated"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

const (
	address = "0.0.0.0:9090"
)

func main() {
	success := make(chan bool)
	err := make(chan string)
	timeout := make(chan bool)

	go healthcheck(success, err)
	go func(t chan<- bool) {
		time.Sleep(5 * time.Second)
		t <- true
	}(timeout)

	select {
	case _ = <-success:
		log.Println("Success")
	case errMsg := <-err:
		log.Fatal(errMsg)
	case _ = <-timeout:
		log.Fatal("Timed out waiting to connect")
	}

}

func healthcheck(success chan<- bool, e chan<- string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		e <- err.Error()
		return
	}
	defer conn.Close()
	c := person.NewPersonClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IlF4UVF1OHoxTDhqTHZvTDliSWFNaWciLCJ0eXAiOiJhdCtqd3QifQ.eyJuYmYiOjE1OTA4NzQxNzYsImV4cCI6MTU5MDg3Nzc3NiwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo1MDAwIiwiYXVkIjoicGVyc29uIiwiY2xpZW50X2lkIjoicGVyc29uX2NsaWVudCIsInNjb3BlIjpbInBlcnNvbiJdfQ.CG2F-ReyykNrHePUc9K1HzSxXDHubIdzbkuXrgS9-MFEZA63Q3kR7-Ewd0crOGUb2mW4Ixt2ZRutGpUJ7cY6_DkWhtJwiq8UPgDONZA_uZBzr5bGPcERQQ8NIrrxemxNA-eNZO-GpNX6EW4xMGrnhBrR3OqrYiVi-gjFeQB2cixmiZkLZr77-IDEoCZZpVFwEOyD-YD8su1RQbrwW_QiAWyyuVmFXJDzjglScqE3inrrg50B9k0F3IhN2a7mPRQnP1Ns7IT4z0kmjghA10oRhtmJxMmasEGQe3BRaXyJu4SAjX-Xd1j0ega4UPwbQQwu6T2z8O-GKvS_yodyHJv8Sg")

	_, err = c.Ping(ctx, &empty.Empty{})
	if err != nil {
		e <- err.Error()
		return
	}
	success <- true
}
