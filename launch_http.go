package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	person "personsvc/generated"

	"golang.org/x/sys/unix"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"
)

type jsonfast struct{}

func (j *jsonfast) Marshal(v interface{}) ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(v)
}

func (j *jsonfast) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.ConfigFastest.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (j *jsonfast) NewDecoder(r io.Reader) runtime.Decoder {
	return jsoniter.ConfigFastest.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (j *jsonfast) NewEncoder(w io.Writer) runtime.Encoder {
	return jsoniter.ConfigFastest.NewEncoder(w)
}

// ContentType returns the Content-Type which this marshaler is responsible for.
func (j *jsonfast) ContentType() string {
	return "application/json"
}

func launchHTTP() error {
	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure(), grpc.WithBlock())
	//conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(UnixDialer))
	if err != nil {
		fmt.Println("could not connect to gRPC service")
		return err
	}

	defer conn.Close()
	/*
		jsonpb := &runtime.JSONPb{
			EmitDefaults: true,
			Indent:       "  ",
			OrigName:     true,
		}
	*/
	j := &jsonfast{}
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, j),
	)
	err = person.RegisterPersonHandler(context.Background(), gwmux, conn)
	if err != nil {
		fmt.Println("could not register handler")
		return err
	}
	swagger := http.FileServer(http.Dir("./web/static/swagger"))
	gwServer := &http.Server{
		//Addr: httpAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				gwmux.ServeHTTP(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/proto") {
				Log.Info("Calling into /proto path")
				http.ServeFile(w, r, "./api/person.proto")
				return
			}

			swagger.ServeHTTP(w, r)
		}),
	}
	lc := net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			if err := conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			}); err != nil {
				return err
			}
			return operr
		},
	}
	ln, err := lc.Listen(context.Background(), "tcp", httpAddress)
	if err != nil {
		return err
	}
	return gwServer.Serve(ln)
	//return gwServer.ListenAndServe()
}

func UnixDialer(addr string, t time.Duration) (net.Conn, error) {
	unix_addr, err := net.ResolveUnixAddr("unix", grpcUnixSocket)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, unix_addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
