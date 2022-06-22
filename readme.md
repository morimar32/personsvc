## Overview ##

This project is meant to be an example/reference of what a simple, straightforward, yet production-ready go microservice. How close it comes to achieving its goals is up for debate. 

#### 30,000 ft. view ####
The service is meant to be a gRPC service, but also supports REST endpoints along with supporting Swagger. All 3 are generated from the same `.proto` file via https://github.com/grpc-ecosystem/grpc-gateway. The generated REST routes through into the gRPC instances [as opposed to redundant implementations]. There is an attempt to structure the solution with hexagonal architecture [with an exception of that linear service-to-service dependency], and tries to follow naming conventions accordingly.

Interfaces are defined between the different layers, and the implementing types also create a factory function that follow inversion of control. There are two reasons for this: first, it does make for easier testability. Second, this seems like it would be a bit more familiar/comfortable for C# developers who may be reviewing and/or possibly implementing. Go doesn't seem to have good options for dependency injection, but the factory methods are straight forward and have the added benefit of causing a compiler error if a function gets added/changed to the interface that hasn't been updated in the implementing class [essentially enforcing an _explicit_ interface behavior].

The majority of the interactions have been abstracted out into the `Makefile` and can be invoked via [incomplete list]:
* `make generate` - generate code from the .proto file
* `make test` - run tests and code coverage
* `make docker` - kick off the docker image creation

The Docker image is built within docker itself, as part of multistage build. Currently this process requires go on the host _only_ because it uses `go mod vendor` to bring in private repos to be copied in. If the base build image was built with the appropriate access tokens, the building of the image could be done on a host without go at all. 

The building process will compile, then run tests, and will also fail if a certain threshold [defined within code] is not met. The static compiled binary is then compressed [via `upx`] and copied into a scratch-built image, along with a 2cnd binary [following the same process] that is a healthcheck. The service runs under a non-root user created during the multistage build. This user has no login capabilities, and no explicit password. The service also is meant to be run with all capabilities dropped. [`cap_drop: ALL`]. As of right now, the image size builds to 19.4mb, and runs in < 30mb RAM.


#### 10,000 ft. view ####
* protobuf/gRPC driven solution
* 3rd party code gen support for REST wrappers around 
* 3rd party code gen support for swagger on top of previously mentioned REST Api
* MSSQL - trying to solve specific/problematic areas
  * `GUID` identifiers
  * `DATETIME` and nullable `DATETIME` columns
* Splunk support via zap
* tracing for app-dynamics via opentracing [and testing via jaeger internally]
* emphasis on security/minimal exposure 
* targeting naive/simplistic approach for hexagonal architecture
* goal of _not_ committing generated code/artifacts
* self-contained build process
* code-coverage as a gating mechanism for quality

#### Overall Status ####
* Done
  * initial implementation 🎉
* Performance To-do/experiments (with benchmarks/pprof dumps)
  * change json serializer (and perf tuning) - https://github.com/json-iterator/go
  * change to gogoproto for optimized (de)serialization - https://github.com/gogo/protobuf
  * change http->gRPC to use unix socket instead of TCP socket (listen on both)
  * experiment with setting a balast (also verify with docker mem limits) - https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap/
  * GC tuning library? - https://github.com/cch123/gogctuner
  * set HTTP socket options for HTTP & gRPC - https://iximiuz.com/en/posts/go-net-http-setsockopt-example/
  * add sync.Pool - https://www.cockroachlabs.com/blog/how-to-optimize-garbage-collection-in-go/
* To-do
  * listen for closing signal
  * add volume mount to docker compose for encryption keys
  * add tracing [app dynamics/jaeger support] - https://medium.com/swlh/distributed-tracing-for-go-microservice-with-opentracing-1fc1aec76b3e
  * make sure correlationid is supported [only add if _not_ passed in]
  * make sure context is added/used properly - https://blog.golang.org/context
  * golang resilience library - https://medium.com/@slok/goresilience-a-go-library-to-improve-applications-resiliency-14d229aee385
  * rate limiting & 
  * extend out where logging is - https://www.oreilly.com/content/how-to-ship-production-grade-go/
  * MUCH MORE TESTS
  * Add fuzzing tests
* Possible upcoming changes
  * switch to generated validation - https://github.com/envoyproxy/protoc-gen-validate
  * prometheus support [grpc & general perf stats] - https://prometheus.io/docs/guides/go-application/
  * k8s scripts
  * add eventing [rabbitmq? zeromq?]
  * slight restructuring of project? conceptually extend to supporting more than just a person object
  * possible graphql support, also routed thru gRPC endpoints - https://github.com/99designs/gqlgen
  * experiment with sonarqube - https://docs.sonarqube.org/latest/analysis/languages/go/
  * explore build tags for combining various interchangable components, such as database

#### Project structure & notes ####

