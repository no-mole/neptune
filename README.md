# Neptune-简单快速的grpc & web框架

## Requirements
* OS: Linux and Mac OS, Windows is on the road!

## Installation

Using go proxy
```shell
go env -w GOPROXY="https://goproxy.cn,direct"
```

Local development
```shell
go env -w  GOPRIVATE="gitlab.com.cn"  ##Optional command，for faster pull dependency locally
```

With [Go module][] support (Go 1.11+), simply add the following import
```shell
import "github.com/no-mole/neptune"
```

To your code, and then `go [build|run|test]` will automatically fetch the
necessary dependencies.

Otherwise, to install the `neptune` package or using `neptune` cli, run the following command:

```console
$ go get github.com/no-mole/neptune
$ neptune help
```

How to generate proto files in different languages

choose your programming language install [protoc 3.20.1](https://github.com/protocolbuffers/protobuf/releases/tag/v3.20.1) 

install protoc-gen-go-grpc
```console
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
```

install protoc-gen-go 
```console
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
```

```console
neptune proto-gen -l golang -path /bar/bar.proto

-l Specifies the programming language, currently supported golang, java, python, php; Default golang
-path Specify the address of the proto file

```


## Features

Neptune abstracts away the details of distributed systems. Here are the main features.

- **Dynamic Config** - Load and hot reload dynamic config from anywhere. The config interface provides a way to load application
  level config from any source such as env vars, file, etcd. You can merge the sources and even define fallbacks.

- **Data Storage** - A simple data store interface to read, write and delete records.

- **Service Discovery** - Automatic service registration and name resolution. Service discovery is at the core of micro service
    development. When service A needs to speak to service B it needs the location of that service. The default discovery mechanism is
    ETCD.

- **Load Balancing** - Client side load balancing built on service discovery. Once we have the addresses of any number of instances
  of a service we now need a way to decide which node to route to. We use random hashed load balancing to provide even distribution
  across the services and retry a different node if there's a problem. 

- **RPC Client/Server** - RPC based request/response with support for bidirectional streaming. We provide an abstraction for synchronous
  communication. A request made to a service will be automatically resolved, load balanced, dialled and streamed.

- **Tracing** - Link tracing is an integral part of a distributed system that is optimized and analyzed by recording call levels and 
  time-consuming.

- **Logging** - logs can record important events as monitoring. neptune supports sending logs to files, stdout, or sending logs to the
  log center through grpc/udp protocol.

- **Caching** - Use gin middleware to support automatic cache control of GET requests.

- **Databases&Queues** - Gorm(postgreSQL/MySQL/Clickhouse)、ES、Mongo、Redis、RabbitMQ、Kafka.