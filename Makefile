current_dir := $(shell pwd)

generate:
	@if [ -d "./generated" ]; then echo "generated  exist"; else mkdir generated; fi

	@docker run -v "$(current_dir)"/proto:/proto:ro -v "$(current_dir)"/generated:/generated:rw --name genr8 morimar/grpc_image
	@docker rm genr8
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
	@docker build . -t person
	@echo y | docker image prune
	@rm -rf vendor

all: generate build

healthchk:
	@cd healthcheck; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -mod vendor -o ping
	@cd healthcheck; upx -9 -q ping; mv ping ../

clean:
	@rm -rf vendor
	@rm person
	@rm client/client
