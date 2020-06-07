module personsvc

go 1.14

require (
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.4
	github.com/joho/godotenv v1.3.0
	github.com/morimar32/helpers v0.0.0-20200601013417-469597584a8b
	go.uber.org/zap v1.15.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200331124033-c3d80250170d // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20200331025713-a30bf2db82d4 // indirect
	google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.23.0
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
)

replace github.com/morimar32/helpers => ../helpers
