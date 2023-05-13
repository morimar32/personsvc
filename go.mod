module personsvc

go 1.16

require (
	github.com/99designs/gqlgen v0.17.30
	github.com/ZachtimusPrime/Go-Splunk-HTTP v0.0.0-20200420213219-094ff9e8d788
	github.com/cch123/gogctuner v0.0.0-20220625123136-31caf07e2832
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc
	github.com/golang/protobuf v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/json-iterator/go v1.1.12
	github.com/morimar32/helpers v0.1.0
	github.com/vektah/gqlparser/v2 v2.5.1
	go.uber.org/zap v1.15.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sys v0.5.0
	google.golang.org/genproto v0.0.0-20200605102947-12044bf5ea91
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v2 v2.3.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
)

//replace github.com/morimar32/helpers => ../helpers
