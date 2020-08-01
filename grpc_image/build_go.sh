#!/bin/sh
cd /
ls /proto/*

protoc -I. -I./include --go_out=plugins=grpc:/generated ./proto/*.proto
protoc -I. -I./include --plugin=protoc-gen-grpc-gateway=${GOPATH}/bin/protoc-gen-grpc-gateway --grpc-gateway_out=logtostderr=true:./generated 	./proto/*.proto
protoc -I. -I./include --plugin=protoc-gen-swagger=${GOPATH}/bin/protoc-gen-swagger --swagger_out=logtostderr=true:./generated ./proto/*.proto

mv /generated/proto/* /generated
rm -rf /generated/proto
