package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	person "personsvc/generated"
)

func init() {
	fmt.Println("HTTP init")
}

func launchHTTP() error {
	conn, err := grpc.Dial(grpcaddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println("could not connect to gRPC service")
		return err
	}

	defer conn.Close()

	jsonpb := &runtime.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)
	err = person.RegisterPersonHandler(context.Background(), gwmux, conn)
	if err != nil {
		fmt.Println("could not register handler")
		return err
	}
	swagger := http.FileServer(http.Dir("./swagger"))
	gwServer := &http.Server{
		Addr: httpAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				gwmux.ServeHTTP(w, r)
				return
			}
			swagger.ServeHTTP(w, r)
		}),
	}

	return gwServer.ListenAndServe()
}
