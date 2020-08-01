FROM golang:alpine

RUN apk update && apk add --no-cache make git gcc libc-dev grpc-dev protobuf-dev 

RUN go get github.com/golang/protobuf/protoc-gen-go
RUN go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN go get -d github.com/envoyproxy/protoc-gen-validate

RUN go install github.com/golang/protobuf/protoc-gen-go
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN cd $GOPATH/src/github.com/envoyproxy/protoc-gen-validate && make build

VOLUME ["/include"]
VOLUME ["/proto"]
VOLUME ["/generated"]

COPY include /include
COPY build_go.sh /
ENTRYPOINT ["/build_go.sh"]