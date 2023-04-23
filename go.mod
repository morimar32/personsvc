module personsvc

go 1.14

require (
	github.com/ZachtimusPrime/Go-Splunk-HTTP v0.0.0-20200420213219-094ff9e8d788
	github.com/cch123/gogctuner v0.0.0-20220625123136-31caf07e2832
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/json-iterator/go v1.1.12
	github.com/morimar32/helpers v0.1.0
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20200331025713-a30bf2db82d4 // indirect
	google.golang.org/genproto v0.0.0-20200605102947-12044bf5ea91
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
)

//replace github.com/morimar32/helpers => ../helpers
