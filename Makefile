generate:
	@if [ -d "./generated" ]; then echo "generated  exist"; else mkdir generated; fi

	@protoc -I. -I${GOPATH}/src/github.com/morimar32/helpers/proto/third_party --go_out=plugins=grpc:./generated proto/*.proto
	@protoc -I. -I${GOPATH}/src/github.com/morimar32/helpers/proto/third_party --plugin=protoc-gen-grpc-gateway=${GOPATH}/bin/protoc-gen-grpc-gateway --grpc-gateway_out=logtostderr=true:./generated 	proto/*.proto
	@protoc -I. -I${GOPATH}/src/github.com/morimar32/helpers/proto/third_party --plugin=protoc-gen-swagger=${GOPATH}/bin/protoc-gen-swagger --swagger_out=logtostderr=true:./generated proto/*.proto

	@mv ./generated/proto/* ./generated && 	rm -rf ./generated/proto
	@mv ./generated/*.json ./swagger

build:
	@go build

deps:
	if [ -d "./swagger" ]; then echo "swagger exists; skipping"; else echo "swagger does not exists"; npm install swagger-ui-dist; mv ./node_modules/swagger-ui-dist ./swagger; rm -rf node_modules; fi

test:
	@go vet
	@go test --race -cover -coverprofile=coverage.out personsvc/service
	@go tool cover -func=coverage.out
	@rm coverage.out

release: test
	@go mod download
	@go mod verify
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o svc
	@upx -9 -q svc

generate_docker_prep:
	@mkdir ${GOPATH}/src/github.com/morimar32 && ln -s ${GOPATH}/src/personsvc/vendor/github.com/morimar32/helpers ${GOPATH}/src/github.com/morimar32/helpers

test_docker:
	@go vet
	@go test -cover -coverprofile=coverage.out personsvc/service
	@go tool cover -func=coverage.out
	@rm coverage.out

release_docker: test_docker
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -mod vendor -o svc
	@upx -9 -q svc

docker:
	@go mod tidy
	@go mod vendor
	@cp -R ../helpers/proto/third_party ./vendor/github.com/morimar32/helpers/proto/
	@docker build . -t person -f ./dockerconfig/Dockerfile-personsvc
	@echo y | docker image prune
	@rm -rf vendor

docker_build_base:
	@docker build . -t svc_build_base -f ./dockerconfig/Dockerfile-buildbase

all: generate build

healthchk:
	@cd healthcheck; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -mod vendor -o ping
	@cd healthcheck; upx -9 -q ping; mv ping ../

clean:
	@rm -rf vendor
	@rm person
	@rm client/client